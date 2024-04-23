package main

import (
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

	// If client is created sub to users and publicChat
	Server.ssePubSub.OnNewClient.Listen(func(c *pubsubsse.Client) {
		log.Infof("[sys]: New client: %s", c.GetID())

		// Create new client
		user := PublicChatRoom.NewUser("User", c)

		// If client disconnects remove client after 10s and delete the listener
		var onStatusChangeID string
		onStatusChangeID = c.OnStatusChange.Listen(func(status pubsubsse.Status) {
			log.Infof("[sys]: Client status change: %s", status)

			if status == pubsubsse.Waiting { // If client is waiting for too long remove client
				time.Sleep(10 * time.Second)

				if c.GetStatus() == pubsubsse.Waiting {
					log.Infof("[sys]: Client Timed out: %s", c.GetID())
					c.OnStatusChange.Remove(onStatusChangeID)
					Server.ssePubSub.RemoveClient(c)

					// Remove user
					PublicChatRoom.RemoveUser(user)
				}
			}
		})
	})

	time.Sleep(500 * time.Second)
}




type Server struct {
	lock sync.Mutex

	id        string
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

	// Handle endpoints
	http.Handle("/", http.FileServer(http.Dir("./web"))) // Serve static files

	// You can write your own endpoints if you want. Just have a look at the examples and modify them to your needs.
	http.HandleFunc("/add/user", func(w http.ResponseWriter, r *http.Request) { pubsubsse.AddClient(ssePubSub, w, r) }) // Add client endpoint
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) { pubsubsse.Event(ssePubSub, w, r) })        // Event SSE endpoint
	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
		log.Fatalf("[sys]: %s", err.Error()) // Start http server
	}()

	return &Server{
		lock:      sync.Mutex{},
		id:        "S-" + uuid.New().String(),
		ssePubSub: ssePubSub,
		chatRooms: make(map[string]*ChatRoom),
	}
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

func (c *ChatRoom) NewUser(name string, client *pubsubsse.Client) *User {
	user := &User{
		lock:      sync.Mutex{},
		Id:   "U-" + uuid.New().String(),
		Name: name,
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
