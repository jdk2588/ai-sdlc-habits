package main

import (
	"errors"
	"testing"
)

func makeRecord(id string, status string) *FulfillmentRecord {
	return &FulfillmentRecord{OrderID: id, Status: status}
}

func TestTransition_PlacedToProcessing(t *testing.T) {
	r := makeRecord("1", StatusPlaced)
	if err := transitionTo(r, StatusProcessing); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != StatusProcessing {
		t.Errorf("want %q, got %q", StatusProcessing, r.Status)
	}
}

func TestTransition_ProcessingToFulfilled(t *testing.T) {
	r := makeRecord("1", StatusProcessing)
	if err := transitionTo(r, StatusFulfilled); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != StatusFulfilled {
		t.Errorf("want %q, got %q", StatusFulfilled, r.Status)
	}
}

func TestTransition_ProcessingToFailed(t *testing.T) {
	r := makeRecord("1", StatusProcessing)
	if err := transitionTo(r, StatusFailed); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != StatusFailed {
		t.Errorf("want %q, got %q", StatusFailed, r.Status)
	}
}

func TestTransition_PlacedToFulfilled_Invalid(t *testing.T) {
	r := makeRecord("1", StatusPlaced)
	err := transitionTo(r, StatusFulfilled)
	if err == nil {
		t.Fatal("expected error for placed->fulfilled, got nil")
	}
	if !errors.Is(err, errInvalidTransition) {
		t.Errorf("want errInvalidTransition in chain, got %v", err)
	}
	if r.Status != StatusPlaced {
		t.Errorf("status must not change on invalid transition, got %q", r.Status)
	}
}

func TestTransition_PlacedToFailed_Invalid(t *testing.T) {
	r := makeRecord("1", StatusPlaced)
	if err := transitionTo(r, StatusFailed); err == nil {
		t.Fatal("expected error for placed->failed, got nil")
	}
	if r.Status != StatusPlaced {
		t.Errorf("status must not change on invalid transition, got %q", r.Status)
	}
}

func TestTransition_FulfilledIsTerminal(t *testing.T) {
	r := makeRecord("1", StatusFulfilled)
	if err := transitionTo(r, StatusProcessing); err == nil {
		t.Error("expected error for fulfilled->processing")
	}
	if err := transitionTo(r, StatusFailed); err == nil {
		t.Error("expected error for fulfilled->failed")
	}
}

func TestFulfill_HappyPath(t *testing.T) {
	r := makeRecord("happy-1", StatusPlaced)
	if err := fulfill(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != StatusFulfilled {
		t.Errorf("want %q, got %q", StatusFulfilled, r.Status)
	}
}

func TestFulfill_InvalidStartState(t *testing.T) {
	r := makeRecord("bad-1", StatusFulfilled)
	if err := fulfill(r); err == nil {
		t.Fatal("expected error when order is already fulfilled")
	}
}
