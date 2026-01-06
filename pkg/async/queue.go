package async

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// MessageQueue defines the interface for a simple message queue
type MessageQueue interface {
	Publish(topic string, payload []byte) error
	Subscribe(topic string, handler func(payload []byte) error) error
	Close() error
}

// MemoryQueue is a simple in-memory implementation of MessageQueue using channels
type MemoryQueue struct {
	topics map[string]chan []byte
	mu     sync.RWMutex
	done   chan struct{}
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		topics: make(map[string]chan []byte),
		done:   make(chan struct{}),
	}
}

func (q *MemoryQueue) Publish(topic string, payload []byte) error {
	q.mu.Lock()
	ch, ok := q.topics[topic]
	if !ok {
		// Buffer size of 100 for simplicity
		ch = make(chan []byte, 100)
		q.topics[topic] = ch
	}
	q.mu.Unlock()

	select {
	case ch <- payload:
		return nil
	case <-q.done:
		return errors.New("queue is closed")
	default:
		return errors.New("queue is full")
	}
}

func (q *MemoryQueue) Subscribe(topic string, handler func(payload []byte) error) error {
	q.mu.Lock()
	ch, ok := q.topics[topic]
	if !ok {
		ch = make(chan []byte, 100)
		q.topics[topic] = ch
	}
	q.mu.Unlock()

	// Start a worker for this topic
	go func() {
		for {
			select {
			case msg := <-ch:
				// Simple retry logic for the handler
				go func(m []byte) {
					// Try indefinitely until success or critical failure
					// In real world, we'd have DLQ (Dead Letter Queue)
					backoff := 1 * time.Second
					for {
						if err := handler(m); err == nil {
							return
						} else {
							fmt.Printf("Error handling message on topic %s: %v. Retrying in %v...\n", topic, err, backoff)
							time.Sleep(backoff)
							if backoff < 60*time.Second {
								backoff *= 2
							}
						}
					}
				}(msg)
			case <-q.done:
				return
			}
		}
	}()

	return nil
}

func (q *MemoryQueue) Close() error {
	close(q.done)
	return nil
}