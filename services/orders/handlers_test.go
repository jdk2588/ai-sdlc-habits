package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func resetStore() {
	ordersMu.Lock()
	defer ordersMu.Unlock()
	orders = make(map[string]*Order)
	nextID = 0
}

func TestCreateOrder_ValidInput_Returns201(t *testing.T) {
	resetStore()
	body := bytes.NewBufferString(`{"item":"widget","quantity":2,"price":9.99}`)
	req := httptest.NewRequest(http.MethodPost, "/orders", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleCreateOrder(nil)(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var order Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if order.ID == "" {
		t.Error("expected order ID to be set")
	}
	if order.OrderStatus != "placed" {
		t.Errorf("expected status 'placed', got %q", order.OrderStatus)
	}
}

func TestCreateOrder_InvalidInput_Returns400(t *testing.T) {
	resetStore()
	cases := []struct {
		name string
		body string
	}{
		{"empty item", `{"item":"","quantity":2,"price":9.99}`},
		{"zero quantity", `{"item":"widget","quantity":0,"price":9.99}`},
		{"negative price", `{"item":"widget","quantity":1,"price":-1}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handleCreateOrder(nil)(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("%s: expected 400, got %d", tc.name, w.Code)
			}
		})
	}
}

func TestGetOrder_KnownID_Returns200(t *testing.T) {
	resetStore()
	order := &Order{Item: "widget", Quantity: 1, Price: 5.0, OrderStatus: "placed"}
	saveOrder(order)

	req := httptest.NewRequest(http.MethodGet, "/orders/"+order.ID, nil)
	w := httptest.NewRecorder()
	handleGetOrder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var got Order
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ID != order.ID {
		t.Errorf("expected order ID %q, got %q", order.ID, got.ID)
	}
}

func TestGetOrder_UnknownID_Returns404(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodGet, "/orders/9999", nil)
	w := httptest.NewRecorder()
	handleGetOrder(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
