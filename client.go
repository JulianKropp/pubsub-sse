package pubsubsse

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/google/uuid"
)

type Status int

const (
	Created Status = iota
	Waiting 
	Receiving
	Timeout
	Stopped
)

// Convert status to string
func StatusToString(s Status) string {
	switch s {
	case Created:
		return "Created"
	case Waiting:
		return "Waiting"
	case Receiving:
		return "Receiving"
	case Timeout:
		return "Timeout"
	case Stopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}

type onEventFunc func(string)

// Client represents a subscriber with a channel to send messages.
type Client struct {
	id     string
	status Status

	stream chan string

	stopchan chan struct{}

	lock sync.Mutex

	clientTimout time.Duration

	sSEPubSubService *SSEPubSubService

	privateTopics map[string]*Topic

	groups map[string]*Group

	// Events:
	OnStatusChange       *eventManager[Status]
	OnNewTopic           *eventManager[*Topic]
	OnNewPublicTopic     *eventManager[*Topic]
	OnNewPrivateTopic    *eventManager[*Topic]
	OnNewGroupTopic      *eventManager[*GroupTopic]
	OnNewGroup           *eventManager[*Group]
	OnSubToTopic         *eventManager[*Topic]
	OnRemoveTopic        *eventManager[*Topic]
	OnRemovePublicTopic  *eventManager[*Topic]
	OnRemovePrivateTopic *eventManager[*Topic]
	OnRemoveGroupTopic   *eventManager[*GroupTopic]
	OnRemoveGroup        *eventManager[*Group]
	OnUnsubFromTopic     *eventManager[*Topic]
}

// Create a new client
func newClient(sSEPubSubService *SSEPubSubService) *Client {
	c := &Client{
		id:     "C-" + uuid.New().String(),
		status: Created,

		stream: make(chan string, 100),

		lock: sync.Mutex{},

		clientTimout: sSEPubSubService.GetClientTimeout(),

		stopchan: nil,

		sSEPubSubService: sSEPubSubService,

		privateTopics: make(map[string]*Topic),

		groups: make(map[string]*Group),

		// Events:
		OnStatusChange:       newEventManager[Status](),
		OnNewTopic:           newEventManager[*Topic](),
		OnNewPublicTopic:     sSEPubSubService.OnNewPublicTopic,
		OnNewPrivateTopic:    newEventManager[*Topic](),
		OnNewGroupTopic:      newEventManager[*GroupTopic](),
		OnNewGroup:           newEventManager[*Group](),
		OnSubToTopic:         newEventManager[*Topic](),
		OnRemoveTopic:        newEventManager[*Topic](),
		OnRemovePublicTopic:  sSEPubSubService.OnRemovePublicTopic,
		OnRemovePrivateTopic: newEventManager[*Topic](),
		OnRemoveGroupTopic:   newEventManager[*GroupTopic](),
		OnRemoveGroup:        newEventManager[*Group](),
		OnUnsubFromTopic:     newEventManager[*Topic](),
	}

	// Start TimeoutCheck
	c.timeoutCheck()

	return c
}

// Stop the client from receiving messages over the event stream
func (c *Client) stop(status ...Status) {
	if c.GetStatus() == Waiting || c.GetStatus() == Timeout || c.GetStatus() == Stopped {
		return
	}

	if len(status) > 0 {
		c.changeStatus(status[0])
	} else {
		// Stop the client
		c.changeStatus(Stopped)
	}

	// Lock the client
	c.lock.Lock()
	defer c.lock.Unlock()

	// Close the stream
	if c.stream != nil {
		close(c.stream)
	}
}

// Get ID
func (c *Client) GetID() string {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.id
}

// Change status
func (c *Client) changeStatus(s Status) {
	c.lock.Lock()
	c.status = s
	c.lock.Unlock()

	// Start TimeoutCheck
	if s == Waiting {
		c.timeoutCheck()
	}

	// Emit event
	c.OnStatusChange.Emit(c.status)
}

// Timeout check
func (c *Client) timeoutCheck() {
	go func() {
		time.Sleep(c.clientTimout)
		if c.GetStatus() != Receiving && c.GetStatus() != Stopped {
			c.changeStatus(Timeout)
		}
	}()
}


// Get Status
func (c *Client) GetStatus() Status {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.status
}

// Get public topics
func (c *Client) GetPublicTopics() map[string]*Topic {
	return c.sSEPubSubService.GetPublicTopics()
}

// Get public topic by id
func (c *Client) GetPublicTopicByID(id string) (*Topic, bool) {
	return c.sSEPubSubService.GetPublicTopicByID(id)
}

// Get private topics
func (c *Client) GetPrivateTopics() map[string]*Topic {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Create a copy of the private topics
	newmap := make(map[string]*Topic)
	for k, v := range c.privateTopics {
		newmap[k] = v
	}

	return newmap
}

// Get private topic by id
func (c *Client) GetPrivateTopicByID(id string) (*Topic, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	t, ok := c.privateTopics[id]
	return t, ok
}

// Add group
func (c *Client) addGroup(g *Group) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.groups[g.GetID()] = g

	// Emit event
	c.OnNewGroup.Emit(g)
	for _, t := range g.GetTopics() {
		c.OnNewTopic.Emit(t)

		gt := &GroupTopic{
			Group: g,
			Topic: t,
		}
		c.OnNewGroupTopic.Emit(gt)
	}
}

// Remove group
func (c *Client) removeGroup(g *Group) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.groups, g.GetID())

	// Emit event
	c.OnRemoveGroup.Emit(g)
	for _, t := range g.GetTopics() {
		c.OnRemoveTopic.Emit(t)
		gt := &GroupTopic{
			Group: g,
			Topic: t,
		}
		c.OnRemoveGroupTopic.Emit(gt)
	}
}

// Get groups
func (c *Client) GetGroups() map[string]*Group {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Create a copy of the groups
	newmap := make(map[string]*Group)
	for k, v := range c.groups {
		newmap[k] = v
	}

	return newmap
}

// Get group by id
func (c *Client) GetGroupByID(id string) (*Group, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	g, ok := c.groups[id]
	return g, ok
}

// Get all topics
func (c *Client) GetAllTopics() map[string]*Topic {
	c.lock.Lock()
	defer c.lock.Unlock()

	newmap := make(map[string]*Topic)
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

// Get topic by id
func (c *Client) GetTopicByID(id string) (*Topic, bool) {
	topics := c.GetAllTopics()

	t, ok := topics[id]
	return t, ok
}

// Get subscribed topics
func (c *Client) GetSubscribedTopics() map[string]*Topic {
	topics := make(map[string]*Topic)
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
func (c *Client) NewPrivateTopic() *Topic {
	t := newTopic(TPrivate)

	c.lock.Lock()
	c.privateTopics[t.GetID()] = t
	c.lock.Unlock()

	// Inform the client about the new topic
	if err := c.sendTopicList(); err != nil {
		log.Warnf("[C:%s]: Warning sending new topic to client: %s", c.GetID(), err)
	}

	// Emit event
	c.OnNewTopic.Emit(t)
	c.OnNewPrivateTopic.Emit(t)
	t.OnNewClient.Emit(c)

	return t
}

// Remove private topic
// 0. Check if topic exists, return error if it does not
// 1. Unsubscribe from the topic
// 2. Remove the topic from the client
// 3. Inform the client about the removed topic by sending the new topic list
func (c *Client) RemovePrivateTopic(t *Topic) {
	// if topic does not exist, return
	if _, ok := c.GetPrivateTopicByID(t.GetID()); !ok {
		log.Errorf("[C:%s]: topic %s does not exist", c.GetID(), t.GetID())
		return
	}

	// Remove this topic from all clients
	for _, c := range t.GetClients() {
		c.Unsub(t) // Try to unsubscribe from the topic
	}

	// Remove topic from client
	c.lock.Lock()
	delete(c.privateTopics, t.GetID())
	c.lock.Unlock()

	// Inform the client about the removed topic by sending the new topic list
	if err := c.sendTopicList(); err != nil {
		log.Warnf("[C:%s]: Warning sending new topic to client: %s", c.GetID(), err)
	}

	// Emit event
	c.OnRemoveTopic.Emit(t)
	c.OnRemovePrivateTopic.Emit(t)
	t.OnRemoveClient.Emit(c)
}

// Subscribe to a topic
// 1. If client can subscribe to this topic, add client to topic and return nil
// 2. Inform the client about the new topic by sending this topic as subscribed
func (c *Client) Sub(topic *Topic) error {
	// if topic exists, add client to topic and return nil
	if t, ok := c.GetTopicByID(topic.GetID()); ok {
		if topic == t {
			t.addClient(c)

			// Inform the client about the new topic by sending this topic as subscribed
			if err := c.sendSubscribedTopic(t); err != nil {
				log.Warnf("[C:%s]: Warning sending new topic to client: %s", c.GetID(), err)
			}

			// Emit event
			c.OnSubToTopic.Emit(t)

			return nil
		}
	}

	return fmt.Errorf("[C:%s]: topic %s does not exist or client can not subscribe to it", c.GetID(), topic.GetID())
}

// Unsubscribe from a topic
// 1. If client is subscribed to this topic, remove client from topic and return nil
// 2. Inform the client about the new topic by sending this topic as unsubscribed
func (c *Client) Unsub(topic *Topic) error {
	// if topic exists and client is subscribed to it, remove client from topic and return nil
	if t, ok := c.GetTopicByID(topic.GetID()); ok {
		if topic == t {
			if !t.IsSubscribed(c) {
				return fmt.Errorf("[C:%s]: client is not subscribed to topic %s", c.GetID(), topic.GetID())
			}
			t.removeClient(c)

			// Inform the client about the new topic by sending this topic as unsubscribed
			if err := c.sendUnsubscribedTopic(t); err != nil {
				log.Warnf("[C:%s]: Warning sending new topic to client: %s", c.GetID(), err)
			}

			// Emit event
			c.OnUnsubFromTopic.Emit(t)

			return nil
		}
	}

	return fmt.Errorf("[C:%s]: topic %s does not exist or client can not unsubscribe from it", c.GetID(), topic.GetID())
}

// send a message to the client
// 1. Marshal the data
// 2. Put the data into the stream to send it to the client
func (c *Client) send(msg interface{}) error {
	// Marshal the data
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Send the data
	if c.GetStatus() == Receiving {
		data := "data: " + string(jsonData) + "\n\n"

		//Try 10 times with 100ms to send data to the stream
		for i := 0; i < 10; i++ {
			select {
			case c.stream <- data:
				// successfully sent
				log.Infof("[C:%s]: push data to stream", c.GetID())
				return nil
			default:
				log.Infof("[C:%s]: stream is full: try: %d", c.GetID(), i)
				time.Sleep(10 * time.Millisecond)
			}
		}
		// handle the case where the channel is full or the client is not receiving
		return fmt.Errorf("[C:%s]: stream is full", c.GetID())
	}
	return fmt.Errorf("[C:%s]: client is not receiving", c.GetID())
}

// sendTopicList sends a message to the client to inform it about the topics
func (c *Client) sendTopicList() error {
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
			ID:   topic.GetID(),
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
func (c *Client) sendSubscribedTopic(topic *Topic) error {
	// Build the JSON data
	fulldata := &eventData{
		Sys: []eventDataSys{
			{
				Type: "subscribed",
				List: []eventDataSysList{
					{
						ID: topic.GetID(),
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
func (c *Client) sendUnsubscribedTopic(topic *Topic) error {
	// Build the JSON data
	fulldata := &eventData{
		Sys: []eventDataSys{
			{
				Type: "unsubscribed",
				List: []eventDataSysList{
					{
						ID: topic.GetID(),
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
func (c *Client) sendInitMSG(onEvent onEventFunc) error {
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
				ID:   topic.GetID(),
				Type: topic.GetType(),
			})
		}
		fulldata.Sys = append(fulldata.Sys, topicData)
	}

	// Append subscribed topics data
	if len(subtopics) > 0 {
		subTopicData := eventDataSys{Type: "subscribed"}
		for _, topic := range subtopics {
			subTopicData.List = append(subTopicData.List, eventDataSysList{ID: topic.GetID()})
		}
		fulldata.Sys = append(fulldata.Sys, subTopicData)
	}

	// Marshal the data
	jsonData, err := json.Marshal(fulldata)
	if err != nil {
		return err
	}

	onEvent("data: " + string(jsonData) + "\n\n")

	// Send JSON data to the client
	return err
}

// Start the client
// 0. Check if client is already receiving
// 1. Set status to Receving and create stop channel
// 2. Send init message to client
// 3. Keep the connection open
// 4. Send message to client if new data is published over the stream
// 5. Stop the client if the stop channel is closed
func (c *Client) Start(ctx context.Context, onEvent onEventFunc) error {
	// Set status to Receving and create stop channel
	if c.GetStatus() == Receiving {
		return fmt.Errorf("[C:%s]: Client is already receiving", c.GetID())
	}

	// Set status to Receving and create stop channel
	c.changeStatus(Receiving)
	c.lock.Lock()
	c.stopchan = make(chan struct{})
	c.stream = make(chan string)
	c.lock.Unlock()

	// Stop the client at the end
	defer func() {
		c.stop(Waiting)
	}()

	if err := c.sendInitMSG(onEvent); err != nil {
		log.Warnf("[C:%s]: Warning sending init message to client: %s", c.GetID(), err)
		return err
	}

	// Keep the connection open until it's closed by the client
loop:
	for {
		select {
		case msg, ok := <-c.stream:
			if !ok {
				log.Infof("[C:%s] Client stopped receiving", c.GetID())
				break loop
			}
			log.Infof("[C:%s] Sending message to client: %s", c.GetID(), msg)
			onEvent(msg)
		case <-ctx.Done():
			log.Infof("[C:%s] Client stopped receiving", c.GetID())
			break loop
		case <-c.stopchan:
			log.Infof("[C:%s] Client stopped receiving", c.GetID())
			break loop
		}
	}
	return nil
}
