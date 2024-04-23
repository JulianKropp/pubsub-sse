package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/google/uuid"

	pubsubsse "github.com/bigbluebutton-bot/pubsub-sse"
)

func main() {
	// Create a new Server and a new ChatRoom
	Server := NewServer(8080)
	PublicChatRoom := Server.NewChatRoom()

	// Handle special endpoints
	Server.http.HandleFunc("/add/user", func(w http.ResponseWriter, r *http.Request) { AddClient(PublicChatRoom, w, r) }) // Add client endpoint
	Server.Start()

	time.Sleep(500 * time.Second)
}

func AddClient(c *ChatRoom, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Create a new user
	user := c.NewUser("User")

	// Send the client ID
	json.NewEncoder(w).Encode(map[string]string{"ok": "true", "client_id": user.client.GetID()})
}

type Server struct {
	lock sync.Mutex

	id        string
	http      *http.ServeMux
	httpport  int
	ssePubSub *pubsubsse.SSEPubSubService
	chatRooms map[string]*ChatRoom
}

type ChatRoom struct {
	lock sync.Mutex `json:"-"`

	Id        string           `json:"id"`
	server    *Server          `json:"-"`
	Users     map[string]*User `json:"users"`
	Messages  []*ChatMessage   `json:"messages"`
	userTopic *pubsubsse.Topic `json:"-"`
	chatTopic *pubsubsse.Topic `json:"-"`
}

type User struct {
	lock sync.Mutex `json:"-"`

	Id     string            `json:"id"`
	Name   string            `json:"name"`
	client *pubsubsse.Client `json:"-"`
}

type ChatMessage struct {
	lock sync.Mutex `json:"-"`

	User    User      `json:"user"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func NewServer(port int) *Server {
	// Create a new SSEPubSubService
	ssePubSub := pubsubsse.NewSSEPubSubService()

	s := &Server{
		lock:      sync.Mutex{},
		id:        "S-" + uuid.New().String(),
		http:      http.NewServeMux(),
		httpport:  port,
		ssePubSub: ssePubSub,
		chatRooms: make(map[string]*ChatRoom),
	}

	// Handle endpoint
	s.http.Handle("/", http.FileServer(http.Dir("./web")))                                                           // Serve static files
	s.http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Event(s.ssePubSub, w, r) }) // Event SSE endpoint

	return s
}

func (s *Server) Start() {
	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(s.httpport), s.http)
		log.Fatalf("[sys]: %s", err.Error()) // Start http server
	}()
}

func (s *Server) NewChatRoom() *ChatRoom {
	s.lock.Lock()
	defer s.lock.Unlock()

	chatRoom := &ChatRoom{
		lock:      sync.Mutex{},
		Id:        "CR-" + uuid.New().String(),
		server:    s,
		Users:     make(map[string]*User),
		Messages:  make([]*ChatMessage, 0),
		userTopic: s.ssePubSub.NewPublicTopic(),
		chatTopic: s.ssePubSub.NewPublicTopic(),
	}

	// Add chatRoom to ChatRooms
	s.chatRooms[chatRoom.Id] = chatRoom

	return chatRoom
}

func (c *ChatRoom) NewUser(name string) *User {
	// Create a new client
	ssePubSub := c.server.ssePubSub
	client := ssePubSub.NewClient()

	user := &User{
		lock:   sync.Mutex{},
		Id:     "U-" + uuid.New().String(),
		Name:   name,
		client: client,
	}

	c.lock.Lock()
	c.Users[user.Id] = user
	c.lock.Unlock()

	// Send user to userTopic
	c.sendUserList()

	// Subscribe user to chatTopic and userTopic
	user.client.Sub(c.chatTopic)
	user.client.Sub(c.userTopic)

	// If client disconnects remove client after 10s and delete the listener
	var onStatusChangeID string
	onStatusChangeID = client.OnStatusChange.Listen(func(status pubsubsse.Status) {
		log.Infof("[sys]: Client status change: %s", status)

		if client.GetStatus() == pubsubsse.Timeout || client.GetStatus() == pubsubsse.Stopped {
			log.Infof("[sys]: Client Timed out: %s", client.GetID())
			client.OnStatusChange.Remove(onStatusChangeID)
			ssePubSub.RemoveClient(client)

			// Remove user
			c.RemoveUser(user)
		}
	})

	return user
}

func (c *ChatRoom) sendUserList() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Send all users to userTopic
	c.userTopic.Pub(c.Users)
}

func (c *ChatRoom) RemoveUser(user *User) {
	// Remove user from users
	c.lock.Lock()
	delete(c.Users, user.Id)
	c.lock.Unlock()

	// Send user to userTopic
	c.sendUserList()
}
