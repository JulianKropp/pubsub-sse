package pubsubsse

import (
	"sync"

	"github.com/apex/log"
	"github.com/google/uuid"
)

// Topic Types
type topicType string

const (
	TPublic  topicType = "public"
	TPrivate topicType = "private"
	TGroup   topicType = "group"
)

// Topic represents a messaging Topic in the SSE pub-sub system.
type Topic struct {
	id      string
	ttype   topicType
	clients map[string]*Instance
	lock    sync.Mutex

	// Events:
	OnNewClient      *eventManager[*Instance]
	OnNewSubOfClient *eventManager[*Instance]
	OnPub            *eventManager[interface{}]
	OnRemoveClient   *eventManager[*Instance]
	OnUnsubOfClient  *eventManager[*Instance]
}

// Create a new topic
func newTopic(ttype topicType) *Topic {
	return &Topic{
		id:      "T-" + uuid.New().String(),
		ttype:   ttype,
		clients: make(map[string]*Instance),

		// Events:
		OnNewClient:      newEventManager[*Instance](),
		OnNewSubOfClient: newEventManager[*Instance](),
		OnPub:            newEventManager[interface{}](),
		OnRemoveClient:   newEventManager[*Instance](),
		OnUnsubOfClient:  newEventManager[*Instance](),
	}
}

// Get ID
func (t *Topic) GetID() string {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.id
}

// Get Type
func (t *Topic) GetType() string {
	t.lock.Lock()
	defer t.lock.Unlock()

	return string(t.ttype)
}

// Add a client to the topic
func (t *Topic) addClient(c *Instance) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.clients[c.id] = c

	// Events
	t.OnNewSubOfClient.Emit(c)
}

// Remove a client from the topic
func (t *Topic) removeClient(c *Instance) {
	t.lock.Lock()
	defer t.lock.Unlock()

	delete(t.clients, c.id)

	// Events
	t.OnUnsubOfClient.Emit(c)
}

// Get all clients in the topic
func (t *Topic) GetClients() map[string]*Instance {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Create a copy of the map
	newmap := make(map[string]*Instance)
	for k, v := range t.clients {
		newmap[k] = v
	}
	return newmap
}

// Check if a client is subscribed to the topic
func (t *Topic) IsSubscribed(c *Instance) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	_, ok := t.clients[c.id]
	return ok
}

type connectionData struct {
	InstanceData []instanceData `json:"instances"`
}

type instanceData struct {
	ID   string    `json:"id"`
	Data eventData `json:"data"`
}

type eventData struct {
	Sys     []eventDataSys     `json:"sys"`
	Updates []eventDataUpdates `json:"updates"`
}

type eventDataSys struct {
	Type string             `json:"type"`
	List []eventDataSysList `json:"list,omitempty"`
}

type eventDataSysList struct {
	ID   string `json:"id"`
	Type string `json:"type,omitempty"` // topics, subscribed, unsubscribed
}

type eventDataUpdates struct {
	Topic string      `json:"topic"`
	Data  interface{} `json:"data"`
}

// Publish a message to all clients in the topic
func (t *Topic) Pub(msg interface{}) error {
	// Build the JSON data
	fulldata := &eventData{
		Updates: []eventDataUpdates{},
	}
	u := eventDataUpdates{
		Topic: t.GetID(),
		Data:  msg,
	}
	fulldata.Updates = append(fulldata.Updates, u)

	// Send the JSON data to all clients
	for _, c := range t.GetClients() {
		err := c.send(fulldata) // ignore error. Fire and forget.
		if err != nil {
			log.Warnf("[T:%s]: Warning sending data to client: %s", t.GetID(), err.Error())
		}
	}

	// Events
	t.OnPub.Emit(fulldata)

	return nil
}
