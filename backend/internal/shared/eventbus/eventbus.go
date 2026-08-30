package eventbus

import (
	"sync"
)

// Event represents a generic domain event.
type Event struct {
	Type    string
	Payload interface{}
}

// EventHandler is a callback function for handling events.
type EventHandler func(event Event)

// EventBus provides in-memory publisher/subscriber capabilities.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]EventHandler
}

// NewEventBus initializes a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]EventHandler),
	}
}

// Subscribe registers a handler for a specific event type.
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

// Publish dispatches an event to all subscribed handlers synchronously or asynchronously.
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	handlers, exists := eb.subscribers[event.Type]
	eb.mu.RUnlock()

	if !exists {
		return
	}

	for _, handler := range handlers {
		// Execute handler in goroutine to decouple event execution
		go handler(event)
	}
}
