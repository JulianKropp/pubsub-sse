package pubsubsse

import (
	"testing"
)

// Tests for:
// +GetID(): string
// +GetID(): string
// +GetTopics(): map[string]*topic
// +GetTopicByID(name string): *topic, bool
// +GetInstances(): map[string]*instance
// +GetInstanceByID(id string): *instance, bool
// +NewTopic(name string): *topic
// +RemoveTopic(t *topic)
// +AddInstance(c *instance)
// +RemoveInstance(c *instance)

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

// TestGroup_GetInstances tests the GetInstances function
func TestGroup_GetInstances(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if len(g.GetInstances()) > 0 {
		t.Error("GetInstances returned a non-empty map")
	}

	instance1 := ssePubSub.NewInstance()
	instance2 := ssePubSub.NewInstance()
	g.AddInstance(instance1)
	instances := g.GetInstances()
	if len(instances) != 1 {
		t.Error("GetInstances returned an empty map")
	}
	if instances[instance1.GetID()] != instance1 {
		t.Error("GetInstances returned the wrong instance")
	}
	if instances[instance2.GetID()] != nil {
		t.Error("GetInstances returned a non-existent instance")
	}
}

// TestGroup_GetInstanceByID tests the GetInstanceByID function
func TestGroup_GetInstanceByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if _, ok := g.GetInstanceByID("test"); ok {
		t.Error("GetInstanceByID returned true for a non-existent instance")
	}

	instance := ssePubSub.NewInstance()
	g.AddInstance(instance)
	instanceg, ok := g.GetInstanceByID(instance.GetID())
	if !ok {
		t.Error("GetInstanceByID returned false for an existing instance")
	}
	if instanceg != instance {
		t.Error("GetInstanceByID returned the wrong instance")
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
	if topic != nil && topic.GetID() != topic.id {
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

// TestGroup_AddInstance tests the AddInstance function
func TestGroup_AddInstance(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	if len(g.GetInstances()) > 0 {
		t.Error("GetInstances returned a non-empty map")
	}

	instance := ssePubSub.NewInstance()
	g.AddInstance(instance)
	if len(g.GetInstances()) != 1 {
		t.Error("AddInstance did not add the instance")
	}
}

// TestGroup_RemoveInstance tests the RemoveInstance function
func TestGroup_RemoveInstance(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	g := ssePubSub.NewGroup()
	instance := ssePubSub.NewInstance()
	g.AddInstance(instance)
	if len(g.GetInstances()) != 1 {
		t.Error("AddInstance did not add the instance")
	}

	g.RemoveInstance(instance)
	if len(g.GetInstances()) > 0 {
		t.Error("RemoveInstance did not remove the instance")
	}
}
