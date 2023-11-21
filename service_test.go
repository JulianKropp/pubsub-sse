package pubsubsse

import (
	"testing"
)

// Tests for:
// +NewClient(): *client
// +RemoveClient(c *client)
// +GetClients(): map[string]*client
// +GetClientByID(id string): *client, bool

// +NewGroup(name string): *group
// +RemoveGroup(g *group)
// +GetGroups(): map[string]*group
// +GetGroupByName(name string): *group, bool

// +NewPublicTopic(name string): *topic
// +RemovePublicTopic(t *topic)
// +GetPublicTopics(): map[string]*topic
// +GetPublicTopicByName(name string): *topic, bool


// Create a new SSEPubSubService
func TestNewSSEPubSubService(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	if ssePubSub == nil {
		t.Error("SSEPubSubService not created")
	}
}

// --------------------------------------------
// Clients
// --------------------------------------------

// Create a new client and get it by id
func TestSSEPubSubService_NewClient(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	clientc := ssePubSub.NewClient()
	client, ok := ssePubSub.GetClientByID(clientc.GetID())
	if !ok {
		t.Error("Client not created: not found")
	}
	if client == nil {
		t.Error("Client not created: nil")
	}
	if clientc != client {
		t.Error("Client not created: wrong pointer")
	}
}

// Create a new client and remove it
func TestSSEPubSubService_RemoveClient(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()
	ssePubSub.RemoveClient(client)
	_, ok := ssePubSub.GetClientByID(client.GetID())
	if ok {
		t.Error("Client not removed: found")
	}
	if len(ssePubSub.GetClients()) != 0 {
		t.Error("Client not removed: Multiple clients exist")
	}
}

// Create a new client and get all clients
func TestSSEPubSubService_GetClients(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	clientc := ssePubSub.NewClient()
	clients := ssePubSub.GetClients()
	if len(clients) != 1 {
		t.Error("Clients not found")
	}
	if clients[clientc.GetID()] == nil {
		t.Error("Client not found")
	}
	if clients[clientc.GetID()] != clientc {
		t.Error("Client not found: wrong pointer")
	}
}

// Create a new client and get it by id
func TestSSEPubSubService_GetClientByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	clientc := ssePubSub.NewClient()
	client, ok := ssePubSub.GetClientByID(clientc.GetID())
	if !ok {
		t.Error("Client not found")
	}
	if client == nil {
		t.Error("Client not found: nil")
	}
	if clientc != client {
		t.Error("Client not found: wrong pointer")
	}
}

// --------------------------------------------
// Groups
// --------------------------------------------

// Create a new group and get it by name
func TestSSEPubSubService_NewGroup(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	groupc := ssePubSub.NewGroup("test")
	group, ok := ssePubSub.GetGroupByName("test")
	if !ok {
		t.Error("Group not created: not found")
	}
	if group == nil {
		t.Error("Group not created: nil")
	}
	if groupc != group {
		t.Error("Group not created: wrong pointer")
	}
}

// Create a new group and remove it
func TestSSEPubSubService_RemoveGroup(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	group := ssePubSub.NewGroup("test")
	ssePubSub.RemoveGroup(group)
	_, ok := ssePubSub.GetGroupByName(group.GetName())
	if ok {
		t.Error("Group not removed")
	}
	if len(ssePubSub.GetGroups()) != 0 {
		t.Error("Group not removed: Multiple groups exist")
	}
}

// Create a new group and get all groups
func TestSSEPubSubService_GetGroups(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	groupc := ssePubSub.NewGroup("test")
	groups := ssePubSub.GetGroups()
	if len(groups) != 1 {
		t.Error("Groups not found")
	}
	if groups["test"] == nil {
		t.Error("Group not found: nil")
	}
	if groups["test"] != groupc {
		t.Error("Group not found: wrong pointer")
	}
}

// Create a new group and get it by name
func TestSSEPubSubService_GetGroupByName(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	groupc := ssePubSub.NewGroup("test")
	group, ok := ssePubSub.GetGroupByName("test")
	if !ok {
		t.Error("Group not found")
	}
	if group == nil {
		t.Error("Group not found: nil")
	}
	if groupc != group {
		t.Error("Group not found: wrong pointer")
	}
}

// --------------------------------------------
// Public Topics
// --------------------------------------------

// Create a new public topic and get it by name
func TestSSEPubSubService_NewPublicTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	topicc := ssePubSub.NewPublicTopic("test")
	topic, ok := ssePubSub.GetPublicTopicByName("test")
	if !ok {
		t.Error("Public topic not created: not found")
	}
	if topic == nil {
		t.Error("Public topic not created: nil")
	}
	if topicc != topic {
		t.Error("Public topic not created: wrong pointer")
	}
}

// Create a new public topic and remove it
func TestSSEPubSubService_RemovePublicTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	topic := ssePubSub.NewPublicTopic("test")
	ssePubSub.RemovePublicTopic(topic)
	_, ok := ssePubSub.GetPublicTopicByName(topic.GetName())
	if ok {
		t.Error("Public topic not removed")
	}
	if len(ssePubSub.GetPublicTopics()) != 0 {
		t.Error("Public topic not removed: Multiple public topics exist")
	}
}

// Create a new public topic and get all public topics
func TestSSEPubSubService_GetPublicTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	topicc := ssePubSub.NewPublicTopic("test")
	topics := ssePubSub.GetPublicTopics()
	if len(topics) != 1 {
		t.Error("Public topics not found")
	}
	if topics["test"] == nil {
		t.Error("Public topic not found: nil")
	}
	if topics["test"] != topicc {
		t.Error("Public topic not found: wrong pointer")
	}
}

// Create a new public topic and get it by name
func TestSSEPubSubService_GetPublicTopicByName(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	topicc := ssePubSub.NewPublicTopic("test")
	topic, ok := ssePubSub.GetPublicTopicByName("test")
	if !ok {
		t.Error("Public topic not found")
	}
	if topic == nil {
		t.Error("Public topic not found: nil")
	}
	if topicc != topic {
		t.Error("Public topic not found: wrong pointer")
	}
}
