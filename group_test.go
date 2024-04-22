package pubsubsse

import (
	"testing"
)

// Tests for:
// +GetID(): string
// +GetID(): string
// +GetTopics(): map[string]*topic
// +GetTopicByID(name string): *topic, bool
// +GetClients(): map[string]*client
// +GetClientByID(id string): *client, bool
// +NewTopic(name string): *topic
// +RemoveTopic(t *topic)
// +AddClient(c *client)
// +RemoveClient(c *client)

// TestGroup_NewGroup tests the NewGroup function
func TestGroup_NewGroup(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if g == nil {
		t.Error("NewGroup returned nil")
	}
}

// TestGroup_GetID tests the GetID function
func TestGroup_GetID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if g.GetID() == "" {
		t.Error("GetID returned an empty string")
	}
}

// TestGroup_GetTopics tests the GetTopics function
func TestGroup_GetTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if len(g.GetTopics()) > 0 {
		t.Error("GetTopics returned a non-empty map")
	}

	topic := g.NewTopic()
	topicID := topic.GetID()
	topics := g.GetTopics()
	if len(topics) != 1 {
		t.Error("GetTopics returned an empty map")
	}
	if topics[topicID] != topic {
		t.Error("GetTopics returned the wrong topic")
	}
}

// TestGroup_GetTopicByID tests the GetTopicByID function
func TestGroup_GetTopicByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if _, ok := g.GetTopicByID("test"); ok {
		t.Error("GetTopicByID returned true for a non-existent topic")
	}

	topic := g.NewTopic()
	topicg, ok := g.GetTopicByID(topic.GetID())
	if !ok {
		t.Error("GetTopicByID returned false for an existing topic")
	}
	if topicg != topic {
		t.Error("GetTopicByID returned the wrong topic")
	}
}

// TestGroup_GetClients tests the GetClients function
func TestGroup_GetClients(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if len(g.GetClients()) > 0 {
		t.Error("GetClients returned a non-empty map")
	}

	client1 := ssePubSub.NewClient()
	client2 := ssePubSub.NewClient()
	g.AddClient(client1)
	clients := g.GetClients()
	if len(clients) != 1 {
		t.Error("GetClients returned an empty map")
	}
	if clients[client1.GetID()] != client1 {
		t.Error("GetClients returned the wrong client")
	}
	if clients[client2.GetID()] != nil {
		t.Error("GetClients returned a non-existent client")
	}
}

// TestGroup_GetClientByID tests the GetClientByID function
func TestGroup_GetClientByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if _, ok := g.GetClientByID("test"); ok {
		t.Error("GetClientByID returned true for a non-existent client")
	}

	client := ssePubSub.NewClient()
	g.AddClient(client)
	clientg, ok := g.GetClientByID(client.GetID())
	if !ok {
		t.Error("GetClientByID returned false for an existing client")
	}
	if clientg != client {
		t.Error("GetClientByID returned the wrong client")
	}
}

// TestGroup_NewTopic tests the NewTopic function
func TestGroup_NewTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	topic := g.NewTopic()
	if topic == nil {
		t.Error("NewTopic returned nil")
	}
	if topic.GetID() != topic.id {
		t.Error("NewTopic returned the wrong name")
	}
}

// TestGroup_RemoveTopic tests the RemoveTopic function
func TestGroup_RemoveTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	topic := g.NewTopic()
	if len(g.GetTopics()) != 1 {
		t.Error("NewTopic did not add the topic")
	}

	g.RemoveTopic(topic)
	if len(g.GetTopics()) > 0 {
		t.Error("RemoveTopic did not remove the topic")
	}
}

// TestGroup_AddClient tests the AddClient function
func TestGroup_AddClient(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if len(g.GetClients()) > 0 {
		t.Error("GetClients returned a non-empty map")
	}

	client := ssePubSub.NewClient()
	g.AddClient(client)
	if len(g.GetClients()) != 1 {
		t.Error("AddClient did not add the client")
	}
}

// TestGroup_RemoveClient tests the RemoveClient function
func TestGroup_RemoveClient(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	client := ssePubSub.NewClient()
	g.AddClient(client)
	if len(g.GetClients()) != 1 {
		t.Error("AddClient did not add the client")
	}

	g.RemoveClient(client)
	if len(g.GetClients()) > 0 {
		t.Error("RemoveClient did not remove the client")
	}
}