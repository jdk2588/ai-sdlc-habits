package main

import (
	"fmt"
	"sync"
	"time"
)

type CreateOrderRequest struct {
	Item     string  `json:"item"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type Order struct {
	ID          string    `json:"order_id"`
	Item        string    `json:"item"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	OrderStatus string    `json:"order_status"`
	CreatedAt   time.Time `json:"created_at"`
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
