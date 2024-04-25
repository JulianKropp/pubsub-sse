package main

import (
	"sync"
	"time"
)

type ChatMessage struct {
	lock sync.Mutex

	user    *User
	message string
	time    time.Time
}

type ChatMessageJSON struct {
	User    *UserJSON `json:"user"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// New ChatMessage
func NewChatMessage(user *User, message string) *ChatMessage {
	return &ChatMessage{
		user:    user,
		message: message,
		time:    time.Now(),
	}
}

// Get user
func (c *ChatMessage) GetUser() *User {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.user
}

// Get message
func (c *ChatMessage) GetMessage() string {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.message
}

// Get time
func (c *ChatMessage) GetTime() time.Time {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.time
}
