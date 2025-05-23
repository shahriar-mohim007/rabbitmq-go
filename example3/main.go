package main

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/streadway/amqp"
)

const (
	rabbitURL    = "amqp://guest:guest@localhost:5672/"
	queueName    = "demo-queue"
	exchangeName = "demo-exchange"
	routingKey   = "demo.key"
)

type RabbitConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	errChan <-chan *amqp.Error
	ctx     context.Context
	cancel  context.CancelFunc
}

func (rc *RabbitConsumer) connect() error {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	errChan := conn.NotifyClose(make(chan *amqp.Error))

	err = ch.ExchangeDeclare(
		exchangeName, "direct", true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		queueName, true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		queueName, routingKey, exchangeName, false, nil,
	)
	if err != nil {
		return err
	}

	rc.conn = conn
	rc.channel = ch
	rc.errChan = errChan
	return nil
}

func (rc *RabbitConsumer) consume() error {
	msgs, err := rc.channel.Consume(
		queueName, "", true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					log.Println("❌ Channel closed")
					return
				}
				log.Printf("📥 Received: %s", msg.Body)
			case <-rc.ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (rc *RabbitConsumer) handleReconnect() {
	go func() {
		for err := range rc.errChan {
			log.Println("⚠️ Connection lost. Attempting reconnect...", err)
			rc.reconnectWithBackoff()
		}
	}()
}

func (rc *RabbitConsumer) reconnectWithBackoff() {
	attempt := 0
	for {
		attempt++
		log.Printf("🔁 Reconnect attempt #%d", attempt)

		err := rc.connect()
		if err == nil {
			err = rc.consume()
			if err == nil {
				log.Println("✅ Reconnected and consuming")
				rc.handleReconnect()
				return
			}
		}

		wait := time.Duration(math.Min(60, math.Pow(2, float64(attempt)))) * time.Second
		log.Printf("⏳ Reconnect failed, retrying in %v", wait)
		time.Sleep(wait)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	rc := &RabbitConsumer{ctx: ctx, cancel: cancel}

	if err := rc.connect(); err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	if err := rc.consume(); err != nil {
		log.Fatalf("❌ Failed to consume: %v", err)
	}

	rc.handleReconnect()

	log.Println("🚀 Ready and listening for messages")
	select {} // block forever
}
