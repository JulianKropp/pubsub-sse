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
	http.HandleFunc("/add/user", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddInstance(ssePubSub, w, r) })                 // Add instance endpoint
	http.HandleFunc("/add/topic/public/", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddPublicTopic(ssePubSub, w, r) })   // Add topic endpoint
	http.HandleFunc("/add/topic/private/", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddPrivateTopic(ssePubSub, w, r) }) // Add topic endpoint
	http.HandleFunc("/sub", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Subscribe(ssePubSub, w, r) })                      // Subscribe endpoint
	http.HandleFunc("/unsub", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Unsubscribe(ssePubSub, w, r) })                  // Unsubscribe endpoint
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Event(ssePubSub, w, r) })                        // Event SSE endpoint
	go func() {
		err := http.ListenAndServe(":8080", nil)
		log.Fatalf("[sys]: %s", err.Error()) // Start http server
	}()

	// Create a new instance and get it by id
	instance := ssePubSub.NewInstance()
	instance, _ = ssePubSub.GetInstanceByID(instance.GetID())
	log.Infof("[sys]: Instance ID: %s", instance.GetID())

	// Create a public topic
	pubTopic := ssePubSub.NewPublicTopic()
	pubTopicID := pubTopic.GetID()

	// Get topic by ID. 3 ways to get a public topic:
	pubTopic, _ = ssePubSub.GetPublicTopicByID(pubTopicID)
	pubTopic, _ = instance.GetTopicByID(pubTopicID)
	pubTopic, _ = instance.GetPublicTopicByID(pubTopicID)

	// Subscribe to the topic
	instance.Sub(pubTopic)

	// Send data to the topic
	pubTopic.Pub(TestData{Testdata: "testdata"})

	// Unsubscribe from topic
	instance.Unsub(pubTopic)

	// Remove public topic
	ssePubSub.RemovePublicTopic(pubTopic)

	// Create a private topic
	privTopic := instance.NewPrivateTopic()
	privTopicID := privTopic.GetID()

	// Get topic by ID. 2 ways to get a private topic:
	privTopic, _ = instance.GetTopicByID(privTopicID)
	privTopic, _ = instance.GetPrivateTopicByID(privTopicID)

	// Subscribe to the topic
	instance.Sub(privTopic)

	// Send data to the topic
	privTopic.Pub(TestData{Testdata: "testdata"})

	// Unsubscribe from topic
	instance.Unsub(privTopic)

	// Remove private topic
	instance.RemovePrivateTopic(privTopic)

	// Create a group
	group := ssePubSub.NewGroup()
	groupID := group.GetID()

	// Get group by ID
	group, _ = ssePubSub.GetGroupByID(groupID)

	// Add instance to group
	group.AddInstance(instance)

	// Get group from instance
	group, _ = instance.GetGroupByID(groupID)

	// Create a group topic
	groupTopic := group.NewTopic()
	groupTopicID := groupTopic.GetID()

	// Get topic by ID. 2 ways to get a group topic:
	groupTopic, _ = group.GetTopicByID(groupTopicID)
	groupTopic, _ = instance.GetTopicByID(groupTopicID)

	// Subscribe to the topic
	instance.Sub(groupTopic)

	// Send data to the topic
	groupTopic.Pub(TestData{Testdata: "testdata"})

	// Unsubscribe from topic
	instance.Unsub(groupTopic)

	// Remove group topic
	group.RemoveTopic(groupTopic)

	// Remove instance from group
	group.RemoveInstance(instance)

	// Remove group
	ssePubSub.RemoveGroup(group)

	// Remove instance
	ssePubSub.RemoveInstance(instance)

	// Create Public topic PUBLIC
	pubTopic = ssePubSub.NewPublicTopic()

	// If instance is created
	ssePubSub.OnNewInstance.Listen(func(c *pubsubsse.Instance) {
		log.Infof("[sys]: New instance: %s", c.GetID())

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
