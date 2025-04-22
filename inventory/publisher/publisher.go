package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type Product struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Variant     string `json:"variant"`
	Description string `json:"description"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"product_created", // queue name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	failOnError(err, "Failed to declare a queue")

	startTime := time.Now()

	for i := int64(1); i <= 1_000_000; i++ { // 10^6 loop
		product := Product{
			ID:          i,
			Type:        fmt.Sprintf("FoodType-%d", i),
			Variant:     fmt.Sprintf("Variant-%d", i),
			Description: fmt.Sprintf("Description for product %d", i),
		}

		body, err := json.Marshal(product)
		failOnError(err, "Failed to marshal product")

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
		failOnError(err, "Failed to publish a message")

		if i%10000 == 0 { // every 10k
			fmt.Printf("Published %d products...\n", i)
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("✅ Published 1,000,000 products successfully in %s\n", elapsed)
}
