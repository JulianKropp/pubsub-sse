package pubsubsse

import (
	"sync"

	"github.com/google/uuid"
)

// eventType represents a function type that takes a pointer of any type.
type eventType[T any] func(t *T)

// EventManager manages events for a specific type.
type eventManager[T any] struct {
	listeners map[string]eventType[T]
	lock      sync.RWMutex
}

// newEventManager creates a new eventManager instance.
func newEventManager[T any]() *eventManager[T] {
	return &eventManager[T]{
		listeners: make(map[string]eventType[T]),
	}
}

// On registers a new listener and returns its ID.
func (em *eventManager[T]) Listen(listener eventType[T]) string {
	em.lock.Lock()
	defer em.lock.Unlock()

	// Generate a unique ID for the listener
	id := uuid.New().String()
	em.listeners[id] = listener
	return id
}

// Remove deletes a listener by its ID.
func (em *eventManager[T]) Remove(id string) {
	em.lock.Lock()
	defer em.lock.Unlock()
	delete(em.listeners, id)
}

// GetListeners returns all current listener IDs.
func (em *eventManager[T]) GetListeners() []string {
	em.lock.RLock()
	defer em.lock.RUnlock()

	ids := make([]string, 0, len(em.listeners))
	for id := range em.listeners {
		ids = append(ids, id)
	}
	return ids
}

// Emit fires all events of this type.
func (em *eventManager[T]) Emit(item *T) {
	em.lock.RLock()
	defer em.lock.RUnlock()
	for _, listener := range em.listeners {
		go listener(item)
	}
}
