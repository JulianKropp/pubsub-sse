package pubsubsse

import (
	"testing"
)

// Tests for:
// +GetID(): string
// +GetID(): string
// +GetType(): string
// +GetInstances(): map[string]*instance
// +IsSubscribed(c *instance): bool
// +Pub(msg interface): error
// -addInstance(c *instance)
// -removeInstance(c *instance)

// TestGetID tests the GetID() method.
func TestGetID(t *testing.T) {
	topic := newTopic("public")
	if topic.GetID() == "" {
		t.Error("Expected topic ID to be non-empty")
	}
}

// TestGetType tests the GetType() method.
func TestGetType(t *testing.T) {
	topic1 := newTopic("public")
	if topic1.GetType() != "public" {
		t.Error("Expected topic type to be \"public\"")
	}

	topic2 := newTopic("private")
	if topic2.GetType() != "private" {
		t.Error("Expected topic type to be \"private\"")
	}

	topic3 := newTopic("group")
	if topic3.GetType() != "group" {
		t.Error("Expected topic type to be \"group\"")
	}
}

// TestGetInstances tests the GetInstances() method.
func TestGetInstances(t *testing.T) {
	topic := newTopic("public")
	if len(topic.GetInstances()) != 0 {
		t.Error("Expected topic to have no instances")
	}

	ssePubSub := NewSSEPubSubService()
	c1 := ssePubSub.NewInstance()
	c2 := ssePubSub.NewInstance()
	topic.addInstance(c1)
	topic.addInstance(c2)

	if len(topic.GetInstances()) != 2 {
		t.Error("Expected topic to have 2 instances")
	}
	if topic.GetInstances()[c1.GetID()] != c1 {
		t.Error("Expected topic to have instance c1")
	}
	if topic.GetInstances()[c2.GetID()] != c2 {
		t.Error("Expected topic to have instance c2")
	}
}

// TestIsSubscribed tests the IsSubscribed() method.
func TestIsSubscribed(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c1 := ssePubSub.NewInstance()

	topic := c1.NewPrivateTopic()

	if topic.IsSubscribed(c1) {
		t.Error("Expected instance to not be subscribed")
	}

	if err := c1.Sub(topic); err != nil {
		t.Error("Expected instance to subscribe to topic")
	}

	if !topic.IsSubscribed(c1) {
		t.Error("Expected instance to be subscribed")
	}

	if err := c1.Unsub(topic); err != nil {
		t.Error("Expected instance to unsubscribe from topic")
	}

	if topic.IsSubscribed(c1) {
		t.Error("Expected instance to not be subscribed")
	}
}

// TestPub tests the Pub() method.
func TestPub(t *testing.T) {
	ssePubSub := NewSSEPubSubService()

	c1 := ssePubSub.NewInstance()
	c2 := ssePubSub.NewInstance()

	// Start /event
	startEventServer(ssePubSub, t, 8082)

	// Start the instance 1
	connected1 := make(chan bool)
	done1 := make(chan bool)
	data1 := []eventData{}
	httpToEvent(t, c1, 8082, connected1, done1, &data1)

	// Start the instance 1
	connected2 := make(chan bool)
	done2 := make(chan bool)
	data2 := []eventData{}
	httpToEvent(t, c2, 8082, connected2, done2, &data2)

	// Wait for the instances to connect
	<-connected1
	<-connected2

	topic := c1.NewPrivateTopic()

	if err := c1.Sub(topic); err != nil {
		t.Error("Expected instance to subscribe to topic")
	}
	if err := c2.Sub(topic); err == nil {
		t.Error("Expected instance not to subscribe to topic")
	}

	testData1 := "testdata"
	if err := topic.Pub(testData1); err != nil {
		t.Error("Expected topic to publish message")
	}

	if err := c1.Unsub(topic); err != nil {
		t.Error("Expected instance to unsubscribe from topic")
	}
	if err := c2.Unsub(topic); err == nil {
		t.Error("Expected instance to not unsubscribe from topic")
	}

	if err := topic.Pub("test"); err != nil {
		t.Error("Expected topic to publish message")
	}

	// Wait for the instances to disconnect
	<-done1
	<-done2

	if len(data1) < 1 {
		t.Error("len(data) < 1")
		return
	}
	for _, d := range data1 {
		if len(d.Updates) > 0 {
			if d.Updates[0].Data != testData1 {
				t.Error("data[].Updates[0].Data != testData")
				return
			}
			return
		}
	}
	t.Error("No Updates received")

}

// TestAddInstance tests the addInstance() method.
func TestAddInstance(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c1 := ssePubSub.NewInstance()

	topic := c1.NewPrivateTopic()

	if len(topic.GetInstances()) != 0 {
		t.Error("Expected topic to have no instances")
	}

	topic.addInstance(c1)

	if len(topic.GetInstances()) != 1 {
		t.Error("Expected topic to have 1 instance")
	}
	if topic.GetInstances()[c1.GetID()] != c1 {
		t.Error("Expected topic to have instance c1")
	}
}

// TestRemoveInstance tests the removeInstance() method.
func TestRemoveInstance(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c1 := ssePubSub.NewInstance()

	topic := c1.NewPrivateTopic()

	if len(topic.GetInstances()) != 0 {
		t.Error("Expected topic to have no instances")
	}

	topic.addInstance(c1)

	if len(topic.GetInstances()) != 1 {
		t.Error("Expected topic to have 1 instance")
	}
	if topic.GetInstances()[c1.GetID()] != c1 {
		t.Error("Expected topic to have instance c1")
	}

	topic.removeInstance(c1)

	if len(topic.GetInstances()) != 0 {
		t.Error("Expected topic to have no instances")
	}
}
