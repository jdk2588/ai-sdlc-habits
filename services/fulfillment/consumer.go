package main

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type publisher interface {
	publish(orderID string) error
}

type amqpPublisher struct {
	ch *amqp.Channel
}

// FulfilledMessage is published to the order.fulfilled queue on success.
// Chaos baseline: Developer 2 chose snake_case JSON keys; Developer 1 used camelCase.
type FulfilledMessage struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

func (p *amqpPublisher) publish(orderID string) error {
	msg := FulfilledMessage{OrderID: orderID, Status: "fulfilled"}
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal fulfilled message: %w", err)
	}
	return p.ch.Publish(
		"fulfillment",
		"order.fulfilled",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// PlacedMessage matches the wire format published by the orders service.
// Developer 1 chose camelCase keys (orderId, qty); Developer 2 had to inspect
// the queue to discover this — it was not documented anywhere.
type PlacedMessage struct {
	OrderID string `json:"orderId"`
	Item    string `json:"item"`
	Qty     int    `json:"qty"`
}

func handleDelivery(d amqp.Delivery, pub publisher) {
	var msg PlacedMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		log.Printf("fulfillment: cannot parse message: %v", err)
		d.Nack(false, false)
		return
	}
	if msg.OrderID == "" {
		log.Printf("fulfillment: message missing orderId")
		d.Nack(false, false)
		return
	}

	r := &FulfillmentRecord{
		OrderID:  msg.OrderID,
		Item:     msg.Item,
		Quantity: msg.Qty,
		Status:   Placed,
	}

	if err := fulfill(r); err != nil {
		log.Printf("fulfillment: processing failed for order %s: %v", msg.OrderID, err)
		d.Nack(false, false)
		return
	}

	if err := pub.publish(r.OrderID); err != nil {
		log.Printf("fulfillment: failed to publish order.fulfilled for %s: %v", r.OrderID, err)
	}

	log.Printf("fulfillment: order %s fulfilled", r.OrderID)
	d.Ack(false)
}

func setupAMQP(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("connect: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("channel: %w", err)
	}
	for _, q := range []string{"order.placed", "order.fulfilled"} {
		if _, err = ch.QueueDeclare(q, true, false, false, false, nil); err != nil {
			ch.Close()
			conn.Close()
			return nil, nil, fmt.Errorf("declare %s: %w", q, err)
		}
	}
	return conn, ch, nil
}

func startConsuming(ch *amqp.Channel, pub publisher) error {
	msgs, err := ch.Consume("order.placed", "fulfillment", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}
	log.Println("fulfillment: consuming from order.placed")
	for d := range msgs {
		handleDelivery(d, pub)
	}
	return nil
}
