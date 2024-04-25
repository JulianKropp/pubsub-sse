package main

import (
	"sync"

	pubsubsse "github.com/bigbluebutton-bot/pubsub-sse"
)

// User
type User struct {
	lock sync.Mutex

	id       string
	name     string
	client   *pubsubsse.Client
	chatRoom *ChatRoom
}

type UserJSON struct {
	Id           string `json:"user_id"`
	Name         string `json:"username"`
	ConnectionID string `json:"connection_id"`
	ClientID	 string `json:"client_id"`
	ChatRoomID   string `json:"chatroom_id"`
}

// Get id
func (u *User) GetID() string {
	u.lock.Lock()
	defer u.lock.Unlock()

	return u.id
}

// Get connection id
func (u *User) GetConnectionID() string {
	u.lock.Lock()
	defer u.lock.Unlock()

	return u.client.GetID()
}

// Get name
func (u *User) GetName() string {
	u.lock.Lock()
	defer u.lock.Unlock()

	return u.name
}

// Set name
func (u *User) SetName(name string) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.name = name

	// pub new user list
	u.chatRoom.sendUserList()
}

// Get chatroom
func (u *User) GetChatRoom() *ChatRoom {
	u.lock.Lock()
	defer u.lock.Unlock()

	return u.chatRoom
}

// Get user as JSON
func (u *User) GetJSON() *UserJSON {
	return &UserJSON{
		Id:           u.GetID(),
		ConnectionID: u.GetConnectionID(),
		Name:         u.GetName(),
		ChatRoomID:   u.GetChatRoom().GetID(),
	}
}
