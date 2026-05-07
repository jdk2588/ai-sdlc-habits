package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	var pub publisher

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL != "" {
		p, err := newAMQPPublisher(rabbitURL)
		if err != nil {
			log.Fatalf("orders: failed to connect to RabbitMQ: %v", err)
		}
		pub = p
		log.Println("orders: connected to RabbitMQ")
	} else {
		log.Println("orders: RABBITMQ_URL not set, publishing disabled")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleCreateOrder(pub)(w, r)
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	})
	mux.HandleFunc("/orders/", handleGetOrder)

	addr := ":8080"
	log.Printf("orders: listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("orders: server error: %v", err)
	}
}
