package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.SetPrefix("publisher ")
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

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	log.Printf("connected, exchange=demo.topic, routing_key=demo.message, interval=%s", time.Second)

	for n := 1; ; n++ {
		select {
		case <-ctx.Done():
			log.Print("stopped")
			return
		case now := <-ticker.C:
			body := fmt.Sprintf("Hello RabbitMQ! #%d at %s", n, now.UTC().Format(time.RFC3339Nano))
			err := ch.Publish("demo.topic", "demo.message", false, false, amqp.Publishing{
				ContentType:  "text/plain",
				DeliveryMode: amqp.Persistent,
				Body:         []byte(body),
			})
			if err != nil {
				log.Fatalf("publish: %v", err)
			}
			log.Printf("sent exchange=demo.topic routing_key=demo.message body=%q", body)
		}
	}
}
