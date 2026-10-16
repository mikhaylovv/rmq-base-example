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
	log.SetPrefix("publisher-confirms ")
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

	if err := ch.Confirm(false); err != nil {
		log.Fatalf("enable publisher confirms: %v", err)
	}
	// The buffer holds every confirmation while the whole batch is published.
	const batchSize = 3
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, batchSize))

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	const confirmTimeout = 10 * time.Second
	log.Printf("connected, exchange=demo.topic, routing_key=demo.message, batch_size=%d, interval=%s, confirm_timeout=%s", batchSize, time.Second, confirmTimeout)

	for n := 1; ; {
		select {
		case <-ctx.Done():
			log.Print("stopped")
			return
		case now := <-ticker.C:
			pending := make(map[uint64]string, batchSize)
			for i := 0; i < batchSize; i++ {
				body := fmt.Sprintf("Hello RabbitMQ! #%d at %s", n, now.UTC().Format(time.RFC3339Nano))
				tag := ch.GetNextPublishSeqNo()
				err := ch.Publish("demo.topic", "demo.message", false, false, amqp.Publishing{
					ContentType:  "text/plain",
					DeliveryMode: amqp.Persistent,
					Body:         []byte(body),
				})
				if err != nil {
					panic(fmt.Sprintf("publish tag=%d: %v", tag, err))
				}
				pending[tag] = body
				log.Printf("sent tag=%d exchange=demo.topic routing_key=demo.message body=%q", tag, body)
				n++
			}

			timer := time.NewTimer(confirmTimeout)
			for len(pending) > 0 {
				select {
				case confirm, ok := <-confirms:
					if !ok {
						panic(fmt.Sprintf("confirmation channel closed: batch incomplete, unconfirmed=%v", pending))
					}
					body, ok := pending[confirm.DeliveryTag]
					if !ok {
						panic(fmt.Sprintf("unexpected confirmation tag=%d", confirm.DeliveryTag))
					}
					if !confirm.Ack {
						panic(fmt.Sprintf("broker nack tag=%d: batch incomplete, unconfirmed=%v", confirm.DeliveryTag, pending))
					}
					delete(pending, confirm.DeliveryTag)
					log.Printf("confirmed tag=%d body=%q", confirm.DeliveryTag, body)
				case <-timer.C:
					panic(fmt.Sprintf("batch confirmation timeout after %s: unconfirmed=%v", confirmTimeout, pending))
				}
			}
			timer.Stop()
			log.Printf("batch confirmed messages=%d", batchSize)
		}
	}
}
