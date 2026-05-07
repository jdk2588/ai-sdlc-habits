# Design — Order Processing: Fulfillment State Machine and Order Query

This document was written before implementation began. Implementation was gated on
review and approval of this design. See the Execution Gate section.

## Scope

### In scope

- `GET /orders/{id}` — read a single order from the orders service by ID
- Fulfillment state machine — status transitions from `placed` through `processing`
  to `fulfilled` or `failed`
- Wire schemas for `order.placed` and `order.fulfilled` messages

### Out of scope

- Retries or dead-letter queues for failed orders
- Order status updates flowing back from fulfillment to orders
- Persistent storage — both services use in-memory maps
- Authentication or authorization on any endpoint

## Components

### orders service (`services/orders/`)

| Component | Role |
|-----------|------|
| `handleGetOrder` | Reads an order by ID from the in-memory store and returns it as JSON |
| `findOrder(id)` | Store lookup by string ID |
| `Order` struct | Data shape returned by `GET /orders/{id}` |

### fulfillment service (`services/fulfillment/`)

| Component | Role |
|-----------|------|
| `FulfillmentRecord` | Per-order state tracked by the fulfillment service |
| `transitionTo(r, next)` | Validates and applies a single status transition |
| `fulfill(r)` | Orchestrates the full state machine run for one order |
| `handleDelivery(d, pub)` | Consumes one `order.placed` delivery and drives `fulfill` |
| `amqpPublisher.publish(orderID)` | Emits `order.fulfilled` on success |

## Data flow

### Happy path

```
Client          orders service         RabbitMQ         fulfillment service
  |                   |                    |                     |
  | POST /orders      |                    |                     |
  |------------------>|                    |                     |
  |                   | saveOrder()        |                     |
  |                   | publish            |                     |
  |                   |------------------->|                     |
  | 201 Created       |                    | deliver order.placed |
  |<------------------|                    |-------------------->|
  |                   |                    |  transitionTo(processing)
  |                   |                    |  transitionTo(fulfilled)
  |                   |                    |  publish order.fulfilled
  |                   |                    |<--------------------|
  |                   |                    |                     |
  | GET /orders/{id}  |                    |                     |
  |------------------>|                    |                     |
  | 200 OK (order)    |                    |                     |
  |<------------------|                    |                     |
```

### Error path: order not found

```
Client          orders service
  |                   |
  | GET /orders/99    |
  |------------------>|
  |                   | findOrder("99") -> not found
  | 404 Not Found     |
  |<------------------|
```

### Error path: RabbitMQ unavailable at startup

Both services call `setupAMQP` at startup. If RabbitMQ is unreachable, `amqp.Dial`
returns an error. The service wraps it with `fmt.Errorf` and calls `log.Fatalf`,
which exits the process. No orders are created or processed until RabbitMQ is
available and the service is restarted.

Neither service degrades gracefully in this version — an unavailable RabbitMQ is a
hard startup failure.

### Error path: malformed `order.placed` message

If `handleDelivery` cannot unmarshal the message body, or if the `order_id` field
is empty, it logs the error and calls `d.Nack(false, false)`. The message is
dropped (no DLQ is configured in this phase). The order is not processed.

### Error path: invalid state transition during fulfillment

`transitionTo` validates each transition and does not mutate the record on error.
If the `processing -> fulfilled` transition fails, `fulfill` attempts to transition
to `failed` and upserts that state before returning the error. The delivery is
Nacked.

## Interfaces

### HTTP: GET /orders/{id}

**Request**

```
GET /orders/{id} HTTP/1.1
```

No request body. `{id}` is the string order ID returned by `POST /orders`.

**Response — 200 OK**

```json
{
  "order_id": "1",
  "item": "widget",
  "quantity": 2,
  "price": 9.99,
  "order_status": "placed",
  "created_at": "2024-01-15T10:30:00Z"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | string | Auto-incremented integer, serialized as a string |
| `item` | string | Item name, non-empty |
| `quantity` | integer | Greater than zero |
| `price` | float | Greater than zero |
| `order_status` | string | One of: `placed`, `processing`, `fulfilled`, `failed` |
| `created_at` | string (RFC3339) | Timestamp set at order creation |

**Response — 404 Not Found**

```json
{"error": "order not found"}
```

### AMQP: order.placed

Published by `orders` after a successful `POST /orders`. Consumed by `fulfillment`.

**Queue:** `order.placed`
**Content-Type:** `application/json`

```json
{
  "order_id": "1",
  "item": "widget",
  "quantity": 2
}
```

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | string | Matches the `order_id` returned by `POST /orders` |
| `item` | string | Copied from the order creation request |
| `quantity` | integer | Copied from the order creation request |

### AMQP: order.fulfilled

Published by `fulfillment` after a successful state machine run.

**Queue:** `order.fulfilled`
**Content-Type:** `application/json`

```json
{
  "order_id": "1",
  "status": "fulfilled"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | string | Matches the `order_id` from the `order.placed` message |
| `status` | string | Always `"fulfilled"` for the success path |

## Fulfillment state machine

```
          order.placed received
                  |
                  v
              [ placed ]
                  |
                  | transitionTo(processing)
                  v
            [ processing ]
            /             \
           /               \
          v                 v
     [ fulfilled ]       [ failed ]
```

Valid transitions:

| From | To | Trigger |
|------|----|---------|
| `placed` | `processing` | Start of processing |
| `processing` | `fulfilled` | Processing succeeded |
| `processing` | `failed` | Processing error |

Any other transition returns `errInvalidTransition`. The `transitionTo` function
does not mutate the record on error, so the original status is preserved.

## Execution gate

This document was reviewed and approved before implementation began. The
`GET /orders/{id}` handler and the fulfillment state machine were not written until
the data flow, interface contracts, and error paths described above had been agreed
upon. Any implementation divergence from this document was treated as a change
requiring re-review.
