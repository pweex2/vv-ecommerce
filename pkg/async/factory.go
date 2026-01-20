package async

import (
	"log"
)

// NewRabbitMQOrMemory attempts to connect to RabbitMQ at the given URL.
// If the connection fails, it logs the error and returns a local in-memory message queue.
// This ensures the application can start even if the external broker is unavailable (e.g. in local dev).
func NewRabbitMQOrMemory(url string) MessageQueue {
	mq, err := NewRabbitMQ(url)
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v. Falling back to In-Memory Queue.", err)
		return NewMemoryQueue()
	}
	log.Println("Success: Connected to RabbitMQ.")
	return mq
}
