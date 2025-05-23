package broker

import (
	"context"
	"log"
	"time"

	"example3/internal/retry"
	"github.com/streadway/amqp"
)

type RabbitMQBroker struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQBroker(url string) (*RabbitMQBroker, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &RabbitMQBroker{conn: conn, channel: ch}, nil
}

func (r *RabbitMQBroker) SetupQueues(mainExchange, queueName, routingKey string) error {
	// DLQ setup
	dlx := mainExchange + ".dlx"
	dlq := queueName + ".dlq"

	err := r.channel.ExchangeDeclare(dlx, "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}
	_, err = r.channel.QueueDeclare(dlq, true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = r.channel.QueueBind(dlq, routingKey, dlx, false, nil)
	if err != nil {
		return err
	}

	// Main exchange + queue
	err = r.channel.ExchangeDeclare(mainExchange, "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}
	args := amqp.Table{
		"x-dead-letter-exchange":    dlx,
		"x-dead-letter-routing-key": routingKey,
	}
	_, err = r.channel.QueueDeclare(queueName, true, false, false, false, args)
	if err != nil {
		return err
	}
	return r.channel.QueueBind(queueName, routingKey, mainExchange, false, nil)
}

func (r *RabbitMQBroker) Consume(queue string, handler func(context.Context, []byte) error) error {
	msgs, err := r.channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			ctx := context.Background()
			err := handler(ctx, d.Body)

			if err != nil {
				attempts := d.Headers["x-retry-count"]
				count := 0
				if val, ok := attempts.(int32); ok {
					count = int(val)
				}

				if count >= 3 {
					log.Println("Sent to DLQ:", string(d.Body))
					_ = d.Reject(false) // false => send to DLQ
				} else {
					// Retry with backoff
					delay := retry.ExponentialBackoff(count)
					log.Printf("Retrying in %v (attempt %d)\n", delay, count+1)
					time.Sleep(delay)

					headers := amqp.Table{"x-retry-count": int32(count + 1)}
					err = r.channel.Publish(
						"task-exchange",
						"task",
						false,
						false,
						amqp.Publishing{
							ContentType: "application/json",
							Body:        d.Body,
							Headers:     headers,
						},
					)
					_ = d.Ack(false) // ack current so it's not redelivered
				}
			} else {
				_ = d.Ack(false)
			}
		}
	}()
	return nil
}
func (r *RabbitMQBroker) Publish(exchange, routingKey string, body []byte, headers amqp.Table) error {
	return r.channel.Publish(
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers:     headers,
		},
	)
}
