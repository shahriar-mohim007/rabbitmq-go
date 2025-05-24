package message_broker

import (
	"context"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	brokerTopics "message_broker/pkg/message_broker/topics"
	"sync"
)

type RabbitMQBroker struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	callbackTasks  map[string]func(args ...string) error
	callbackTasksM sync.RWMutex
}

func NewRabbitMQBroker(amqpURI string) (*RabbitMQBroker, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQBroker{
		conn:          conn,
		channel:       ch,
		callbackTasks: make(map[string]func(args ...string) error),
	}, nil
}

func (r *RabbitMQBroker) Register(handlerMap map[brokerTopics.SubscriptionConfig]func(ctx context.Context, task []byte) error) error {
	for subConfig, handler := range handlerMap {
		rmqConf := subConfig.RabbitMQ

		// Declare durable exchange
		if rmqConf.Exchange != "" {
			err := r.channel.ExchangeDeclare(
				rmqConf.Exchange,
				"direct", // or "fanout", "topic"
				true,     // durable — must be true for persistence
				false,    // auto-deleted
				false,    // internal
				false,    // no-wait
				nil,
			)
			if err != nil {
				return fmt.Errorf("failed to declare exchange %s: %w", rmqConf.Exchange, err)
			}
		}

		// Declare durable queue (durable: true, autoDelete: false, exclusive: false)
		queue, err := r.channel.QueueDeclare(
			rmqConf.QueueName,
			true,  // durable — ensure queue survives broker restart
			false, // autoDelete — false to keep queue after consumers disconnect
			false, // exclusive — false so multiple consumers can connect and queue survives
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", rmqConf.QueueName, err)
		}

		// Bind queue to exchange with routing key
		if rmqConf.Exchange != "" && rmqConf.RoutingKey != "" {
			err = r.channel.QueueBind(
				queue.Name,
				rmqConf.RoutingKey,
				rmqConf.Exchange,
				false,
				nil,
			)
			if err != nil {
				return fmt.Errorf("failed to bind queue %s to exchange %s with routing key %s: %w", queue.Name, rmqConf.Exchange, rmqConf.RoutingKey, err)
			}
		}

		// Consume with manual ack
		msgs, err := r.channel.Consume(
			queue.Name,
			"",    // consumer tag
			false, // autoAck — false for manual ack/nack, prevents message loss
			false, // exclusive — false so multiple consumers possible
			false, // noLocal — not supported by RabbitMQ, set false
			false, // noWait
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to register consumer for queue %s: %w", queue.Name, err)
		}

		go func(msgs <-chan amqp.Delivery, handlerFunc func(ctx context.Context, task []byte) error) {
			for d := range msgs {
				ctx := context.Background()
				err := handlerFunc(ctx, d.Body)
				if err != nil {
					_ = d.Nack(false, true) // requeue on error
				} else {
					_ = d.Ack(false) // manual ack on success
				}
			}
		}(msgs, handler)
	}

	return nil
}

func (r *RabbitMQBroker) RegisterCallbackTask(taskName string, callBackFunc func(args ...string) error) error {
	r.callbackTasksM.Lock()
	defer r.callbackTasksM.Unlock()

	if _, exists := r.callbackTasks[taskName]; exists {
		return fmt.Errorf("callback task %s already registered", taskName)
	}

	r.callbackTasks[taskName] = callBackFunc
	return nil
}

func (r *RabbitMQBroker) Send(config brokerTopics.PublisherConfig) error {
	p := config.RabbitMQ
	if p.Exchange == "" {
		p.Exchange = brokerTopics.DefaultExchange
	}
	if p.RoutingKey == "" {
		p.RoutingKey = brokerTopics.DefaultRoutingKey
	}

	err := r.channel.Publish(
		p.Exchange,
		p.RoutingKey,
		p.Mandatory,
		p.Immediate,
		amqp.Publishing{
			DeliveryMode: p.DeliveryMode,
			Priority:     p.Priority,
			ContentType:  p.ContentType,
			Headers:      p.Headers,
			Body:         p.Payload,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (r *RabbitMQBroker) Retry(err error) error {
	if err != nil {
		fmt.Println("Retrying after error:", err)
		return nil
	}
	return errors.New("no error to retry")
}

func (r *RabbitMQBroker) Close() error {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
