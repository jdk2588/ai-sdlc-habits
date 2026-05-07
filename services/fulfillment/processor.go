package main

import (
	"errors"
	"fmt"
	"sync"
)

// OrderStatus is a typed status enum.
// Chaos baseline: Developer 2 used typed int constants; Developer 1 used plain strings.
type OrderStatus int

const (
	Placed OrderStatus = iota
	Processing
	Fulfilled
	Failed
)

func (s OrderStatus) String() string {
	switch s {
	case Placed:
		return "placed"
	case Processing:
		return "processing"
	case Fulfilled:
		return "fulfilled"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}

// ProcessingError is a typed error for state machine failures.
// Chaos baseline: Developer 2 uses custom error types; Developer 1 uses errors.New / fmt.Errorf.
type ProcessingError struct {
	OrderID string
	Err     error
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("order %s: %v", e.OrderID, e.Err)
}

func (e *ProcessingError) Unwrap() error { return e.Err }

var errInvalidTransition = errors.New("invalid state transition")

// FulfillmentRecord is the fulfillment service's view of an order.
// Chaos baseline: JSON tags use PascalCase; Developer 1's Order struct uses mixed snake/camel/Pascal.
// "Quantity" vs Developer 1's "qty" abbreviation.
type FulfillmentRecord struct {
	OrderID  string      `json:"OrderID"`
	Item     string      `json:"Item"`
	Quantity int         `json:"Quantity"`
	Status   OrderStatus `json:"-"`
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

// transitionTo advances r to next, returning ProcessingError on invalid transition.
// r.Status is not modified on error.
func transitionTo(r *FulfillmentRecord, next OrderStatus) error {
	switch {
	case r.Status == Placed && next == Processing,
		r.Status == Processing && next == Fulfilled,
		r.Status == Processing && next == Failed:
		r.Status = next
		return nil
	default:
		return &ProcessingError{
			OrderID: r.OrderID,
			Err:     fmt.Errorf("%w: %s -> %s", errInvalidTransition, r.Status, next),
		}
	}
}

// fulfill runs the state machine: placed -> processing -> fulfilled.
// On transition failure after reaching processing, status is set to failed.
func fulfill(r *FulfillmentRecord) error {
	if err := transitionTo(r, Processing); err != nil {
		return err
	}
	upsert(r)
	if err := transitionTo(r, Fulfilled); err != nil {
		transitionTo(r, Failed)
		upsert(r)
		return err
	}
	upsert(r)
	return nil
}
