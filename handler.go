package pubsubsse

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apex/log"
)

// AddClient handles HTTP requests for adding a new client.
func AddClient(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Create a new client
	c := s.NewClient()

	// Send the client ID

	json.NewEncoder(w).Encode(map[string]string{"ok": "true", "client_id": c.GetID()})
}

// AddPublicTopic handles HTTP requests for adding a new public topic.
func AddPublicTopic(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET clientID and topic from request body
	topic := r.URL.Query().Get("topic")

	// Create a new public topic
	t := s.NewPublicTopic(topic)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true", "topic_name": t.GetName()})
}

// AddPrivateTopic handles HTTP requests for adding a new private topic.
func AddPrivateTopic(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET clientID and topic from request body
	clientID := r.URL.Query().Get("client_id")
	topic := r.URL.Query().Get("topic")

	// Get the client
	client, ok := s.GetClientByID(clientID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "client not found"})
		return
	}

	// Create a new private topic
	t := client.NewPrivateTopic(topic)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true", "topic_name": t.GetName()})
}

// Subscribe handles HTTP requests for client subscriptions.
func Subscribe(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET clientID and topic from request body
	clientID := r.URL.Query().Get("client_id")
	topic := r.URL.Query().Get("topic")

	// Get the client
	client, ok := s.GetClientByID(clientID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "client not found"})
		return
	}

	// Get the topic
	t, ok := client.GetTopicByName(topic)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "topic not found"})
		return
	}

	// Subscribe to the topic
	if err := client.Sub(t); err != nil {
		log.Errorf("Error subscribing to topic %s: %s", topic, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

// Unsubscribe handles HTTP requests for client unsubscriptions.
func Unsubscribe(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET clientID and topic from request body
	clientID := r.URL.Query().Get("client_id")
	topic := r.URL.Query().Get("topic")

	// Get the client
	client, ok := s.GetClientByID(clientID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "client not found"})
		return
	}

	// Get the topic
	t, ok := client.GetTopicByName(topic)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "topic not found"})
		return
	}

	// Unsubscribe from the topic
	if err := client.Unsub(t); err != nil {
		log.Errorf("Error unsubscribing from topic %s: %s", topic, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

// Event
func Event(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET clientID and topic from request body
	clientID := r.URL.Query().Get("client_id")

	// Get the client
	client, ok := s.GetClientByID(clientID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "client not found"})
		return
	}

	// Test if client is already receiving
	if client.GetStatus() == Receving {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "client is already receiving"})
		return
	}

	// SSE-specific headers
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Get the request's context. If the connection closes, the context will be canceled.
	ctx := r.Context()

	// Keep the connection open until it's closed by the client or client is removed
	// OnEvent: Send message to client if new data is published
	client.Start(ctx, func(msg string) {
		fmt.Fprintf(w, "%s", msg)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	})
}
