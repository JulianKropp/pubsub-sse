package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/apex/log"
	pubsubsse "github.com/bigbluebutton-bot/pubsub-sse"
)

func main() {
	// Create a new Server and a new ChatRoom
	Server := NewServer(8080)
	chatroom := Server.NewChatRoom()
	log.Infof("Chatroom ID: %s", chatroom.GetID())

	// Handle endpoint
	// handle favorite icon for all paths
	Server.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./web/favicon.ico") }) // Serve favicon.ico
	Server.Handle("/", http.FileServer(http.Dir("./web/root")))                                                                   // Serve static files
	Server.Handle("/chat", http.FileServer(http.Dir("./web/chat")))                                                               // Serve chat.html
	Server.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) { Event(Server, w, r) })                             // Event SSE endpoint
	Server.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) { Join(Server, w, r) })                               // Add client endpoint
	Server.Start()

	time.Sleep(500 * time.Second)
}

func Join(s *Server, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get username and chatroom_id from request
	username := r.URL.Query().Get("username")
	chatroom_id := r.URL.Query().Get("chatroom_id")

	if username == "" {
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "Invalid username"})
		return
	}
	chatroom, ok := s.GetChatRoomByID(chatroom_id)
	if ok == false {
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "Invalid chatroom_id"})
		return
	}

	// Create a new user
	user := chatroom.NewUser(username)

	// Send the client ID
	json.NewEncoder(w).Encode(map[string]string{"ok": "true", "connection_id": user.client.GetID(), "user_id": user.GetID(), "chatroom_id": chatroom.GetID()})
}

// Event
func Event(s *Server, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET connectionID and topic from request body
	connectionID := r.URL.Query().Get("connection_id")

	// Get the client
	client, ok := s.GetClientByID(connectionID)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "connection not found"})
		return
	}

	// Test if client is already receiving
	if client.GetStatus() == pubsubsse.Receiving {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{"ok": "false", "error": "connection is already receiving"})
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
