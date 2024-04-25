package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	pubsubsse "github.com/bigbluebutton-bot/pubsub-sse"
	"github.com/google/uuid"
)

type Server struct {
	lock       sync.Mutex
	id         string
	httpServer *http.Server
	httpMux    *http.ServeMux // This is the http handler
	httpport   int
	ssePubSub  *pubsubsse.SSEPubSubService
	chatRooms  map[string]*ChatRoom
}

func NewServer(port int) *Server {
	mux := http.NewServeMux() // Create a new ServeMux
	s := &Server{
		id:        "S-" + uuid.New().String(),
		httpport:  port,
		ssePubSub: pubsubsse.NewSSEPubSubService(),
		chatRooms: make(map[string]*ChatRoom),
		httpMux:   mux, // Assign the created ServeMux to httpMux
		httpServer: &http.Server{ // Initialize the http.Server with the ServeMux
			Addr:    ":" + strconv.Itoa(port),
			Handler: mux,
		},
	}
	return s
}

// Start server
func (s *Server) Start() {
	go func() {
		log.Printf("[sys]: Starting server with ID %s on port %d", s.id, s.httpport)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("[sys]: %s", err.Error()) // Only fatal if error is not due to a server closed
		}
	}()
}

// Stop server
func (s *Server) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Printf("[sys]: Stopping server with ID %s", s.id)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("[sys]: Error stopping server: %s", err)
	}
}

// Restart server
func (s *Server) Restart() {
	log.Printf("[sys]: Restarting server with ID %s", s.id)
	s.Stop()
	// Ensure the server has some time to release resources, if needed
	time.Sleep(1 * time.Second)
	s.Start()
}

func (s *Server) NewChatRoom() *ChatRoom {
	s.lock.Lock()
	defer s.lock.Unlock()

	chatRoom := &ChatRoom{
		lock:      sync.Mutex{},
		id:        "CR-" + uuid.New().String(),
		server:    s,
		users:     make(map[string]*User),
		messages:  make([]*ChatMessage, 0),
		userTopic: s.ssePubSub.NewPublicTopic(),
		chatTopic: s.ssePubSub.NewPublicTopic(),
	}

	// Add chatRoom to ChatRooms
	s.chatRooms[chatRoom.id] = chatRoom

	return chatRoom
}

// Get id
func (s *Server) GetID() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.id
}

func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.httpMux.HandleFunc(pattern, handler) // Use httpMux to add the handler
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.httpMux.Handle(pattern, handler) // Use httpMux to add the handler
}

// Get port
func (s *Server) GetPort() int {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.httpport
}

// Set port and restart server
func (s *Server) SetPort(port int) {
	s.lock.Lock()
	s.httpport = port
	s.lock.Unlock()

	s.Restart()
}

// GetChatRoomByID
func (s *Server) GetChatRoomByID(id string) (*ChatRoom, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	chatRoom, ok := s.chatRooms[id]
	if !ok {
		return nil, false
	}
	return chatRoom, true
}

//Get client by id
func (s *Server) GetClientByID(id string) (*pubsubsse.Client, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	client, ok := s.ssePubSub.GetClientByID(id)
	if !ok {
		return nil, false
	}
	return client, true
}