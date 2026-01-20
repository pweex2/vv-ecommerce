package async

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Ensure RabbitMQ implements MessageQueue interface at compile time
var _ MessageQueue = (*RabbitMQ)(nil)

// RabbitMQ implements MessageQueue interface
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQ) Publish(topic string, payload []byte) error {
	// Declare the queue to ensure it exists
	q, err := r.channel.QueueDeclare(
		topic, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = r.channel.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         payload,
			DeliveryMode: amqp.Persistent, // Make message persistent
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (r *RabbitMQ) Subscribe(topic string, handler func(payload []byte) error) error {
	// Declare the queue
	q, err := r.channel.QueueDeclare(
		topic, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (IMPORTANT: manual ack)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			// Call handler
			err := handler(d.Body)
			if err != nil {
				// Requeue logic or Dead Letter Queue could be here
				log.Printf("Error processing message: %v", err)
				d.Ack(false) // Ack to avoid loop for now
			} else {
				d.Ack(false)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
