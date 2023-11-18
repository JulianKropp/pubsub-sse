package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"github.com/bigbluebutton-bot/pubsub-sse"
)

func main() {
	// Create a new SSEPubSubService
	ssePubSub := pubsubsse.NewSSEPubSubService()

	// Handle endpoints
	// You can write your own endpoints if you want. Just have a look at the examples and modify them to your needs.
	http.HandleFunc("/add/user", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddClient(ssePubSub, w, r) })                 // Add client endpoint
	http.HandleFunc("/add/topic/public/", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddPublicTopic(ssePubSub, w, r) })   // Add topic endpoint
	http.HandleFunc("/add/topic/private/", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddPrivateTopic(ssePubSub, w, r) }) // Add topic endpoint
	http.HandleFunc("/sub", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Subscribe(ssePubSub, w, r) })                      // Subscribe endpoint
	http.HandleFunc("/unsub", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Unsubscribe(ssePubSub, w, r) })                  // Unsubscribe endpoint
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Event(ssePubSub, w, r) })                        // Event SSE endpoint
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil)) // Start http server
	}()

	// Create a new client and get it by id
	client := ssePubSub.NewClient()
	client, _ = ssePubSub.GetClientByID(client.GetID())
	fmt.Println("Client ID:", client.GetID())

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic("server/status")

	// Get topic by name. 3 ways to get a public topic:
	pubTopic, _ = ssePubSub.GetPublicTopicByName("server/status")
	pubTopic, _ = client.GetTopicByName("server/status")
	pubTopic, _ = client.GetPublicTopicByName("server/status")

	// Subscribe to the topic
	client.Sub(pubTopic)

	// Send data to the topic
	pubTopic.Pub(TestData{Testdata: "testdata"})

	// Unsubscribe from topic
	client.Unsub(pubTopic)

	// Remove public topic
	ssePubSub.RemovePublicTopic(pubTopic)

	// Create a private topic
	privTopic := client.NewPrivateTopic("test/server")

	// Get topic by name. 2 ways to get a private topic:
	privTopic, _ = client.GetTopicByName("test/server")
	privTopic, _ = client.GetPrivateTopicByName("test/server")

	// Subscribe to the topic
	client.Sub(privTopic)

	// Send data to the topic
	privTopic.Pub(TestData{Testdata: "testdata"})

	// Unsubscribe from topic
	client.Unsub(privTopic)

	// Remove private topic
	client.RemovePrivateTopic(privTopic)

	// // Create a group
	// group := ssePubSub.NewGroup("testgroup")
	// group, _ = ssePubSub.GetGroupByName("testgroup")

	// // Add client to group
	// group.AddClient(client)

	// // Get group from client
	// group, _ = client.GetGroupByName("testgroup")

	// // Create a group topic
	// groupTopic := group.NewTopic("test/group")

	// // Get topic by name. 3 ways to get a group topic:
	// groupTopic, _ = group.GetTopicByName("test/group")
	// groupTopic, _ = client.GetTopicByName("test/group")
	// groupTopic, _ = client.GetGroupTopicByName("test/group")

	// // Subscribe to the topic
	// client.Sub(groupTopic)

	// // Send data to the topic
	// groupTopic.Pub(TestData{Testdata: "testdata"})

	// // Unsubscribe from topic
	// client.Unsub(groupTopic)

	// // Remove group topic
	// group.RemoveTopic(groupTopic)

	// // Remove client from group
	// group.RemoveClient(client)

	// // Remove group
	// client.RemoveGroup(group)

	// Remove client
	ssePubSub.RemoveClient(client)

	time.Sleep(500 * time.Second)
}

type TestData struct {
	Testdata string `json:"testdata"`
}
