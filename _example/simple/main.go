package main

import (
	"github.com/apex/log"
	"net/http"
	"time"

	pubsubsse "github.com/bigbluebutton-bot/pubsub-sse"
)

func main() {
	// Create a new SSEPubSubService
	ssePubSub := pubsubsse.NewSSEPubSubService()

	// Handle endpoints
	http.Handle("/", http.FileServer(http.Dir("./web"))) // Serve static files

	// You can write your own endpoints if you want. Just have a look at the examples and modify them to your needs.
	http.HandleFunc("/add/user", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddClient(ssePubSub, w, r) })                 // Add client endpoint
	http.HandleFunc("/add/topic/public/", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddPublicTopic(ssePubSub, w, r) })   // Add topic endpoint
	http.HandleFunc("/add/topic/private/", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddPrivateTopic(ssePubSub, w, r) }) // Add topic endpoint
	http.HandleFunc("/sub", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Subscribe(ssePubSub, w, r) })                      // Subscribe endpoint
	http.HandleFunc("/unsub", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Unsubscribe(ssePubSub, w, r) })                  // Unsubscribe endpoint
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Event(ssePubSub, w, r) })                        // Event SSE endpoint
	go func() {
		err := http.ListenAndServe(":8080", nil)
		log.Fatalf("[sys]: %s", err.Error()) // Start http server
	}()

	// Create a new client and get it by id
	client := ssePubSub.NewClient()
	client, _ = ssePubSub.GetClientByID(client.GetID())
	log.Infof("[sys]: Client ID: %s", client.GetID())

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Get topic by ID. 3 ways to get a public topic:
	pubTopic, _ = ssePubSub.GetPublicTopicByID(pubTopicID)
	pubTopic, _ = client.GetTopicByID(pubTopicID)
	pubTopic, _ = client.GetPublicTopicByID(pubTopicID)

	// Subscribe to the topic
	client.Sub(pubTopic)

	// Send data to the topic
	pubTopic.Pub(TestData{Testdata: "testdata"})

	// Unsubscribe from topic
	client.Unsub(pubTopic)

	// Remove public topic
	ssePubSub.RemovePublicTopic(pubTopic)

	// Create a private topic
	privTopic := client.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Get topic by ID. 2 ways to get a private topic:
	privTopic, _ = client.GetTopicByID(privTopicID)
	privTopic, _ = client.GetPrivateTopicByID(privTopicID)

	// Subscribe to the topic
	client.Sub(privTopic)

	// Send data to the topic
	privTopic.Pub(TestData{Testdata: "testdata"})

	// Unsubscribe from topic
	client.Unsub(privTopic)

	// Remove private topic
	client.RemovePrivateTopic(privTopic)

	// Create a group
	group := ssePubSub.NewGroup()
	groupID := group.GetID()

	// Get group by ID
	group, _ = ssePubSub.GetGroupByID(groupID)

	// Add client to group
	group.AddClient(client)

	// Get group from client
	group, _ = client.GetGroupByID(groupID)

	// Create a group topic
	groupTopic := group.NewTopic()
	groupTopicID := groupTopic.GetID()

	// Get topic by ID. 2 ways to get a group topic:
	groupTopic, _ = group.GetTopicByID(groupTopicID)
	groupTopic, _ = client.GetTopicByID(groupTopicID)

	// Subscribe to the topic
	client.Sub(groupTopic)

	// Send data to the topic
	groupTopic.Pub(TestData{Testdata: "testdata"})

	// Unsubscribe from topic
	client.Unsub(groupTopic)

	// Remove group topic
	group.RemoveTopic(groupTopic)

	// Remove client from group
	group.RemoveClient(client)

	// Remove group
	ssePubSub.RemoveGroup(group)

	// Remove client
	ssePubSub.RemoveClient(client)

	// Create Public topic PUBLIC
	pubTopic = ssePubSub.NewPublicTopic()

	// If client is created
	ssePubSub.OnNewClient.Listen(func(c *pubsubsse.Client) {
		log.Infof("[sys]: New client: %s", c.GetID())

		// Subscribe to public topic
		c.Sub(pubTopic)
	})

	// Send data to PUBLIC every 5s
	go func() {
		for {
			pubTopic.Pub("DATAAAAAA")
			time.Sleep(5 * time.Second)
		}
	}()

	time.Sleep(500 * time.Second)
}

type TestData struct {
	Testdata string `json:"testdata"`
}
