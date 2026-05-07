package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type placedMessage struct {
	OrderID  string `json:"order_id"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

func handleCreateOrder(pub publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := validateCreateOrder(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		order := &Order{
			Item:        req.Item,
			Quantity:    req.Quantity,
			Price:       req.Price,
			OrderStatus: "placed",
			CreatedAt:   time.Now(),
		}
		saveOrder(order)

		if pub != nil {
			body, err := json.Marshal(placedMessage{
				OrderID:  order.ID,
				Item:     order.Item,
				Quantity: order.Quantity,
			})
			if err != nil {
				log.Printf("orders: failed to marshal order.placed for %s: %v", order.ID, err)
			} else if err := pub.Publish("order.placed", body); err != nil {
				log.Printf("orders: failed to publish order.placed for %s: %v", order.ID, err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(order)
	}
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "missing order id")
		return
	}
	id := parts[1]

	order, ok := findOrder(id)
	if !ok {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
