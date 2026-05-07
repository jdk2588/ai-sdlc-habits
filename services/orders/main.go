package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	var pub publisher

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL != "" {
		p, err := newAMQPPublisher(rabbitURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "orders: failed to connect to RabbitMQ: %v\n", err)
			os.Exit(1)
		}
		pub = p
		fmt.Println("orders: connected to RabbitMQ")
	} else {
		fmt.Fprintln(os.Stderr, "orders: RABBITMQ_URL not set, publishing disabled")
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
	fmt.Printf("orders: listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "orders: server error: %v\n", err)
		os.Exit(1)
	}
}
