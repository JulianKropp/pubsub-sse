package pubsubsse

import (
	"testing"
	"time"
)

// removeInstance
func TestConnection_removeInstance(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	connection := instance.connection

	if instance.connection != connection {
		t.Errorf("Expected connection to be %v, got %v", connection, instance.connection)
	}

	ssePubSub.RemoveInstance(instance)

	if connection.GetInstances()[instance.GetID()] != nil {
		t.Errorf("Expected instance to be removed from connection, got %v", connection.GetInstances()[instance.GetID()])
	}
}

// GetID
func TestConnection_GetID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	connection := ssePubSub.NewInstance().connection

	if connection.GetID() != connection.id {
		t.Errorf("Expected connection id to be %v, got %v", connection.id, connection.GetID())
	}
}

// GetStatus
func TestConnection_GetStatus(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	connection := ssePubSub.NewInstance().connection

	if connection.GetStatus() != connection.status {
		t.Errorf("Expected connection status to be %v, got %v", connection.status, connection.GetStatus())
	}
}

// GetInstances
func TestConnection_GetInstances(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	connection := instance.connection

	// Add some instances
	for i := 0; i < 10; i++ {
		ssePubSub.NewInstance(instance.GetConnectionID())
	}

	for k, v := range connection.GetInstances() {
		if v != connection.instances[k] {
			t.Errorf("Expected instance to be %v, got %v", connection.instances[k], v)
		}
	}
}

// changeStatus
func TestConnection_changeStatus(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	connection := ssePubSub.NewInstance().connection

	conList := []Status{Created, Waiting, Receiving, Timeout, Stopped}

	for _, s := range conList {
		connection.changeStatus(s)
		if connection.GetStatus() != s {
			t.Errorf("Expected connection status to be %v, got %v", s, connection.GetStatus())
		}
	}
}

// timeoutCheck
func TestConnection_timeoutCheck(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	ssePubSub.SetInstanceTimeout(100 * time.Millisecond)
	connection := ssePubSub.NewInstance().connection

	time.Sleep(200 * time.Millisecond)

	connection.timeoutCheck()
	if connection.GetStatus() != Timeout {
		t.Errorf("Expected connection status to be %v, got %v", Timeout, connection.GetStatus())
	}
}