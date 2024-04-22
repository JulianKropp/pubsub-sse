package pubsubsse

import (
	"sync"
	"testing"
	"time"
)

// Tests for:
// +NewEventManager[T any]() *EventManager[T]
// +Listen(listener EventType[T]) string
// +Remove(id string)
// +GetListeners() []string
// +Emit(item *T)

// Create a new EventManager
func TestNewEventManager(t *testing.T) {
	manager := NewEventManager[int]()
	if manager == nil {
		t.Fatal("NewEventManager should not return nil")
	}
	if len(manager.listeners) != 0 {
		t.Errorf("New event manager should have no registered listeners")
	}
}

// Register a new listener
func TestEventManager_Listen(t *testing.T) {
	manager := NewEventManager[int]()
	received := false
	listener := func(t *int) {
		received = true
		*t = *t + 1
	}
	id := manager.Listen(listener)
	if id == "" {
		t.Errorf("Listen should return a non-empty ID")
	}

	manager.Emit(new(int))
	time.Sleep(2 * time.Second)
	if !received {
		t.Errorf("Listener was not triggered by Emit")
	}
}

// Remove a listener
func TestEventManager_Remove(t *testing.T) {
	manager := NewEventManager[int]()
	received := false
	listener := func(t *int) {
		received = true
	}
	id := manager.Listen(listener)
	manager.Remove(id)
	manager.Emit(new(int))
	if received {
		t.Errorf("Listener was not removed")
	}
}

// Get all listeners
func TestEventManager_GetListeners(t *testing.T) {
	manager := NewEventManager[int]()
	listener1 := func(t *int) {}
	id := manager.Listen(listener1)
	listeners := manager.GetListeners()
	if len(listeners) != 1 {
		t.Errorf("GetListeners should return 1 listener")
	}
	if listeners[0] != id {
		t.Errorf("GetListeners should return the correct listener ID")
	}

	// create a new listener
	listener2 := func(t *int) {}
	manager.Listen(listener2)
	listeners = manager.GetListeners()
	if len(listeners) != 2 {
		t.Errorf("GetListeners should return 2 listeners")
	}
}

// Emit an event
func TestEventManager_Emit(t *testing.T) {
	manager := NewEventManager[int]()
	var received int
	listener := func(t *int) {
		received = *t
	}
	manager.Listen(listener)
	item := 42
	manager.Emit(&item)
	time.Sleep(2 * time.Second)
	if received != item {
		t.Errorf("Emit should pass the correct item to the listener")
	}
}

func TestConcurrentAccess(t *testing.T) {
	manager := NewEventManager[int]()
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			manager.Listen(func(t *int) { *t = *t + 1 })
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			listeners := manager.GetListeners()
			if len(listeners) > 0 {
				manager.Remove(listeners[0])
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			manager.Emit(new(int))
		}
	}()
	wg.Wait()
}