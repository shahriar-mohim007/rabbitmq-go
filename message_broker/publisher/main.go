// main_publisher.go
package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
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

	err = broker.Send(topics.PublisherConfig{
		RabbitMQ: topics.RabbitMQPublisher{
			Exchange:     "health-exchange",
			RoutingKey:   "health-routing-key",
			ContentType:  "text/plain",
			Payload:      []byte("Hello from publisher!"),
			DeliveryMode: amqp.Persistent,
		},
	})
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	log.Println("Message published successfully")
}
