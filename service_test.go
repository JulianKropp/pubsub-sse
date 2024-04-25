package pubsubsse

import (
	"testing"
	"time"
)

// Tests for:
// +NewInstance(): *instance
// +RemoveInstance(c *instance)
// +GetInstances(): map[string]*instance
// +GetInstanceByID(id string): *instance, bool

// +NewGroup(ID string): *group
// +RemoveGroup(g *group)
// +GetGroups(): map[string]*group
// +GetGroupByID(ID string): *group, bool

// +NewPublicTopic(ID string): *topic
// +RemovePublicTopic(t *topic)
// +GetPublicTopics(): map[string]*topic
// +GetPublicTopicByID(ID string): *topic, bool

// Get ID
func TestSSEPubSubService_GetID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	if ssePubSub.GetID() == "" {
		t.Error("ID not found")
	}
}

// Create a new SSEPubSubService
func TestNewSSEPubSubService(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	if ssePubSub == nil {
		t.Error("SSEPubSubService not created")
	}
}

// --------------------------------------------
// Instances
// --------------------------------------------

// Create a new instance and get it by id
func TestSSEPubSubService_NewInstance(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instancec := ssePubSub.NewInstance()
	instance, ok := ssePubSub.GetInstanceByID(instancec.GetID())
	if !ok {
		t.Error("Instance not created: not found")
	}
	if instance == nil {
		t.Error("Instance not created: nil")
	}
	if instancec != instance {
		t.Error("Instance not created: wrong pointer")
	}
}

// Create a new instance and remove it
func TestSSEPubSubService_RemoveInstance(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	ssePubSub.RemoveInstance(instance)
	_, ok := ssePubSub.GetInstanceByID(instance.GetID())
	if ok {
		t.Error("Instance not removed: found")
	}
	if len(ssePubSub.GetInstances()) != 0 {
		t.Error("Instance not removed: Multiple instances exist")
	}
}

// Create a new instance and get all instances
func TestSSEPubSubService_GetInstances(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instancec := ssePubSub.NewInstance()
	instances := ssePubSub.GetInstances()
	if len(instances) != 1 {
		t.Error("Instances not found")
	}
	if instances[instancec.GetID()] == nil {
		t.Error("Instance not found")
	}
	if instances[instancec.GetID()] != instancec {
		t.Error("Instance not found: wrong pointer")
	}
}

// Create a new instance and get it by id
func TestSSEPubSubService_GetInstanceByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instancec := ssePubSub.NewInstance()
	instance, ok := ssePubSub.GetInstanceByID(instancec.GetID())
	if !ok {
		t.Error("Instance not found")
	}
	if instance == nil {
		t.Error("Instance not found: nil")
	}
	if instancec != instance {
		t.Error("Instance not found: wrong pointer")
	}
}

// Test get/set instance timout
func TestSSEPubSubService_GetInstanceTimeout(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	ssePubSub.SetInstanceTimeout(5 * time.Second)
	if ssePubSub.GetInstanceTimeout() != 5*time.Second {
		t.Error("Instance timeout not set")
	}
}

// --------------------------------------------
// Groups
// --------------------------------------------

// Create a new group and get it by ID
func TestSSEPubSubService_NewGroup(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	groupc := ssePubSub.NewGroup()
	group, ok := ssePubSub.GetGroupByID(groupc.GetID())
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
	group := ssePubSub.NewGroup()
	ssePubSub.RemoveGroup(group)
	_, ok := ssePubSub.GetGroupByID(group.GetID())
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
	groupc := ssePubSub.NewGroup()
	groupcID := groupc.GetID()
	groups := ssePubSub.GetGroups()
	if len(groups) != 1 {
		t.Error("Groups not found")
	}
	if groups[groupcID] == nil {
		t.Error("Group not found: nil")
	}
	if groups[groupcID] != groupc {
		t.Error("Group not found: wrong pointer")
	}
}

// Create a new group and get it by ID
func TestSSEPubSubService_GetGroupByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	groupc := ssePubSub.NewGroup()
	group, ok := ssePubSub.GetGroupByID(groupc.GetID())
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

// Create a new public topic and get it by ID
func TestSSEPubSubService_NewPublicTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	topicc := ssePubSub.NewPublicTopic()
	topic, ok := ssePubSub.GetPublicTopicByID(topicc.GetID())
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
	topic := ssePubSub.NewPublicTopic()
	ssePubSub.RemovePublicTopic(topic)
	_, ok := ssePubSub.GetPublicTopicByID(topic.GetID())
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
	topicc := ssePubSub.NewPublicTopic()
	topiccID := topicc.GetID()
	topics := ssePubSub.GetPublicTopics()
	if len(topics) != 1 {
		t.Error("Public topics not found")
	}
	if topics[topiccID] == nil {
		t.Error("Public topic not found: nil")
	}
	if topics[topiccID] != topicc {
		t.Error("Public topic not found: wrong pointer")
	}
}

// Create a new public topic and get it by ID
func TestSSEPubSubService_GetPublicTopicByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	topicc := ssePubSub.NewPublicTopic()
	topic, ok := ssePubSub.GetPublicTopicByID(topicc.GetID())
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
