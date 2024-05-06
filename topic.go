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
	id        string
	ttype     topicType
	instances map[string]*Instance
	lock      sync.Mutex

	// Events:
	OnNewInstance      *eventManager[*Instance]
	OnNewSubOfInstance *eventManager[*Instance]
	OnPub              *eventManager[interface{}]
	OnRemoveInstance   *eventManager[*Instance]
	OnUnsubOfInstance  *eventManager[*Instance]
}

// Create a new topic
func newTopic(ttype topicType) *Topic {
	return &Topic{
		id:        "T-" + uuid.New().String(),
		ttype:     ttype,
		instances: make(map[string]*Instance),

		// Events:
		OnNewInstance:      newEventManager[*Instance](),
		OnNewSubOfInstance: newEventManager[*Instance](),
		OnPub:              newEventManager[interface{}](),
		OnRemoveInstance:   newEventManager[*Instance](),
		OnUnsubOfInstance:  newEventManager[*Instance](),
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

// Add a instance to the topic
func (t *Topic) addInstance(i *Instance) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.instances[i.id] = i

	// Events
	t.OnNewSubOfInstance.Emit(i)
}

// Remove a instance from the topic
func (t *Topic) removeInstance(i *Instance) {
	t.lock.Lock()
	defer t.lock.Unlock()

	delete(t.instances, i.id)

	// Events
	t.OnUnsubOfInstance.Emit(i)
}

// Get all instances in the topic
func (t *Topic) GetInstances() map[string]*Instance {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Create a copy of the map
	newmap := make(map[string]*Instance)
	for k, v := range t.instances {
		newmap[k] = v
	}
	return newmap
}

// Check if a instance is subscribed to the topic
func (t *Topic) IsSubscribed(i *Instance) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	_, ok := t.instances[i.id]
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

// Publish a message to all instances in the topic
func (t *Topic) Pub(msg interface{}) error {
	// Build the JSON data
	fulldata := &connectionData{
		InstanceData: []instanceData{
			{
				ID: "",
				Data: eventData{
					Updates: []eventDataUpdates{
						{
							Topic: t.GetID(),
							Data:  msg,
						},
					},
				},
			},
		},
	}

	// Send the JSON data to all instances
	for _, c := range t.GetInstances() {
		fulldata.InstanceData[0].ID = c.GetID()
		err := c.send(fulldata) // ignore error. Fire and forget.
		if err != nil {
			log.Warnf("[T:%s]: Warning sending data to connection: %s", t.GetID(), err.Error())
		}
	}

	// Events
	t.OnPub.Emit(fulldata)

	return nil
}
