package main

import (
	"example3/internal/broker"
	"example3/internal/handler"
	"github.com/streadway/amqp"
	"log"
)

func main() {
	rmq, err := broker.NewRabbitMQBroker("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}

	err = rmq.SetupQueues("task-exchange", "task-queue", "task")
	if err != nil {
		log.Fatal("Failed to setup queues:", err)
	}

	// Publish a test message with retry count 0
	messageBody := []byte(`{"task": "do something important"}`)
	headers := amqp.Table{"x-retry-count": int32(0)}

	err = rmq.Publish("task-exchange", "task", messageBody, headers)
	if err != nil {
		log.Fatal("Failed to publish message:", err)
	}
	log.Println("Test message published")

	// Start consuming messages
	err = rmq.Consume("task-queue", handler.HandleTask)
	if err != nil {
		log.Fatal("Failed to start consumer:", err)
	}

	log.Println("Consumer started, waiting for messages...")
	select {} // keep main alive
}
