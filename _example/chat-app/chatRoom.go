package main

import (
	"sync"
	"time"

	"github.com/apex/log"
	pubsubsse "github.com/bigbluebutton-bot/pubsub-sse"
	"github.com/google/uuid"
)

type ChatRoom struct {
	lock sync.Mutex

	id        string
	server    *Server
	users     map[string]*User
	messages  []*ChatMessage
	userTopic *pubsubsse.Topic
	chatTopic *pubsubsse.Topic
}

type ChatRoomJSON struct {
	Id       string           `json:"id"`
	Users    map[string]*User `json:"users"`
	Messages []*ChatMessage   `json:"messages"`
}

func (c *ChatRoom) NewUser(name string) *User {
	// Create a new client
	ssePubSub := c.server.ssePubSub
	client := ssePubSub.NewClient()

	user := &User{
		lock:     sync.Mutex{},
		id:       "U-" + uuid.New().String(),
		name:     name,
		client:   client,
		chatRoom: c,
	}

	c.lock.Lock()
	c.users[user.id] = user
	c.lock.Unlock()

	// Send user to userTopic
	c.sendUserList()

	// Subscribe user to chatTopic and userTopic
	user.client.Sub(c.chatTopic)
	user.client.Sub(c.userTopic)

	// If client disconnects remove user
	var onStatusChangeID string
	onStatusChangeID = client.OnStatusChange.Listen(func(status pubsubsse.Status) {
		log.Infof("[sys]: Client status change: %s", pubsubsse.StatusToString(status))

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

func (c *ChatRoom) RemoveUser(user *User) {
	// Remove user from users
	c.lock.Lock()
	delete(c.users, user.id)
	c.lock.Unlock()

	// Send user to userTopic
	c.sendUserList()
}

func (c *ChatRoom) sendUserList() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Send all users to userTopic
	c.userTopic.Pub(c.users)
}

// Get id
func (c *ChatRoom) GetID() string {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.id
}

// Get users
func (c *ChatRoom) GetUsers() map[string]*User {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.users
}

// Get messages
func (c *ChatRoom) GetMessages() []*ChatMessage {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.messages
}

// Add message
func (c *ChatRoom) AddMessage(user *User, message string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	chatMessage := &ChatMessage{
		user:    user,
		message: message,
		time:    time.Now(),
	}

	c.messages = append(c.messages, chatMessage)

	// Send message to chatTopic
	c.chatTopic.Pub(chatMessage)
}

// Clear chat
func (c *ChatRoom) ClearChat() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.messages = make([]*ChatMessage, 0)
}

// User to JSON
func (c *ChatRoom) ToJSON() *ChatRoomJSON {
	return &ChatRoomJSON{
		Id:       c.GetID(),
		Users:    c.GetUsers(),
		Messages: c.GetMessages(),
	}
}
