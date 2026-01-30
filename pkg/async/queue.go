package async

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// MessageQueue defines the interface for a simple message queue
type MessageQueue interface {
	Publish(topic string, payload []byte, traceHeaders map[string]string) error
	Subscribe(topic string, handler func(payload []byte, traceHeaders map[string]string) error) error
	Close() error
}

// Ensure MemoryQueue implements MessageQueue interface at compile time
var _ MessageQueue = (*MemoryQueue)(nil)

// MemoryQueue is a simple in-memory implementation of MessageQueue using channels
type MemoryQueue struct {
	topics map[string]chan message
	mu     sync.RWMutex
	done   chan struct{}
}

type message struct {
	payload []byte
	headers map[string]string
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		topics: make(map[string]chan message),
		done:   make(chan struct{}),
	}
}

func (q *MemoryQueue) Publish(topic string, payload []byte, traceHeaders map[string]string) error {
	q.mu.Lock()
	ch, ok := q.topics[topic]
	if !ok {
		// Buffer size of 100 for simplicity
		ch = make(chan message, 100)
		q.topics[topic] = ch
	}
	q.mu.Unlock()

	msg := message{
		payload: payload,
		headers: traceHeaders,
	}

	select {
	case ch <- msg:
		return nil
	case <-q.done:
		return errors.New("queue is closed")
	default:
		return errors.New("queue is full")
	}
}

func (q *MemoryQueue) Subscribe(topic string, handler func(payload []byte, traceHeaders map[string]string) error) error {
	q.mu.Lock()
	ch, ok := q.topics[topic]
	if !ok {
		ch = make(chan message, 100)
		q.topics[topic] = ch
	}
	q.mu.Unlock()

	// Start a worker for this topic
	go func() {
		for {
			select {
			case msg := <-ch:
				// Simple retry logic for the handler
				go func(m message) {
					// Try indefinitely until success or critical failure
					// In real world, we'd have DLQ (Dead Letter Queue)
					backoff := 1 * time.Second
					for {
						if err := handler(m.payload, m.headers); err == nil {
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
