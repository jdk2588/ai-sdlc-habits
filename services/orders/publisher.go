package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type publisher interface {
	Publish(routingKey string, body []byte) error
}

type amqpPublisher struct {
	ch *amqp.Channel
}

func (p *amqpPublisher) Publish(routingKey string, body []byte) error {
	return p.ch.Publish(
		"orders",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func newAMQPPublisher(url string) (*amqpPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if _, err = ch.QueueDeclare("order.placed", true, false, false, false, nil); err != nil {
		return nil, err
	}
	return &amqpPublisher{ch: ch}, nil
}
