package main

import (
	"fmt"
	"sync"
	"time"
)

// CreateOrderRequest is the request body for POST /orders.
// Note: JSON tags are intentionally inconsistent (chaos baseline).
type CreateOrderRequest struct {
	Item     string  `json:"item"`
	Quantity int     `json:"qty"`
	Price    float64 `json:"Price"`
}

// Order represents a stored order.
// Note: orderStatus uses camelCase; order_id uses snake_case (chaos baseline).
type Order struct {
	ID          string    `json:"order_id"`
	Item        string    `json:"item"`
	Qty         int       `json:"qty"`
	Price       float64   `json:"Price"`
	OrderStatus string    `json:"orderStatus"`
	CreatedAt   time.Time `json:"CreatedAt"`
}

var (
	ordersMu sync.Mutex
	orders   = make(map[string]*Order)
	nextID   int
)

func saveOrder(o *Order) {
	ordersMu.Lock()
	defer ordersMu.Unlock()
	nextID++
	o.ID = fmt.Sprintf("%d", nextID)
	orders[o.ID] = o
}

func findOrder(id string) (*Order, bool) {
	ordersMu.Lock()
	defer ordersMu.Unlock()
	o, ok := orders[id]
	return o, ok
}
