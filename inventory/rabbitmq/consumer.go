package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"inventory/service"
)

func StartConsumer(service service.InventoryService) {
	// 1. Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(fmt.Errorf("failed to connect to RabbitMQ: %w", err))
	}
	defer conn.Close()

	// 2. Open a channel
	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Errorf("failed to open a channel: %w", err))
	}
	defer ch.Close()

	// 3. Declare the queue
	q, err := ch.QueueDeclare(
		"product_created", // queue name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		panic(fmt.Errorf("failed to declare a queue: %w", err))
	}

	// 4. Start consuming messages
	msgs, err := ch.Consume(
		q.Name,
		"",
		false, // 🔥 manual ack enabled
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(fmt.Errorf("failed to register a consumer: %w", err))
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Println("📩 Received a message:", string(d.Body))

			err := service.HandleProductCreated(d.Body)
			if err != nil {
				fmt.Println("❌ Error processing message:", err)

				// 🔥 Manual Nack
				// requeue = true -> abar retry korbe
				err = d.Nack(false, true)
				if err != nil {
					fmt.Println("❗Failed to nack message:", err)
				}
			} else {
				fmt.Println("✅ Product saved successfully!")

				// 🔥 Manual Ack
				err = d.Ack(false)
				if err != nil {
					fmt.Println("❗Failed to ack message:", err)
				}
			}

			// Optional: Slow down processing if needed
			// time.Sleep(10 * time.Millisecond)
		}
	}()

	fmt.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
