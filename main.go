package main

import (
	"log"
	"net/http"
	"time"
)


// Explanation of the code:
// ssePubSub is Handler which send live data to the client using SSE.
// The client receves a json string over SSE which looks like this:
// {
// 	"sys": [
// 		{
// 			"topics": [
// 				{"name": "topic1", "type": "PUBLIC"},
// 				{"name": "topic2/test", "type": "PRIVATE"},
// 				{"name": "topic3/test/test", "type": "PRIVATE"},
// 				{"name": "topic4/status", "type": "GROUP"},
// 				{"name": "topic5", "type": "PUBLIC"},
// 				{"name": "topic10", "type": "PUBLIC"},
// 				{"name": "topic11/test/test", "type": "PRIVATE"}
// 			]
// 		},
// 		{
// 			"subscribed": ["topic1", "topic2/test", "topic3/test/test", "topic4/status"]
// 		},
// 		{
// 			"unsubsribed": ["topic5", "topic10", "topic11/test/test"]
// 		}
// 	],
//    "updates":[
//       {
//          "topic1":{
//          }
//       },
//       {
//          "topic2/test":{
//          }
//       }
//    ]
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
	ssePubSub := NewSSEPubSubHandler()

	http.HandleFunc("/add", ssePubSub.AddClient)      // Add client endpoint
	http.HandleFunc("/sub", ssePubSub.Subscribe)      // Subscribe endpoint
	http.HandleFunc("/unsusb", ssePubSub.Unsubscribe) // Unsubscribe endpoint
	http.HandleFunc("/event", ssePubSub.Event)        // SSE endpoint
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Create a new client
	clientID := "client1"
	ssePubSub.NewClient(clientID)


	err := ssePubSub.NewPublicTopic("server/status") // Create a new topic which is public to all clients
	if err != nil {
		log.Println(err)
	}

	for {
		time.Sleep(5 * time.Second)

		// Publish a message to all clients
		data := TestData{
			Testdata: "testdata",
		}
		err = ssePubSub.Pub("server/status", data)
		if err != nil {
			log.Println(err)
		}
	}

	// lastclient := ""

	// clients := ssePubSub.GetClients()
	// for _, client := range clients {
	// 	log.Println(client)
	// 	// Publish a message to a specific client
	// 	topic := "private/test/topic"
	// 	data := TestData{
	// 		Testdata: client.id,
	// 	}
	// 	client.Pub(topic, data)

	// 	lastclient = client.id
	// }

	// client := clients[lastclient]
	// if client == nil {
	// 	panic("client is nil")
	// }

	// err = ssePubSub.NewPrivateTopic("private/private/topic", client) // Create a new topic which is private Private to one clients
	// if err != nil {
	// 	log.Println(err)
	// }

	// // List all public, private and private private topics of a client
	// topics := client.GetTopics()
	// for _, topic := range topics {
	// 	log.Println(topic)
	// }

	// // Subscribe a client to a topic
	// if err := client.Sub("private/private/topic"); err != nil { // Subscribe to topic private and to all sub topics like private/test/topic and private/private/topic, ...
	// 	log.Println(err)
	// }

	// // Get a list of all topics a client is subscribed to
	// topics = client.GetSubscribedTopics()
	// for _, topic := range topics {
	// 	log.Println(topic)
	// }

	// // Unsubscribe a client from a topic
	// if err := client.Unsub("private/private/topic"); err != nil {
	// 	log.Println(err)
	// }

	// // Remove a client from the server
	// if err := ssePubSub.RemoveClient(client.id); err != nil {
	// 	log.Println(err)
	// }

	// // List all public topics
	// topics = ssePubSub.GetTopics()
	// for _, topic := range topics {
	// 	log.Println(topic)
	// }

	// // Remove a public or private topic from the server and unsubscribe all clients from it
	// if err := ssePubSub.RemovePublicTopic("server/status"); err != nil {
	// 	log.Println(err)
	// }

	// // List all public topics
	// topics = ssePubSub.GetTopics()
	// for _, topic := range topics {
	// 	log.Println(topic)
	// }
}
