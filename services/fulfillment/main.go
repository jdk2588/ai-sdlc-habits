package main

import (
	"log"
	"os"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Fatal("fulfillment: RABBITMQ_URL is required")
	}

	conn, ch, err := setupAMQP(rabbitURL)
	if err != nil {
		log.Fatalf("fulfillment: failed to set up AMQP: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	pub := &amqpPublisher{ch: ch}

	if err := startConsuming(ch, pub); err != nil {
		log.Fatalf("fulfillment: consumer error: %v", err)
	}
}
