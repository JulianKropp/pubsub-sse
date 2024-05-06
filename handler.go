package pubsubsse

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apex/log"
)

// AddInstance handles HTTP requests for adding a new instance.
func AddInstance(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get connection_id
	var i *Instance
	connectionID := r.URL.Query().Get("connection_id")
	if connectionID == "" {
		i = s.NewInstance()
	} else {
		// Create a new instance
		i = s.NewInstance(connectionID)
	}

	// Send the instance ID
	json.NewEncoder(w).Encode(map[string]string{"ok": "true", "instance_id": i.GetID(), "connection_id": i.GetConnectionID()})
}

// Update handles HTTP requests for updating the connection_id of an instance.
func UpdateInstance(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET instanceID and connectionID from request body
	instanceID := r.URL.Query().Get("instance_id")
	connectionID := r.URL.Query().Get("connection_id")

	// Get the instance
	instance, ok := s.GetInstanceByID(instanceID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "instance not found"})
		return
	}

	// Update the connection_id
	if err := instance.ChangeConnection(connectionID); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "connection_id not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

// Remove handles HTTP requests for removing an instance.
func RemoveInstance(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET instanceID from request body
	instanceID := r.URL.Query().Get("instance_id")

	// Get the instance
	instance, ok := s.GetInstanceByID(instanceID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "instance not found"})
		return
	}

	// Remove the instance
	s.RemoveInstance(instance)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

// AddPublicTopic handles HTTP requests for adding a new public topic.
func AddPublicTopic(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Create a new public topic
	t := s.NewPublicTopic()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true", "topic_id": t.GetID()})
}

// AddPrivateTopic handles HTTP requests for adding a new private topic.
func AddPrivateTopic(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET instanceID and topic from request body
	instanceID := r.URL.Query().Get("instance_id")

	// Get the instance
	instance, ok := s.GetInstanceByID(instanceID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "instance not found"})
		return
	}

	// Create a new private topic
	t := instance.NewPrivateTopic()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true", "topic_id": t.GetID()})
}

// Subscribe handles HTTP requests for instance subscriptions.
func Subscribe(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET instanceID and topic from request body
	instanceID := r.URL.Query().Get("instance_id")
	topic := r.URL.Query().Get("topic")

	// Get the instance
	instance, ok := s.GetInstanceByID(instanceID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "instance not found"})
		return
	}

	// Get the topic
	t, ok := instance.GetTopicByID(topic)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "topic not found"})
		return
	}

	// Subscribe to the topic
	if err := instance.Sub(t); err != nil {
		log.Errorf("Error subscribing to topic %s: %s", topic, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
}

// Unsubscribe handles HTTP requests for instance unsubscriptions.
func Unsubscribe(s *SSEPubSubService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET instanceID and topic from request body
	instanceID := r.URL.Query().Get("instance_id")
	topic := r.URL.Query().Get("topic")

	// Get the instance
	instance, ok := s.GetInstanceByID(instanceID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "instance not found"})
		return
	}

	// Get the topic
	t, ok := instance.GetTopicByID(topic)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "topic not found"})
		return
	}

	// Unsubscribe from the topic
	if err := instance.Unsub(t); err != nil {
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

	// GET instanceID and topic from request body
	instanceID := r.URL.Query().Get("instance_id")

	// Get the instance
	instance, ok := s.GetInstanceByID(instanceID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "instance not found"})
		return
	}

	// Test if instance is already receiving
	if instance.GetStatus() == Receiving {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "instance is already receiving"})
		return
	}

	// SSE-specific headers
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Get the request's context. If the connection closes, the context will be canceled.
	ctx := r.Context()

	// Keep the connection open until it's closed by the instance or instance is removed
	// OnEvent: Send message to instance if new data is published
	instance.Start(ctx, func(msg string) {
		fmt.Fprintf(w, "%s", msg)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	})
}
