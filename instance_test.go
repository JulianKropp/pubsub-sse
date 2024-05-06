package pubsubsse

import (
	"testing"
	"time"
)

// TestInstance_GetStatus tests Instance.GetStatus()
func TestInstance_GetID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c := ssePubSub.NewInstance()
	if c.GetID() != c.id {
		t.Errorf("Instance.GetID() != c.id")
	}
}

// GetConnectionID
func TestInstance_GetConnectionID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c := ssePubSub.NewInstance()
	if c.GetConnectionID() != c.connection.id {
		t.Errorf("Instance.GetConnectionID() != c.connection.id")
	}
}

// getConnection
func TestInstance_getConnection(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c := ssePubSub.NewInstance()
	if c.getConnection() != c.connection {
		t.Errorf("Instance.getConnection() != c.connection")
	}
}

// ChangeConnection
func TestInstance_ChangeConnection(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c := ssePubSub.NewInstance()
	c2 := ssePubSub.NewInstance()
	c.ChangeConnection(c2.connection.GetID())
	if c.connection != c2.connection {
		t.Errorf("Instance.connection != c2.connection")
	}
}

// TestInstance_GetStatus tests Instance.GetStatus()
func TestInstance_GetStatus(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	ssePubSub.SetInstanceTimeout(1 * time.Second)
	instance := ssePubSub.NewInstance()
	if instance.GetStatus() != Created {
		t.Errorf("Instance.GetStatus() != Created")
	}

	// Start /event
	startEventServer(ssePubSub, t, 8080)

	// Start the instance
	done := make(chan bool)
	connected := make(chan bool)
	httpToEvent(t, instance, 8080, connected, done, &[]connectionData{})
	<-connected

	if instance.GetStatus() != Receiving {
		t.Errorf("Instance.GetStatus() != Receving")
	}

	<-done

	time.Sleep(100 * time.Millisecond)

	if instance.GetStatus() != Waiting {
		t.Errorf("Instance.GetStatus() != Waiting: got %d", instance.GetStatus())
	}

	time.Sleep(3 * time.Second)

	if instance.GetStatus() != Timeout {
		t.Errorf("Instance.GetStatus() != Timeout")
	}
}

// -----------------------------
// Public Topics
// -----------------------------

// TestInstance_GetPublicTopics tests Instance.GetPublicTopics()
func TestInstance_GetPublicTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	if len(instance.GetPublicTopics()) != 0 {
		t.Errorf("len(instance.GetPublicTopics()) != 0")
	}

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Get topic
	pubTopics := instance.GetPublicTopics()
	if len(pubTopics) != 1 {
		t.Errorf("len(pubTopics) != 1")
	}
	if pubTopics[pubTopicID] != pubTopic {
		t.Errorf("pubTopics[\"%v\"] != \"%v\"", pubTopics[pubTopicID], pubTopic)
	}
}

// TestInstance_GetPublicTopicByID tests Instance.GetPublicTopicByID()
func TestInstance_GetPublicTopicByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()

	// Get topic
	topic, ok := instance.GetPublicTopicByID(pubTopic.GetID())
	if !ok {
		t.Errorf("!ok")
	}
	if topic != pubTopic {
		t.Errorf("topic != pubTopic")
	}
}

// -----------------------------
// Private Topics
// -----------------------------

// TestInstance_NewPrivateTopic tests Instance.NewPrivateTopic()
func TestInstance_NewPrivateTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()

	// Create a private topic
	privTopic := instance.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Get topic
	privTopics := instance.GetPrivateTopics()
	if len(privTopics) != 1 {
		t.Errorf("len(privTopics) != 1")
	}
	if privTopics[privTopicID] != privTopic {
		t.Errorf("privTopics[\"%v\"] != \"%v\"", privTopics[privTopicID], privTopic)
	}
}

// TestInstance_RemovePrivateTopic tests Instance.RemovePrivateTopic()
func TestInstance_RemovePrivateTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()

	// Create a private topic
	privTopic := instance.NewPrivateTopic()

	// Remove topic
	instance.RemovePrivateTopic(privTopic)

	// Get topic
	privTopics := instance.GetPrivateTopics()
	if len(privTopics) != 0 {
		t.Errorf("%d != 0", len(privTopics))
	}
}

// TestInstance_GetPrivateTopics tests Instance.GetPrivateTopics()
func TestInstance_GetPrivateTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	if len(instance.GetPrivateTopics()) != 0 {
		t.Errorf("len(instance.GetPrivateTopics()) != 0")
	}

	// Create a private topic
	privTopic := instance.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Get topic
	privTopics := instance.GetPrivateTopics()
	if len(privTopics) != 1 {
		t.Errorf("len(privTopics) != 1")
	}
	if privTopics[privTopicID] != privTopic {
		t.Errorf("privTopics[\"%v\"] != \"%v\"", privTopics[privTopicID], privTopic)
	}
}

// TestInstance_GetPrivateTopicByID tests Instance.GetPrivateTopicByID()
func TestInstance_GetPrivateTopicByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()

	// Create a private topic
	privTopic := instance.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Get topic
	topic, ok := instance.GetPrivateTopicByID(privTopicID)
	if !ok {
		t.Errorf("!ok")
	}
	if topic != privTopic {
		t.Errorf("topic != privTopic")
	}
}

/// -----------------------------
// Sub/Unsub
// -----------------------------

// TestInstance_Sub tests Instance.Sub()
func TestInstance_Sub(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Subscribe to topic
	instance.Sub(pubTopic)

	// Get topic
	topics := instance.GetSubscribedTopics()
	if len(topics) != 1 {
		t.Errorf("len(topics) != 1")
	}
	if topics[pubTopicID] != pubTopic {
		t.Errorf("topics[\"%v\"] != \"%v\"", pubTopicID, pubTopic)
	}
}

// TestInstance_Unsub tests Instance.Unsub()
func TestInstance_Unsub(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()

	// Subscribe to topic
	instance.Sub(pubTopic)

	// Unsubscribe from topic
	instance.Unsub(pubTopic)

	// Get topic
	topics := instance.GetSubscribedTopics()
	if len(topics) != 0 {
		t.Errorf("len(topics) != 0")
	}
}

// -----------------------------
// Groups
// -----------------------------

// TestInstance_GetGroups tests Instance.GetGroups()
func TestInstance_GetGroups(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	if len(instance.GetGroups()) != 0 {
		t.Errorf("len(instance.GetGroups()) != 0")
	}

	// Create a group
	group := ssePubSub.NewGroup()

	// Add instance to group
	group.AddInstance(instance)
	groupID := group.GetID()

	// Get group
	groups := instance.GetGroups()
	if len(groups) != 1 {
		t.Errorf("len(groups) != 1")
	}
	if groups[groupID] != group {
		t.Errorf("groups[\"%v\"] != \"%v\"", groups[groupID], group)
	}
}

// TestInstance_GetGroupByID tests Instance.GetGroupByID()
func TestInstance_GetGroupByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()

	// Create a group
	group := ssePubSub.NewGroup()

	// Add instance to group
	group.AddInstance(instance)
	groupID := group.GetID()

	// Get group
	g, ok := instance.GetGroupByID(groupID)
	if !ok {
		t.Errorf("!ok")
	}
	if g != group {
		t.Errorf("g != group")
	}
}


// addGroup
func TestInstance_addGroup(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	group := ssePubSub.NewGroup()
	group.AddInstance(instance)

	if group.instances[instance.GetID()] != instance {
		t.Errorf("group.instances[instance.GetID()] != instance")
	}

	if instance.groups[group.GetID()] != group {
		t.Errorf("instance.groups[group.GetID()] != group")
	}
}

// removeGroup
func TestInstance_removeGroup(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	group := ssePubSub.NewGroup()
	group.AddInstance(instance)

	group.RemoveInstance(instance)

	if group.instances[instance.GetID()] != nil {
		t.Errorf("group.instances[instance.GetID()] != nil")
	}

	if instance.groups[group.GetID()] != nil {
		t.Errorf("instance.groups[group.GetID()] != nil")
	}
}

// -----------------------------
// Topics
// -----------------------------

// TestInstance_GetAllTopics tests Instance.GetAllTopics()
func TestInstance_GetAllTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	if len(instance.GetAllTopics()) != 0 {
		t.Errorf("len(instance.GetAllTopics()) != 0")
	}

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Create a private topic
	privTopic := instance.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Create group
	group := ssePubSub.NewGroup()
	group.AddInstance(instance)
	groupTopic := group.NewTopic()
	groupTopicID := groupTopic.GetID()

	// Get topic
	topics := instance.GetAllTopics()
	if len(topics) != 3 {
		t.Errorf("%d != 3", len(topics))
	}
	if topics[pubTopicID] != pubTopic {
		t.Errorf("topics[\"%v\"] != \"%v\"", pubTopicID, pubTopic)
	}
	if topics[privTopicID] != privTopic {
		t.Errorf("topics[\"%v\"] != \"%v\"", privTopicID, privTopic)
	}
	if topics[groupTopicID] != groupTopic {
		t.Errorf("topics[\"%v\"] != \"%v\"", groupTopicID, groupTopic)
	}
}

// TestInstance_GetTopicByID tests Instance.GetTopicByID()
func TestInstance_GetTopicByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()

	// Create a private topic
	privTopic := instance.NewPrivateTopic()

	// Create group
	group := ssePubSub.NewGroup()
	group.AddInstance(instance)
	groupTopic := group.NewTopic()

	// Get topic
	topic, ok := instance.GetTopicByID(pubTopic.GetID())
	if !ok {
		t.Errorf("!ok")
	}
	if topic != pubTopic {
		t.Errorf("topic != pubTopic")
	}

	topic, ok = instance.GetTopicByID(privTopic.GetID())
	if !ok {
		t.Errorf("!ok")
	}
	if topic != privTopic {
		t.Errorf("topic != privTopic")
	}

	topic, ok = instance.GetTopicByID(groupTopic.GetID())
	if !ok {
		t.Errorf("!ok")
	}
	if topic != groupTopic {
		t.Errorf("topic != groupTopic")
	}
}

// TestInstance_GetSubscribedTopics tests Instance.GetSubscribedTopics()
func TestInstance_GetSubscribedTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()
	if len(instance.GetSubscribedTopics()) != 0 {
		t.Errorf("len(instance.GetSubscribedTopics()) != 0")
	}

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Create a private topic
	privTopic := instance.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Create group
	group := ssePubSub.NewGroup()
	group.AddInstance(instance)
	groupTopic := group.NewTopic()
	groupTopicID := groupTopic.GetID()

	// Subscribe to topics
	instance.Sub(pubTopic)
	instance.Sub(privTopic)
	instance.Sub(groupTopic)

	// Get topic
	topics := instance.GetSubscribedTopics()
	if len(topics) != 3 {
		t.Errorf("%d != 3", len(topics))
	}
	if topics[pubTopicID] != pubTopic {
		t.Errorf("topics[\"%v\"] != \"%v\"", pubTopicID, pubTopic)
	}
	if topics[privTopicID] != privTopic {
		t.Errorf("topics[\"%v\"] != \"%v\"", privTopicID, privTopic)
	}
	if topics[groupTopicID] != groupTopic {
		t.Errorf("topics[\"%v\"] != \"%v\"", groupTopicID, groupTopic)
	}
}

// -----------------------------
// send
// -----------------------------

// TestInstance_send tests Instance.send()
func TestInstance_send(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	instance := ssePubSub.NewInstance()

	// Create topic and subscribe
	topic := instance.NewPrivateTopic()
	if err := instance.Sub(topic); err != nil {
		t.Error(err)
		return
	}

	// Start /event
	startEventServer(ssePubSub, t, 8081)

	// Start the instance
	connected := make(chan bool)
	done := make(chan bool)
	data := []connectionData{}
	httpToEvent(t, instance, 8081, connected, done, &data)
	<-connected

	testData := "testdata"

	if instance.GetStatus() != Receiving {
		t.Errorf("instance.GetStatus() != Receving")
		return
	}

	// Send data to instance
	if err := topic.Pub(testData); err != nil {
		t.Error(err)
		return
	}

	// Wait for the instance to receive the data
	<-done

	if len(data) < 1 {
		t.Errorf("len(data) < 1")
		return
	}
	for _, d := range data {
		if len(d.InstanceData) > 0 {
			if len(d.InstanceData[0].Data.Updates) > 0 {
				if d.InstanceData[0].Data.Updates[0].Data != testData {
					t.Errorf("data[].Updates[0].Data != testData")
					return
				}
				return
			}
		}
	}
	t.Errorf("No Updates received")

}
