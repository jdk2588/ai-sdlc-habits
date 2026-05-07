package main

import (
	"errors"
	"fmt"
	"sync"
)

const (
	StatusPlaced     = "placed"
	StatusProcessing = "processing"
	StatusFulfilled  = "fulfilled"
	StatusFailed     = "failed"
)

var errInvalidTransition = errors.New("invalid state transition")

type FulfillmentRecord struct {
	OrderID  string `json:"order_id"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
	Status   string `json:"-"`
}

var (
	mu      sync.RWMutex
	records = make(map[string]*FulfillmentRecord)
)

func upsert(r *FulfillmentRecord) {
	mu.Lock()
	defer mu.Unlock()
	records[r.OrderID] = r
}

// transitionTo advances r to next, returning an error on invalid transition.
// r.Status is not modified on error.
func transitionTo(r *FulfillmentRecord, next string) error {
	switch {
	case r.Status == StatusPlaced && next == StatusProcessing,
		r.Status == StatusProcessing && next == StatusFulfilled,
		r.Status == StatusProcessing && next == StatusFailed:
		r.Status = next
		return nil
	default:
		return fmt.Errorf("order %s: %w: %s -> %s", r.OrderID, errInvalidTransition, r.Status, next)
	}
}

// fulfill runs the state machine: placed -> processing -> fulfilled.
// On transition failure after reaching processing, status is set to failed.
func fulfill(r *FulfillmentRecord) error {
	if err := transitionTo(r, StatusProcessing); err != nil {
		return err
	}
	upsert(r)
	if err := transitionTo(r, StatusFulfilled); err != nil {
		transitionTo(r, StatusFailed)
		upsert(r)
		return err
	}
	upsert(r)
	return nil
}
