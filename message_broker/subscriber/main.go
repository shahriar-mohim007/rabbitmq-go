// main_subscriber.go
package main

import (
	"context"
	"log"
	"message_broker/pkg/message_broker"
	"message_broker/pkg/message_broker/topics"
)

func main() {
	broker, err := message_broker.NewRabbitMQBroker("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer broker.Close()

	subConfig := topics.SubscriptionConfig{
		RabbitMQ: topics.RabbitMQSubscriber{
			QueueName:  "test-queue",
			Exchange:   "test-exchange",
			RoutingKey: "test-key",
			Durable:    true,
		},
	}

	err = broker.Register(map[topics.SubscriptionConfig]func(ctx context.Context, task []byte) error{
		subConfig: func(ctx context.Context, task []byte) error {
			log.Printf("Received message: %s", string(task))
			return nil
		},
	})
	if err != nil {
		log.Fatalf("Failed to register subscriber: %v", err)
	}

	log.Println("Listening for messages...")
	select {} // keep alive
}
