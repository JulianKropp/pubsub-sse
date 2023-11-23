package pubsubsse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"
)

// Tests for:
// +GetID(): string
// +GetStatus(): status

// +GetPublicTopics(): map[string]*topic
// +GetPublicTopicByName(name string): *topic, bool

// +NewPrivateTopic(name string): *topic
// +RemovePrivateTopic(t *topic)
// +GetPrivateTopics(): map[string]*topic
// +GetPrivateTopicByName(name string): *topic, bool

// +GetGroups(): map[string]*group
// +GetGroupByName(name string): *group, bool

// +GetAllTopics(): map[string]*topic
// +GetTopicByName(name string): *topic, bool
// +GetSubscribedTopics(): map[string]*topic

// +Sub(topic *topic): error
// +Unsub(topic *topic): error

// +OnEvent(f OnEventFunc)
// +RemoveOnEvent()
// +Start(ctx Context)

// -send(msg interface): error
// -sendTopicList(): error
// -sendSubscribedTopic(topic *topic): error
// -sendUnsubscribedTopic(topic *topic): error
// -sendInitMSG(): error
// -addGroup(g *group)
// -removeGroup(g *group)
// -stop()

// -----------------------------
// General
// -----------------------------

// Start /event server
func startEventServer(ssePubSub *sSEPubSubService, t *testing.T, port int) {
	srv := http.NewServeMux()
	srv.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) { Event(ssePubSub, w, r) }) // Event SSE endpoint
	go func() {
		errx := http.ListenAndServe(":"+strconv.Itoa(port), srv)
		if errx != nil {
			t.Error("http Listen err", errx)
		}
	}()
	time.Sleep(100 * time.Millisecond)
}

// Makes an http request to localhost:8080/event
// This will be an SSE connection and will be open for 10s
func httpToEvent(t *testing.T, client *client, port int, connected, done chan bool, returnValue *[]eventData) {
	hclient := http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:"+strconv.Itoa(port)+"/event?client_id="+client.GetID(), nil)
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := hclient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	// Set a timeout for the SSE connection
	timeout := time.After(3 * time.Second)
	stream := make(chan string)

	con := false

	// Goroutine to read from the SSE stream
	go func() {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				close(stream)
				t.Logf("Error reading from SSE stream: %s", err.Error())
				return
			}
			stream <- string(line)
		}
	}()

	// Listen for messages and timeout
	go func() {
		for {
			select {
			case <-timeout:
				fmt.Println("timeout")
				// Send connected signal
				if !con {
					con = true
					connected <- true
				}
				done <- true
				resp.Body.Close()
				return
			case message, ok := <-stream:
				if !ok {
					// Stream closed, exit loop
					return
				}

				fmt.Println("message: ", message)

				var rvalue eventData
				// Unmarshal the JSON data
				err = json.Unmarshal([]byte(message), &rvalue)
				if err != nil {
					t.Error(err)
					continue
				}

				// Send connected signal
				if !con {
					con = true
					connected <- true
				}

				*returnValue = append(*returnValue, rvalue)
			}
		}
	}()
}

// -----------------------------
//ID/Status
// -----------------------------

// TestClient_GetStatus tests Client.GetStatus()
func TestClient_GetID(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	c := ssePubSub.NewClient()
	if c.GetID() == "" {
		t.Error("Client.GetID() == \"\"")
	}
}

// TestClient_GetStatus tests Client.GetStatus()
func TestClient_GetStatus(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()
	if client.GetStatus() != Waiting {
		t.Error("Client.GetStatus() != Waiting")
	}

	// Start /event
	startEventServer(ssePubSub, t, 8080)

	// Start the client
	done := make(chan bool)
	connected := make(chan bool)
	httpToEvent(t, client, 8080, connected, done, &[]eventData{})
	<-connected

	if client.GetStatus() != Receving {
		t.Error("Client.GetStatus() != Receving")
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
		t.Error("len(client.GetPublicTopics()) != 0")
	}

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic("test")

	// Get topic
	pubTopics := client.GetPublicTopics()
	if len(pubTopics) != 1 {
		t.Error("len(pubTopics) != 1")
	}
	if pubTopics["test"] != pubTopic {
		t.Error("pubTopics[\"test\"] != pubTopic")
	}
}

// TestClient_GetPublicTopicByName tests Client.GetPublicTopicByName()
func TestClient_GetPublicTopicByName(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic("test")

	// Get topic
	topic, ok := client.GetPublicTopicByName("test")
	if !ok {
		t.Error("!ok")
	}
	if topic != pubTopic {
		t.Error("topic != pubTopic")
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
	privTopic := client.NewPrivateTopic("test")

	// Get topic
	privTopics := client.GetPrivateTopics()
	if len(privTopics) != 1 {
		t.Error("len(privTopics) != 1")
	}
	if privTopics["test"] != privTopic {
		t.Error("privTopics[\"test\"] != privTopic")
	}
}

// TestClient_RemovePrivateTopic tests Client.RemovePrivateTopic()
func TestClient_RemovePrivateTopic(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a private topic
	privTopic := client.NewPrivateTopic("test")

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
		t.Error("len(client.GetPrivateTopics()) != 0")
	}

	// Create a private topic
	privTopic := client.NewPrivateTopic("test")

	// Get topic
	privTopics := client.GetPrivateTopics()
	if len(privTopics) != 1 {
		t.Error("len(privTopics) != 1")
	}
	if privTopics["test"] != privTopic {
		t.Error("privTopics[\"test\"] != privTopic")
	}
}

// TestClient_GetPrivateTopicByName tests Client.GetPrivateTopicByName()
func TestClient_GetPrivateTopicByName(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a private topic
	privTopic := client.NewPrivateTopic("test")

	// Get topic
	topic, ok := client.GetPrivateTopicByName("test")
	if !ok {
		t.Error("!ok")
	}
	if topic != privTopic {
		t.Error("topic != privTopic")
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
		t.Error("len(client.GetGroups()) != 0")
	}

	// Create a group
	group := ssePubSub.NewGroup("test")

	// Add client to group
	group.AddClient(client)

	// Get group
	groups := client.GetGroups()
	if len(groups) != 1 {
		t.Error("len(groups) != 1")
	}
	if groups["test"] != group {
		t.Error("groups[\"test\"] != group")
	}
}

// TestClient_GetGroupByName tests Client.GetGroupByName()
func TestClient_GetGroupByName(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a group
	group := ssePubSub.NewGroup("test")

	// Add client to group
	group.AddClient(client)

	// Get group
	g, ok := client.GetGroupByName("test")
	if !ok {
		t.Error("!ok")
	}
	if g != group {
		t.Error("g != group")
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
		t.Error("len(client.GetAllTopics()) != 0")
	}

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic("test1")

	// Create a private topic
	privTopic := client.NewPrivateTopic("test2")

	// Create group
	group := ssePubSub.NewGroup("testGroup")
	group.AddClient(client)
	groupTopic := group.NewTopic("test3")

	// Get topic
	topics := client.GetAllTopics()
	if len(topics) != 3 {
		t.Error("len(topics) != 3")
	}
	if topics["test1"] != pubTopic {
		t.Error("topics[\"test\"] != pubTopic")
	}
	if topics["test2"] != privTopic {
		t.Error("topics[\"test\"] != privTopic")
	}
	if topics["test3"] != groupTopic {
		t.Error("topics[\"test\"] != groupTopic")
	}
}

// TestClient_GetTopicByName tests Client.GetTopicByName()
func TestClient_GetTopicByName(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic("test1")

	// Create a private topic
	privTopic := client.NewPrivateTopic("test2")

	// Create group
	group := ssePubSub.NewGroup("testGroup")
	group.AddClient(client)
	groupTopic := group.NewTopic("test3")

	// Get topic
	topic, ok := client.GetTopicByName("test1")
	if !ok {
		t.Error("!ok")
	}
	if topic != pubTopic {
		t.Error("topic != pubTopic")
	}

	topic, ok = client.GetTopicByName("test2")
	if !ok {
		t.Error("!ok")
	}
	if topic != privTopic {
		t.Error("topic != privTopic")
	}

	topic, ok = client.GetTopicByName("test3")
	if !ok {
		t.Error("!ok")
	}
	if topic != groupTopic {
		t.Error("topic != groupTopic")
	}
}

// TestClient_GetSubscribedTopics tests Client.GetSubscribedTopics()
func TestClient_GetSubscribedTopics(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()
	if len(client.GetSubscribedTopics()) != 0 {
		t.Error("len(client.GetSubscribedTopics()) != 0")
	}

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic("test1")

	// Create a private topic
	privTopic := client.NewPrivateTopic("test2")

	// Create group
	group := ssePubSub.NewGroup("testGroup")
	group.AddClient(client)
	groupTopic := group.NewTopic("test3")

	// Subscribe to topics
	client.Sub(pubTopic)
	client.Sub(privTopic)
	client.Sub(groupTopic)

	// Get topic
	topics := client.GetSubscribedTopics()
	if len(topics) != 3 {
		t.Error("len(topics) != 3")
	}
	if topics["test1"] != pubTopic {
		t.Error("topics[\"test\"] != pubTopic")
	}
	if topics["test2"] != privTopic {
		t.Error("topics[\"test\"] != privTopic")
	}
	if topics["test3"] != groupTopic {
		t.Error("topics[\"test\"] != groupTopic")
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
	pubTopic := ssePubSub.NewPublicTopic("test1")

	// Subscribe to topic
	client.Sub(pubTopic)

	// Get topic
	topics := client.GetSubscribedTopics()
	if len(topics) != 1 {
		t.Error("len(topics) != 1")
	}
	if topics["test1"] != pubTopic {
		t.Error("topics[\"test\"] != pubTopic")
	}
}

// TestClient_Unsub tests Client.Unsub()
func TestClient_Unsub(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic("test1")

	// Subscribe to topic
	client.Sub(pubTopic)

	// Unsubscribe from topic
	client.Unsub(pubTopic)

	// Get topic
	topics := client.GetSubscribedTopics()
	if len(topics) != 0 {
		t.Error("len(topics) != 0")
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
	topic := client.NewPrivateTopic("test")
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
		t.Error("client.GetStatus() != Receving")
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
		t.Error("len(data) < 1")
		return
	}
	for _, d := range data {
		if len(d.Updates) > 0 {
			if d.Updates[0].Data != testData {
				t.Error("data[].Updates[0].Data != testData")
				return
			}
			return
		}
	}
	t.Error("No Updates received")

}
