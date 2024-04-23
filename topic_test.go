package pubsubsse

import (
	"testing"
)

// Tests for:
// +GetID(): string
// +GetID(): string
// +GetType(): string
// +GetClients(): map[string]*client
// +IsSubscribed(c *client): bool
// +Pub(msg interface): error
// -addClient(c *client)
// -removeClient(c *client)

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

// TestGetClients tests the GetClients() method.
func TestGetClients(t *testing.T) {
	topic := newTopic("public")
	if len(topic.GetClients()) != 0 {
		t.Error("Expected topic to have no clients")
	}

	ssePubSub := NewSSEPubSubService()
	c1 := ssePubSub.NewClient()
	c2 := ssePubSub.NewClient()
	topic.addClient(c1)
	topic.addClient(c2)

	if len(topic.GetClients()) != 2 {
		t.Error("Expected topic to have 2 clients")
	}
	if topic.GetClients()[c1.GetID()] != c1 {
		t.Error("Expected topic to have client c1")
	}
	if topic.GetClients()[c2.GetID()] != c2 {
		t.Error("Expected topic to have client c2")
	}
}

// TestIsSubscribed tests the IsSubscribed() method.
func TestIsSubscribed(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c1 := ssePubSub.NewClient()

	topic := c1.NewPrivateTopic()

	if topic.IsSubscribed(c1) {
		t.Error("Expected client to not be subscribed")
	}

	if err := c1.Sub(topic); err != nil {
		t.Error("Expected client to subscribe to topic")
	}

	if !topic.IsSubscribed(c1) {
		t.Error("Expected client to be subscribed")
	}

	if err := c1.Unsub(topic); err != nil {
		t.Error("Expected client to unsubscribe from topic")
	}

	if topic.IsSubscribed(c1) {
		t.Error("Expected client to not be subscribed")
	}
}

// TestPub tests the Pub() method.
func TestPub(t *testing.T) {
	ssePubSub := NewSSEPubSubService()

	c1 := ssePubSub.NewClient()
	c2 := ssePubSub.NewClient()

	// Start /event
	startEventServer(ssePubSub, t, 8082)

	// Start the client 1
	connected1 := make(chan bool)
	done1 := make(chan bool)
	data1 := []eventData{}
	httpToEvent(t, c1, 8082, connected1, done1, &data1)

	// Start the client 1
	connected2 := make(chan bool)
	done2 := make(chan bool)
	data2 := []eventData{}
	httpToEvent(t, c2, 8082, connected2, done2, &data2)

	// Wait for the clients to connect
	<-connected1
	<-connected2

	topic := c1.NewPrivateTopic()

	if err := c1.Sub(topic); err != nil {
		t.Error("Expected client to subscribe to topic")
	}
	if err := c2.Sub(topic); err == nil {
		t.Error("Expected client not to subscribe to topic")
	}

	testData1 := "testdata"
	if err := topic.Pub(testData1); err != nil {
		t.Error("Expected topic to publish message")
	}

	if err := c1.Unsub(topic); err != nil {
		t.Error("Expected client to unsubscribe from topic")
	}
	if err := c2.Unsub(topic); err == nil {
		t.Error("Expected client to not unsubscribe from topic")
	}

	if err := topic.Pub("test"); err != nil {
		t.Error("Expected topic to publish message")
	}

	// Wait for the clients to disconnect
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

// TestAddClient tests the addClient() method.
func TestAddClient(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c1 := ssePubSub.NewClient()

	topic := c1.NewPrivateTopic()

	if len(topic.GetClients()) != 0 {
		t.Error("Expected topic to have no clients")
	}

	topic.addClient(c1)

	if len(topic.GetClients()) != 1 {
		t.Error("Expected topic to have 1 client")
	}
	if topic.GetClients()[c1.GetID()] != c1 {
		t.Error("Expected topic to have client c1")
	}
}

// TestRemoveClient tests the removeClient() method.
func TestRemoveClient(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c1 := ssePubSub.NewClient()

	topic := c1.NewPrivateTopic()

	if len(topic.GetClients()) != 0 {
		t.Error("Expected topic to have no clients")
	}

	topic.addClient(c1)

	if len(topic.GetClients()) != 1 {
		t.Error("Expected topic to have 1 client")
	}
	if topic.GetClients()[c1.GetID()] != c1 {
		t.Error("Expected topic to have client c1")
	}

	topic.removeClient(c1)

	if len(topic.GetClients()) != 0 {
		t.Error("Expected topic to have no clients")
	}
}
