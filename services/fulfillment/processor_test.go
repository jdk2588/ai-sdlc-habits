package main

import (
	"errors"
	"testing"
)

func makeRecord(id string, status OrderStatus) *FulfillmentRecord {
	return &FulfillmentRecord{OrderID: id, Status: status}
}

func TestTransition_PlacedToProcessing(t *testing.T) {
	r := makeRecord("1", Placed)
	if err := transitionTo(r, Processing); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != Processing {
		t.Errorf("want Processing, got %v", r.Status)
	}
}

func TestTransition_ProcessingToFulfilled(t *testing.T) {
	r := makeRecord("1", Processing)
	if err := transitionTo(r, Fulfilled); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != Fulfilled {
		t.Errorf("want Fulfilled, got %v", r.Status)
	}
}

func TestTransition_ProcessingToFailed(t *testing.T) {
	r := makeRecord("1", Processing)
	if err := transitionTo(r, Failed); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != Failed {
		t.Errorf("want Failed, got %v", r.Status)
	}
}

func TestTransition_PlacedToFulfilled_Invalid(t *testing.T) {
	r := makeRecord("1", Placed)
	err := transitionTo(r, Fulfilled)
	if err == nil {
		t.Fatal("expected error for Placed->Fulfilled, got nil")
	}
	var pe *ProcessingError
	if !errors.As(err, &pe) {
		t.Errorf("want ProcessingError, got %T", err)
	}
	if r.Status != Placed {
		t.Errorf("status must not change on invalid transition, got %v", r.Status)
	}
}

func TestTransition_PlacedToFailed_Invalid(t *testing.T) {
	r := makeRecord("1", Placed)
	if err := transitionTo(r, Failed); err == nil {
		t.Fatal("expected error for Placed->Failed, got nil")
	}
	if r.Status != Placed {
		t.Errorf("status must not change on invalid transition, got %v", r.Status)
	}
}

func TestTransition_FulfilledIsTerminal(t *testing.T) {
	r := makeRecord("1", Fulfilled)
	if err := transitionTo(r, Processing); err == nil {
		t.Error("expected error for Fulfilled->Processing")
	}
	if err := transitionTo(r, Failed); err == nil {
		t.Error("expected error for Fulfilled->Failed")
	}
}

func TestFulfill_HappyPath(t *testing.T) {
	r := makeRecord("happy-1", Placed)
	if err := fulfill(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != Fulfilled {
		t.Errorf("want Fulfilled, got %v", r.Status)
	}
}

func TestFulfill_InvalidStartState(t *testing.T) {
	r := makeRecord("bad-1", Fulfilled)
	if err := fulfill(r); err == nil {
		t.Fatal("expected error when order is already Fulfilled")
	}
}
