package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Explanation of the code:
// ssePubSub is Handler which send live data to the client using SSE.
// The client receves a json string over SSE which looks like this:
// {
// 	"sys":[
// 	   {
// 		  "type":"topics",
// 		  "list":[
// 			 {
// 				"name":"server/status",
// 				"type":"public"
// 			 },
// 			 {
// 				"name":"test/server",
// 				"type":"private"
// 			 },
// 			 {
// 				"name":"test/server2",
// 				"type":"private"
// 			 }
// 		  ]
// 	   },
// 	   {
// 		  "type":"subscribed",
// 		  "list":[
// 			 {
// 				"name":"server/status"
// 			 },
// 			 {
// 				"name":"test/server"
// 			 }
// 		  ]
// 	   },
// 	   {
// 		  "type":"unsubscribed",
// 		  "list":[
// 			 {
// 				"name":"test5/server"
// 			 }
// 		  ]
// 	   }
// 	],
// 	"updates":[
// 	   {
// 		  "topic":"server/status",
// 		  "data":{
// 			 "testdata":"testdata"
// 		  }
// 	   }
// 	]
//  }

// Empty example:
// {
// 	"sys": null,
// 	"updates": null
// }

// Data update example:
// {
// 	"sys": null,
// 	"updates":[
// 	   {
// 		  "topic":"topic",
// 		  "data":{
//			 "testdata":"testdata"
// 		  }
// 	   }
// 	]
// }

// There are 3 types of topics:
// PUBLIC: All clients can subscribed to this topic. Data send in this topic will be send to all clients.
// PRIVATE: Only thhis client can subscribed to this topic. Data send in this topic will be send to only this client
// GROUP: All clients in the same group can subscribed to this topic. Data send in this topic will be send to all clients in the same group.

// Topics are case sensetive and are limited to the alphabet, numbers, underscores.
// Topics can have sub topics seperated by a slash.
// Subtopics can have subtopics too.
// Example: topic1/test/test
// If you subscribe to a topic, you will also subscribe to all subtopics.
// Example: If you subscribe to topic1, you will also subscribe to topic1/test and topic1/test/test, ...
// Example: If you subscribe to topic1/test, you will also subscribe to topic1/test/test, ...

// if somehow multiple topics with the same name of different types exists, the order of the types is:
// 1. PRIVATE
// 2. GROUP
// 3. PUBLIC
// This ensures PRIVATE and GROUP topics can not be overridden and made public exededently.

// sys.topics: Event: List of all public, group and private topics
// sys.subscribed: Event: List of all public, group and private topics the client is subscribed to
// sys.unsubscribed: Event: List of all public, group and private topics the client is no longer subscribed to
// updates: List of all public, group and private topics the client is subscribed to and which have been updated with new data

// Only changes will be send to the client:

// If a topic is added, the hole sys.list will be send to the client

// If a topic is removed, the hole new sys.list will be send to the client, and
// only this topic which was removed will be send in sys.unsubsribed

// If a client is subscribed to a topic, only the topic will be send to the client in sys.subscribed.

// If a client is unsubscribed from a topic, only the topic will be send to the client in sys.unsubscribed.

// If there is a new update for a topic, only this topic will be send to the client in updates.

type TestData struct {
	Testdata string `json:"testdata"`
}

func main() {
	// Create a new SSEPubSubService
	ssePubSub := NewSSEPubSubService()

	// Handle endpoints
	// You can write your own endpoints if you want. Just have a look at the examples and modify them to your needs.
	http.HandleFunc("/add/user", func(w http.ResponseWriter, r *http.Request) { AddClient(ssePubSub, w, r) })                 // Add client endpoint
	http.HandleFunc("/add/topic/public/", func(w http.ResponseWriter, r *http.Request) { AddPublicTopic(ssePubSub, w, r) })   // Add topic endpoint
	http.HandleFunc("/add/topic/private/", func(w http.ResponseWriter, r *http.Request) { AddPrivateTopic(ssePubSub, w, r) }) // Add topic endpoint
	http.HandleFunc("/sub", func(w http.ResponseWriter, r *http.Request) { Subscribe(ssePubSub, w, r) })                      // Subscribe endpoint
	http.HandleFunc("/unsub", func(w http.ResponseWriter, r *http.Request) { Unsubscribe(ssePubSub, w, r) })                  // Unsubscribe endpoint
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) { Event(ssePubSub, w, r) })                        // Event SSE endpoint
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
