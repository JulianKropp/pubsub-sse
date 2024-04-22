package pubsubsse

import (
	"sync"
	"github.com/google/uuid"
)

// EventType represents a function type that takes a pointer of any type.
type EventType[T any] func(t *T)

// EventManager manages events for a specific type.
type EventManager[T any] struct {
	listeners map[string]EventType[T]
	lock      sync.RWMutex
}

// NewEventManager creates a new EventManager instance.
func NewEventManager[T any]() *EventManager[T] {
	return &EventManager[T]{
		listeners: make(map[string]EventType[T]),
	}
}

// On registers a new listener and returns its ID.
func (em *EventManager[T]) Listen(listener EventType[T]) string {
	em.lock.Lock()
	defer em.lock.Unlock()

	// Generate a unique ID for the listener
	id := uuid.New().String()
	em.listeners[id] = listener
	return id
}

// Remove deletes a listener by its ID.
func (em *EventManager[T]) Remove(id string) {
	em.lock.Lock()
	defer em.lock.Unlock()
	delete(em.listeners, id)
}

// GetListeners returns all current listener IDs.
func (em *EventManager[T]) GetListeners() []string {
	em.lock.RLock()
	defer em.lock.RUnlock()

	ids := make([]string, 0, len(em.listeners))
	for id := range em.listeners {
		ids = append(ids, id)
	}
	return ids
}

// Emit fires all events of this type.
func (em *EventManager[T]) Emit(item *T) {
	em.lock.RLock()
	defer em.lock.RUnlock()
	for _, listener := range em.listeners {
		go listener(item)
	}
}