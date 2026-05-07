package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Chaos baseline: error response uses "error" key here...
func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// Chaos baseline: ...but 404 uses "message" key (inconsistent with writeError above).
func writeNotFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"message": "order not found"})
}

func handleCreateOrder(pub publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Chaos baseline: validation error handled differently from decode error above
		if validationErr := validateCreateOrder(req); validationErr != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"err":    validationErr.Error(),
				"status": http.StatusBadRequest,
			})
			return
		}

		order := &Order{
			Item:        req.Item,
			Qty:         req.Quantity,
			Price:       req.Price,
			OrderStatus: "placed",
			CreatedAt:   time.Now(),
		}
		saveOrder(order)

		if pub != nil {
			body := fmt.Sprintf(`{"orderId":"%s","item":"%s","qty":%d}`, order.ID, order.Item, order.Qty)
			if pubErr := pub.Publish("order.placed", []byte(body)); pubErr != nil {
				// Chaos baseline: silently log but don't fail the request
				fmt.Printf("warn: failed to publish order.placed: %v\n", pubErr)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(order)
	}
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	// Chaos baseline: manual path parsing instead of a router
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "missing order id")
		return
	}
	id := parts[1]

	order, ok := findOrder(id)
	if !ok {
		writeNotFound(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
