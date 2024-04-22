package pubsubsse

import (
	"testing"
)

// -----------------------------
//ID/Status
// -----------------------------

// TestClient_GetStatus tests Client.GetStatus()
func TestClient_GetID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c := ssePubSub.NewClient()
	if c.GetID() == "" {
		t.Errorf("Client.GetID() == \"\"")
	}
}

// TestClient_GetStatus tests Client.GetStatus()
func TestClient_GetStatus(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()
	if client.GetStatus() != Waiting {
		t.Errorf("Client.GetStatus() != Waiting")
	}

	// Start /event
	startEventServer(ssePubSub, t, 8080)

	// Start the client
	done := make(chan bool)
	connected := make(chan bool)
	httpToEvent(t, client, 8080, connected, done, &[]eventData{})
	<-connected

	if client.GetStatus() != Receving {
		t.Errorf("Client.GetStatus() != Receving")
	}
}

// -----------------------------
// Public Topics
// -----------------------------

// TestClient_GetPublicTopics tests Client.GetPublicTopics()
func TestClient_GetPublicTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()
	if len(client.GetPublicTopics()) != 0 {
		t.Errorf("len(client.GetPublicTopics()) != 0")
	}

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Get topic
	pubTopics := client.GetPublicTopics()
	if len(pubTopics) != 1 {
		t.Errorf("len(pubTopics) != 1")
	}
	if pubTopics[pubTopicID] != pubTopic {
		t.Errorf("pubTopics[\"%v\"] != \"%v\"", pubTopics[pubTopicID], pubTopic)
	}
}

// TestClient_GetPublicTopicByID tests Client.GetPublicTopicByID()
func TestClient_GetPublicTopicByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()

	// Get topic
	topic, ok := client.GetPublicTopicByID(pubTopic.GetID())
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

// TestClient_NewPrivateTopic tests Client.NewPrivateTopic()
func TestClient_NewPrivateTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a private topic
	privTopic := client.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Get topic
	privTopics := client.GetPrivateTopics()
	if len(privTopics) != 1 {
		t.Errorf("len(privTopics) != 1")
	}
	if privTopics[privTopicID] != privTopic {
		t.Errorf("privTopics[\"%v\"] != \"%v\"", privTopics[privTopicID], privTopic)
	}
}

// TestClient_RemovePrivateTopic tests Client.RemovePrivateTopic()
func TestClient_RemovePrivateTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a private topic
	privTopic := client.NewPrivateTopic()

	// Remove topic
	client.RemovePrivateTopic(privTopic)

	// Get topic
	privTopics := client.GetPrivateTopics()
	if len(privTopics) != 0 {
		t.Errorf("%d != 0", len(privTopics))
	}
}

// TestClient_GetPrivateTopics tests Client.GetPrivateTopics()
func TestClient_GetPrivateTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()
	if len(client.GetPrivateTopics()) != 0 {
		t.Errorf("len(client.GetPrivateTopics()) != 0")
	}

	// Create a private topic
	privTopic := client.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Get topic
	privTopics := client.GetPrivateTopics()
	if len(privTopics) != 1 {
		t.Errorf("len(privTopics) != 1")
	}
	if privTopics[privTopicID] != privTopic {
		t.Errorf("privTopics[\"%v\"] != \"%v\"", privTopics[privTopicID], privTopic)
	}
}

// TestClient_GetPrivateTopicByID tests Client.GetPrivateTopicByID()
func TestClient_GetPrivateTopicByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a private topic
	privTopic := client.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Get topic
	topic, ok := client.GetPrivateTopicByID(privTopicID)
	if !ok {
		t.Errorf("!ok")
	}
	if topic != privTopic {
		t.Errorf("topic != privTopic")
	}
}

// -----------------------------
// Groups
// -----------------------------

// TestClient_GetGroups tests Client.GetGroups()
func TestClient_GetGroups(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()
	if len(client.GetGroups()) != 0 {
		t.Errorf("len(client.GetGroups()) != 0")
	}

	// Create a group
	group := ssePubSub.NewGroup()

	// Add client to group
	group.AddClient(client)
	groupID := group.GetID()

	// Get group
	groups := client.GetGroups()
	if len(groups) != 1 {
		t.Errorf("len(groups) != 1")
	}
	if groups[groupID] != group {
		t.Errorf("groups[\"%v\"] != \"%v\"", groups[groupID], group)
	}
}

// TestClient_GetGroupByID tests Client.GetGroupByID()
func TestClient_GetGroupByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a group
	group := ssePubSub.NewGroup()

	// Add client to group
	group.AddClient(client)
	groupID := group.GetID()

	// Get group
	g, ok := client.GetGroupByID(groupID)
	if !ok {
		t.Errorf("!ok")
	}
	if g != group {
		t.Errorf("g != group")
	}
}

// -----------------------------
// Topics
// -----------------------------

// TestClient_GetAllTopics tests Client.GetAllTopics()
func TestClient_GetAllTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()
	if len(client.GetAllTopics()) != 0 {
		t.Errorf("len(client.GetAllTopics()) != 0")
	}

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Create a private topic
	privTopic := client.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Create group
	group := ssePubSub.NewGroup()
	group.AddClient(client)
	groupTopic := group.NewTopic()
	groupTopicID := groupTopic.GetID()

	// Get topic
	topics := client.GetAllTopics()
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

// TestClient_GetTopicByID tests Client.GetTopicByID()
func TestClient_GetTopicByID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()

	// Create a private topic
	privTopic := client.NewPrivateTopic()

	// Create group
	group := ssePubSub.NewGroup()
	group.AddClient(client)
	groupTopic := group.NewTopic()

	// Get topic
	topic, ok := client.GetTopicByID(pubTopic.GetID())
	if !ok {
		t.Errorf("!ok")
	}
	if topic != pubTopic {
		t.Errorf("topic != pubTopic")
	}

	topic, ok = client.GetTopicByID(privTopic.GetID())
	if !ok {
		t.Errorf("!ok")
	}
	if topic != privTopic {
		t.Errorf("topic != privTopic")
	}

	topic, ok = client.GetTopicByID(groupTopic.GetID())
	if !ok {
		t.Errorf("!ok")
	}
	if topic != groupTopic {
		t.Errorf("topic != groupTopic")
	}
}

// TestClient_GetSubscribedTopics tests Client.GetSubscribedTopics()
func TestClient_GetSubscribedTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()
	if len(client.GetSubscribedTopics()) != 0 {
		t.Errorf("len(client.GetSubscribedTopics()) != 0")
	}

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Create a private topic
	privTopic := client.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Create group
	group := ssePubSub.NewGroup()
	group.AddClient(client)
	groupTopic := group.NewTopic()
	groupTopicID := groupTopic.GetID()

	// Subscribe to topics
	client.Sub(pubTopic)
	client.Sub(privTopic)
	client.Sub(groupTopic)

	// Get topic
	topics := client.GetSubscribedTopics()
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
// Sub/Unsub
// -----------------------------

// TestClient_Sub tests Client.Sub()
func TestClient_Sub(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Subscribe to topic
	client.Sub(pubTopic)

	// Get topic
	topics := client.GetSubscribedTopics()
	if len(topics) != 1 {
		t.Errorf("len(topics) != 1")
	}
	if topics[pubTopicID] != pubTopic {
		t.Errorf("topics[\"%v\"] != \"%v\"", pubTopicID, pubTopic)
	}
}

// TestClient_Unsub tests Client.Unsub()
func TestClient_Unsub(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()

	// Subscribe to topic
	client.Sub(pubTopic)

	// Unsubscribe from topic
	client.Unsub(pubTopic)

	// Get topic
	topics := client.GetSubscribedTopics()
	if len(topics) != 0 {
		t.Errorf("len(topics) != 0")
	}
}

// -----------------------------
// Private
// -----------------------------

// TestClient_send tests Client.send()
func TestClient_send(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create topic and subscribe
	topic := client.NewPrivateTopic()
	if err := client.Sub(topic); err != nil {
		t.Error(err)
		return
	}

	// Start /event
	startEventServer(ssePubSub, t, 8081)

	// Start the client
	connected := make(chan bool)
	done := make(chan bool)
	data := []eventData{}
	httpToEvent(t, client, 8081, connected, done, &data)
	<-connected

	testData := "testdata"

	if client.GetStatus() != Receving {
		t.Errorf("client.GetStatus() != Receving")
		return
	}

	// Send data to client
	if err := topic.Pub(testData); err != nil {
		t.Error(err)
		return
	}

	// Wait for the client to receive the data
	<-done

	if len(data) < 1 {
		t.Errorf("len(data) < 1")
		return
	}
	for _, d := range data {
		if len(d.Updates) > 0 {
			if d.Updates[0].Data != testData {
				t.Errorf("data[].Updates[0].Data != testData")
				return
			}
			return
		}
	}
	t.Errorf("No Updates received")

}
