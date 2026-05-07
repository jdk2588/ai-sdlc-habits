package main

import "testing"

func TestValidateCreateOrderRequest_EmptyItem(t *testing.T) {
	req := CreateOrderRequest{Item: "", Quantity: 1, Price: 9.99}
	if err := validateCreateOrder(req); err == nil {
		t.Fatal("expected error for empty item, got nil")
	}
}

func TestValidateCreateOrderRequest_ZeroQuantity(t *testing.T) {
	req := CreateOrderRequest{Item: "widget", Quantity: 0, Price: 9.99}
	if err := validateCreateOrder(req); err == nil {
		t.Fatal("expected error for zero quantity, got nil")
	}
}

func TestValidateCreateOrderRequest_NegativeQuantity(t *testing.T) {
	req := CreateOrderRequest{Item: "widget", Quantity: -1, Price: 9.99}
	if err := validateCreateOrder(req); err == nil {
		t.Fatal("expected error for negative quantity, got nil")
	}
}

func TestValidateCreateOrderRequest_ZeroPrice(t *testing.T) {
	req := CreateOrderRequest{Item: "widget", Quantity: 1, Price: 0}
	if err := validateCreateOrder(req); err == nil {
		t.Fatal("expected error for zero price, got nil")
	}
}

func TestValidateCreateOrderRequest_NegativePrice(t *testing.T) {
	req := CreateOrderRequest{Item: "widget", Quantity: 1, Price: -1.5}
	if err := validateCreateOrder(req); err == nil {
		t.Fatal("expected error for negative price, got nil")
	}
}

func TestValidateCreateOrderRequest_ValidInput(t *testing.T) {
	req := CreateOrderRequest{Item: "widget", Quantity: 2, Price: 9.99}
	if err := validateCreateOrder(req); err != nil {
		t.Fatalf("expected no error for valid input, got: %v", err)
	}
}
