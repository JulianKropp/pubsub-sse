package pubsubsse

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/apex/log"
	"github.com/google/uuid"
)

type status int

const (
	Waiting status = iota
	Receving
)

type OnEventFunc func(string)

// Client represents a subscriber with a channel to send messages.
type client struct {
	id     string
	status status

	stream  chan string
	onEvent OnEventFunc

	stopchan chan struct{}

	lock sync.Mutex

	sSEPubSubService *sSEPubSubService

	privateTopics map[string]*topic

	groups map[string]*group
}

// Create a new client
func newClient(sSEPubSubService *sSEPubSubService) *client {
	return &client{
		id:     uuid.New().String(),
		status: Waiting,

		stream:  make(chan string),
		onEvent: nil,

		lock: sync.Mutex{},

		stopchan: nil,

		sSEPubSubService: sSEPubSubService,

		privateTopics: make(map[string]*topic),

		groups: make(map[string]*group),
	}
}

// Stop the client from receiving messages over the event stream
func (c *client) stop() {
	// Lock the client
	c.lock.Lock()
	defer c.lock.Unlock()

	// Stop the client
	c.status = Waiting

	// Close the stream
	if c.stream != nil {
		close(c.stream)
	}
}

// Get ID
func (c *client) GetID() string {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.id
}

// Get Status
func (c *client) GetStatus() status {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.status
}

// Get public topics
func (c *client) GetPublicTopics() map[string]*topic {
	return c.sSEPubSubService.GetPublicTopics()
}

// Get public topic by name
func (c *client) GetPublicTopicByName(name string) (*topic, bool) {
	return c.sSEPubSubService.GetPublicTopicByName(name)
}

// Get private topics
func (c *client) GetPrivateTopics() map[string]*topic {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Create a copy of the private topics
	newmap := make(map[string]*topic)
	for k, v := range c.privateTopics {
		newmap[k] = v
	}

	return newmap
}

// Get private topic by name
func (c *client) GetPrivateTopicByName(name string) (*topic, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	t, ok := c.privateTopics[name]
	return t, ok
}

// Add group
func (c *client) addGroup(g *group) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.groups[g.GetName()] = g
}

// Remove group
func (c *client) removeGroup(g *group) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.groups, g.GetName())
}

// Get groups
func (c *client) GetGroups() map[string]*group {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Create a copy of the groups
	newmap := make(map[string]*group)
	for k, v := range c.groups {
		newmap[k] = v
	}

	return newmap
}

// Get group by name
func (c *client) GetGroupByName(name string) (*group, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	g, ok := c.groups[name]
	return g, ok
}

// Get all topics
func (c *client) GetAllTopics() map[string]*topic {
	c.lock.Lock()
	defer c.lock.Unlock()

	newmap := make(map[string]*topic)
	for k, v := range c.GetPublicTopics() {
		newmap[k] = v
	}
	for _, v := range c.groups {
		for k, v := range v.GetTopics() {
			newmap[k] = v
		}
	}
	for k, v := range c.privateTopics {
		newmap[k] = v
	}
	return newmap
}

// Get topic by name
func (c *client) GetTopicByName(name string) (*topic, bool) {
	topics := c.GetAllTopics()

	t, ok := topics[name]
	return t, ok
}

// Get subscribed topics
func (c *client) GetSubscribedTopics() map[string]*topic {
	topics := make(map[string]*topic)
	for k, v := range c.GetAllTopics() {
		if v.IsSubscribed(c) {
			topics[k] = v
		}
	}
	return topics
}

// New private topic
// 0. Check if topic already exists, return it if it does
// 1. Create a new private topic
// 2. Add the topic to the client
// 3. Inform the client about the new topic
func (c *client) NewPrivateTopic(name string) *topic {
	// if topic exists, return it
	if t, ok := c.GetPrivateTopicByName(name); ok {
		return t
	}

	t := newTopic(name, Private)

	c.lock.Lock()
	c.privateTopics[t.GetName()] = t
	c.lock.Unlock()

	// Inform the client about the new topic
	if err := c.sendTopicList(); err != nil {
		log.Errorf("[C:%s]: Error sending new topic to client: %s", c.GetID(), err)
	}

	return t
}

// Remove private topic
// 1. Unsubscribe from the topic
// 2. Remove the topic from the client
// 3. Inform the client about the removed topic by sending the new topic list
func (c *client) RemovePrivateTopic(t *topic) {
	// Remove this topic from all clients
	for _, c := range t.GetClients() {
		c.Unsub(t) // Try to unsubscribe from the topic
	}

	// Remove topic from client
	c.lock.Lock()
	delete(c.privateTopics, t.id)
	c.lock.Unlock()

	// Inform the client about the removed topic by sending the new topic list
	if err := c.sendTopicList(); err != nil {
		log.Errorf("[C:%s]: Error sending new topic to client: %s", c.GetID(), err)
	}
}

// Subscribe to a topic
// 1. If client can subscribe to this topic, add client to topic and return nil
// 2. Inform the client about the new topic by sending this topic as subscribed
func (c *client) Sub(topic *topic) error {
	// if topic exists, add client to topic and return nil
	if t, ok := c.GetTopicByName(topic.GetName()); ok {
		if topic == t {
			t.addClient(c)

			// Inform the client about the new topic by sending this topic as subscribed
			if err := c.sendSubscribedTopic(t); err != nil {
				log.Errorf("[C:%s]: Error sending new topic to client: %s", c.GetID(), err)
			}

			return nil
		}
	}

	return fmt.Errorf("[C:%s]: topic %s does not exist or client can not subscribe to it", c.GetID(), topic.GetName())
}

// Unsubscribe from a topic
// 1. If client is subscribed to this topic, remove client from topic and return nil
// 2. Inform the client about the new topic by sending this topic as unsubscribed
func (c *client) Unsub(topic *topic) error {
	// if topic exists and client is subscribed to it, remove client from topic and return nil
	if t, ok := c.GetTopicByName(topic.GetName()); ok {
		if topic == t {
			if !t.IsSubscribed(c) {
				return fmt.Errorf("[C:%s]: client is not subscribed to topic %s", c.GetID(), topic.GetName())
			}
			t.removeClient(c)

			// Inform the client about the new topic by sending this topic as unsubscribed
			if err := c.sendUnsubscribedTopic(t); err != nil {
				log.Errorf("[C:%s]: Error sending new topic to client: %s", c.GetID(), err)
			}

			return nil
		}
	}

	return fmt.Errorf("[C:%s]: topic %s does not exist or client can not unsubscribe from it", c.GetID(), topic.GetName())
}

// send a message to the client
// 1. Marshal the data
// 2. Put the data into the stream to send it to the client
func (c *client) send(msg interface{}) error {
	// Marshal the data
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Send the data
	if c.GetStatus() == Receving {
		c.lock.Lock()
		c.stream <- string(jsonData)
		c.lock.Unlock()
	} else {
		return fmt.Errorf("[C:%s]: client is not receiving", c.id)
	}

	return nil
}

// sendTopicList sends a message to the client to inform it about the topics
func (c *client) sendTopicList() error {
	// Get all topics
	topics := c.GetAllTopics()

	// Build the JSON data
	fulldata := &eventData{
		Sys: []eventDataSys{
			{
				Type: "topics",
				List: []eventDataSysList{},
			},
		},
	}

	// Append topics data
	for _, topic := range topics {
		t := eventDataSysList{
			Name: topic.GetName(),
			Type: topic.GetType(),
		}

		fulldata.Sys[0].List = append(fulldata.Sys[0].List, t)
	}

	// Send the JSON data to the client
	if err := c.send(fulldata); err != nil {
		return err
	}

	return nil
}

// sendSubscribedTopic sends a message to the client to inform it about the subscribed topic
func (c *client) sendSubscribedTopic(topic *topic) error {
	// Build the JSON data
	fulldata := &eventData{
		Sys: []eventDataSys{
			{
				Type: "subscribed",
				List: []eventDataSysList{
					{
						Name: topic.GetName(),
					},
				},
			},
		},
	}

	// Send the JSON data to the client
	if err := c.send(fulldata); err != nil {
		return err
	}

	return nil
}

// sendUnsubscribedTopic sends a message to the client to inform it about the unsubscribed topic
func (c *client) sendUnsubscribedTopic(topic *topic) error {
	// Build the JSON data
	fulldata := &eventData{
		Sys: []eventDataSys{
			{
				Type: "unsubscribed",
				List: []eventDataSysList{
					{
						Name: topic.GetName(),
					},
				},
			},
		},
	}

	// Send the JSON data to the client
	if err := c.send(fulldata); err != nil {
		return err
	}

	return nil
}

// sendInitMSG generates the initial message to send to the client
// It contains all topics and subscribed topics
func (c *client) sendInitMSG() error {
	// Get all topics and subscribed topics
	topics := c.GetAllTopics()
	subtopics := c.GetSubscribedTopics()

	// Build the JSON data
	fulldata := &eventData{
		Sys: make([]eventDataSys, 0, 2),
	}

	// Append topics data
	if len(topics) > 0 {
		topicData := eventDataSys{Type: "topics"}
		for _, topic := range topics {
			topicData.List = append(topicData.List, eventDataSysList{
				Name: topic.GetName(),
				Type: topic.GetType(),
			})
		}
		fulldata.Sys = append(fulldata.Sys, topicData)
	}

	// Append subscribed topics data
	if len(subtopics) > 0 {
		subTopicData := eventDataSys{Type: "subscribed"}
		for _, topic := range subtopics {
			subTopicData.List = append(subTopicData.List, eventDataSysList{Name: topic.GetName()})
		}
		fulldata.Sys = append(fulldata.Sys, subTopicData)
	}

	// Send JSON data to the client
	return c.send(fulldata)
}

// OnEvent sets the OnEvent function
// This function will be called when a new message needs to be send to the client
func (c *client) OnEvent(f OnEventFunc) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.onEvent = f
}

// RemoveOnEvent removes the OnEvent function
func (c *client) RemoveOnEvent() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.onEvent = nil
}

// Start the client
// 0. Check if client is already receiving
// 1. Set status to Receving and create stop channel
// 2. Send init message to client
// 3. Keep the connection open
// 4. Send message to client if new data is published over the stream
// 5. Stop the client if the stop channel is closed
func (c *client) Start(ctx context.Context) {
	// Set status to Receving and create stop channel
	if c.GetStatus() == Receving {
		return
	}

	// Set status to Receving and create stop channel
	c.lock.Lock()
	c.stopchan = make(chan struct{})
	c.status = Receving
	c.lock.Unlock()

	go func() {
		if err := c.sendInitMSG(); err != nil {
			log.Errorf("[C:%s]: Error sending init message to client: %s", c.GetID(), err)
		}
	}()

	// Keep the connection open until it's closed by the client
	for {
		select {
		case msg := <-c.stream:
			log.Infof("[C:%s] Sending message to client: %s", c.GetID(), msg)
			if c.onEvent == nil {
				log.Warnf("[C:%s] Client has no OnEvent function", c.GetID())
				continue
			}
			c.onEvent(msg)
		case <-ctx.Done():
			log.Infof("[C:%s] Client stopped receiving", c.GetID())
			c.stop()
			return
		case <-c.stopchan:
			log.Infof("[C:%s] Client stopped receiving", c.GetID())
			return
		}
	}
}
