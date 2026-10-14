package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.SetPrefix("consumer ")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := amqp.Dial("amqp://demo:demo@localhost:5672/")
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("open channel: %v", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare("demo.topic", "topic", true, false, false, false, nil); err != nil {
		log.Fatalf("declare exchange: %v", err)
	}

	q, err := ch.QueueDeclare("hello", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("declare queue: %v", err)
	}
	if err := ch.QueueBind(q.Name, "demo.*", "demo.topic", false, nil); err != nil {
		log.Fatalf("bind queue: %v", err)
	}

	messages, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("consume: %v", err)
	}
	log.Printf("connected, exchange=demo.topic, binding_key=demo.*, waiting for messages, queue=%s", q.Name)

	for {
		select {
		case <-ctx.Done():
			log.Print("stopped")
			return
		case message, ok := <-messages:
			if !ok {
				log.Fatalf("delivery channel closed")
			}
			log.Printf("received queue=%s exchange=%s routing_key=%s body=%q", q.Name, message.Exchange, message.RoutingKey, message.Body)

			if err := message.Ack(false); err != nil {
				log.Fatalf("ack: %v", err)
			}
		}
	}
}
