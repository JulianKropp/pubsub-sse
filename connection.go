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

type onEventFunc func(string)

type Connection struct {
	lock sync.Mutex

	id     string
	status Status

	sSEPubSubService *SSEPubSubService

	stream chan string

	stopchan chan struct{}

	connectionTimout time.Duration

	// Clients
	instances map[string]*Instance
}

// Instance represents a subscriber with a channel to send messages.
type Instance struct {
	lock sync.Mutex

	id string

	connection *Connection

	privateTopics map[string]*Topic

	groups map[string]*Group

	// Events:
	OnStatusChange *eventManager[Status]
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

// New connection
func (s *SSEPubSubService) newConnection() *Connection {
	return &Connection{
		id:     "C-" + uuid.New().String(),
		status: Created,

		stream: make(chan string, 100),

		lock: sync.Mutex{},

		connectionTimout: 5 * time.Second,

		stopchan: nil,

		instances: make(map[string]*Instance),
	}
}

// Create a new client
func (c *Connection) newInstance() *Instance {
	return &Instance{
		lock: sync.Mutex{},

		id: "I-" + uuid.New().String(),

		connection: c,

		privateTopics: make(map[string]*Topic),

		groups: make(map[string]*Group),

		// Events:
		OnStatusChange: newEventManager[Status](),
		OnNewTopic:           newEventManager[*Topic](),
		OnNewPublicTopic:     c.sSEPubSubService.OnNewPublicTopic,
		OnNewPrivateTopic:    newEventManager[*Topic](),
		OnNewGroupTopic:      newEventManager[*GroupTopic](),
		OnNewGroup:           newEventManager[*Group](),
		OnSubToTopic:         newEventManager[*Topic](),
		OnRemoveTopic:        newEventManager[*Topic](),
		OnRemovePublicTopic:  c.sSEPubSubService.OnRemovePublicTopic,
		OnRemovePrivateTopic: newEventManager[*Topic](),
		OnRemoveGroupTopic:   newEventManager[*GroupTopic](),
		OnRemoveGroup:        newEventManager[*Group](),
		OnUnsubFromTopic:     newEventManager[*Topic](),
	}
}

// Remove instance
// 1. Unsubscribe from all topics
// 2. Remove all private topics
// 3. Stop the client
// 4. Remove client from sSEPubSubService
func (c *Connection) removeInstance(i *Instance) {
	// Unsubscribe from all topics
	alltopics := i.GetAllTopics()
	for _, t := range alltopics {
		if err := i.Unsub(t); err != nil {
			log.Warnf("[C:%s]: Warning unsubscribing from topic %s: %s", i.GetID(), t.GetID(), err)
		}
	}

	// Remove public topics
	for _, t := range i.GetPublicTopics() {
		i.OnRemoveTopic.Emit(t)
		i.OnRemovePublicTopic.Emit(t)
		t.OnRemoveClient.Emit(i)
	}

	// Remove all private topics
	for _, t := range i.GetPrivateTopics() {
		i.RemovePrivateTopic(t)
	}

	// Remove all groups
	groups := i.GetGroups()
	for _, g := range groups {
		g.RemoveClient(i)
	}

	// Lock the sSEPubSubService
	c.lock.Lock()
	defer c.lock.Unlock()

	// Remove client from sSEPubSubService
	delete(c.instances, i.GetID())

	// Emit event
	c.sSEPubSubService.OnRemoveClient.Emit(i)
}

// Stop the client from receiving messages over the event stream
func (c *Connection) stop(status ...Status) {
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

// Get connection id
func (c *Connection) GetID() string {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.id
}

// GetInstances
func (c *Connection) GetInstances() map[string]*Instance {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.instances
}

// Get ID
func (i *Instance) GetID() string {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.id
}

// Change status
func (c *Connection) changeStatus(s Status) {
	c.lock.Lock()
	c.status = s
	c.lock.Unlock()

	// Start TimeoutCheck
	if s == Waiting {
		go func() {
			time.Sleep(c.connectionTimout)
			if c.GetStatus() == Waiting {
				c.changeStatus(Timeout)
			}
		}()
	}

	log.Infof("[C:%s]: Status changed to %s", c.GetID(), s)

	// Emit event
	for _, i := range c.GetInstances() {
		i.OnStatusChange.Emit(c.status)
	}
}

func (i *Instance) GetStatus() Status {
	return i.connection.status
}

// Get Status
func (c *Connection) GetStatus() Status {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.status
}

// Get public topics
func (i *Instance) GetPublicTopics() map[string]*Topic {
	return i.connection.sSEPubSubService.GetPublicTopics()
}

// Get public topic by id
func (i *Instance) GetPublicTopicByID(id string) (*Topic, bool) {
	return i.connection.sSEPubSubService.GetPublicTopicByID(id)
}

// Get private topics
func (i *Instance) GetPrivateTopics() map[string]*Topic {
	i.lock.Lock()
	defer i.lock.Unlock()

	// Create a copy of the private topics
	newmap := make(map[string]*Topic)
	for k, v := range i.privateTopics {
		newmap[k] = v
	}

	return newmap
}

// Get private topic by id
func (i *Instance) GetPrivateTopicByID(id string) (*Topic, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	t, ok := i.privateTopics[id]
	return t, ok
}

// Add group
func (i *Instance) addGroup(g *Group) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.groups[g.GetID()] = g

	// Emit event
	i.OnNewGroup.Emit(g)
	for _, t := range g.GetTopics() {
		i.OnNewTopic.Emit(t)

		gt := &GroupTopic{
			Group: g,
			Topic: t,
		}
		i.OnNewGroupTopic.Emit(gt)
	}
}

// Remove group
func (i *Instance) removeGroup(g *Group) {
	i.lock.Lock()
	defer i.lock.Unlock()

	delete(i.groups, g.GetID())

	// Emit event
	i.OnRemoveGroup.Emit(g)
	for _, t := range g.GetTopics() {
		i.OnRemoveTopic.Emit(t)
		gt := &GroupTopic{
			Group: g,
			Topic: t,
		}
		i.OnRemoveGroupTopic.Emit(gt)
	}
}

// Get groups
func (i *Instance) GetGroups() map[string]*Group {
	i.lock.Lock()
	defer i.lock.Unlock()

	// Create a copy of the groups
	newmap := make(map[string]*Group)
	for k, v := range i.groups {
		newmap[k] = v
	}

	return newmap
}

// Get group by id
func (i *Instance) GetGroupByID(id string) (*Group, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	g, ok := i.groups[id]
	return g, ok
}

// Get all topics
func (i *Instance) GetAllTopics() map[string]*Topic {
	i.lock.Lock()
	defer i.lock.Unlock()

	newmap := make(map[string]*Topic)
	for k, v := range i.GetPublicTopics() {
		newmap[k] = v
	}
	for _, v := range i.groups {
		for k, v := range v.GetTopics() {
			newmap[k] = v
		}
	}
	for k, v := range i.privateTopics {
		newmap[k] = v
	}
	return newmap
}

// Get topic by id
func (i *Instance) GetTopicByID(id string) (*Topic, bool) {
	topics := i.GetAllTopics()

	t, ok := topics[id]
	return t, ok
}

// Get subscribed topics
func (i *Instance) GetSubscribedTopics() map[string]*Topic {
	topics := make(map[string]*Topic)
	for k, v := range i.GetAllTopics() {
		if v.IsSubscribed(i) {
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
func (i *Instance) NewPrivateTopic() *Topic {
	t := newTopic(TPrivate)

	i.lock.Lock()
	i.privateTopics[t.GetID()] = t
	i.lock.Unlock()

	// Inform the client about the new topic
	if err := i.sendTopicList(); err != nil {
		log.Warnf("[C:%s]: Warning sending new topic to client: %s", i.GetID(), err)
	}

	// Emit event
	i.OnNewTopic.Emit(t)
	i.OnNewPrivateTopic.Emit(t)
	t.OnNewClient.Emit(i)

	return t
}

// Remove private topic
// 0. Check if topic exists, return error if it does not
// 1. Unsubscribe from the topic
// 2. Remove the topic from the client
// 3. Inform the client about the removed topic by sending the new topic list
func (i *Instance) RemovePrivateTopic(t *Topic) {
	// if topic does not exist, return
	if _, ok := i.GetPrivateTopicByID(t.GetID()); !ok {
		log.Errorf("[C:%s]: topic %s does not exist", i.GetID(), t.GetID())
		return
	}

	// Remove this topic from all clients
	for _, i := range t.GetClients() {
		i.Unsub(t) // Try to unsubscribe from the topic
	}

	// Remove topic from client
	i.lock.Lock()
	delete(i.privateTopics, t.GetID())
	i.lock.Unlock()

	// Inform the client about the removed topic by sending the new topic list
	if err := i.sendTopicList(); err != nil {
		log.Warnf("[C:%s]: Warning sending new topic to client: %s", i.GetID(), err)
	}

	// Emit event
	i.OnRemoveTopic.Emit(t)
	i.OnRemovePrivateTopic.Emit(t)
	t.OnRemoveClient.Emit(i)
}

// Subscribe to a topic
// 1. If client can subscribe to this topic, add client to topic and return nil
// 2. Inform the client about the new topic by sending this topic as subscribed
func (i *Instance) Sub(topic *Topic) error {
	// if topic exists, add client to topic and return nil
	if t, ok := i.GetTopicByID(topic.GetID()); ok {
		if topic == t {
			t.addClient(i)

			// Inform the client about the new topic by sending this topic as subscribed
			if err := i.sendSubscribedTopic(t); err != nil {
				log.Warnf("[C:%s]: Warning sending new topic to client: %s", i.GetID(), err)
			}

			// Emit event
			i.OnSubToTopic.Emit(t)

			return nil
		}
	}

	return fmt.Errorf("[C:%s]: topic %s does not exist or client can not subscribe to it", i.GetID(), topic.GetID())
}

// Unsubscribe from a topic
// 1. If client is subscribed to this topic, remove client from topic and return nil
// 2. Inform the client about the new topic by sending this topic as unsubscribed
func (i *Instance) Unsub(topic *Topic) error {
	// if topic exists and client is subscribed to it, remove client from topic and return nil
	if t, ok := i.GetTopicByID(topic.GetID()); ok {
		if topic == t {
			if !t.IsSubscribed(i) {
				return fmt.Errorf("[C:%s]: client is not subscribed to topic %s", i.GetID(), topic.GetID())
			}
			t.removeClient(i)

			// Inform the client about the new topic by sending this topic as unsubscribed
			if err := i.sendUnsubscribedTopic(t); err != nil {
				log.Warnf("[C:%s]: Warning sending new topic to client: %s", i.GetID(), err)
			}

			// Emit event
			i.OnUnsubFromTopic.Emit(t)

			return nil
		}
	}

	return fmt.Errorf("[C:%s]: topic %s does not exist or client can not unsubscribe from it", i.GetID(), topic.GetID())
}

func (i *Instance) send(msg interface{}) error {
	return i.connection.send(msg)
}

// send a message to the client
// 1. Marshal the data
// 2. Put the data into the stream to send it to the client
func (c *Connection) send(msg interface{}) error {
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
func (i *Instance) sendTopicList() error {
	// Get all topics
	topics := i.GetAllTopics()

	// Build the JSON data
	fulldata := &connectionData{
		InstanceData: []instanceData{
			{
				ID: i.GetID(),
				Data: eventData{
					Sys: []eventDataSys{
						{
							Type: "topics",
							List: []eventDataSysList{},
						},
					},
				},
			},
		},
	}

	// Append topics data
	for _, topic := range topics {
		t := eventDataSysList{
			ID:   topic.GetID(),
			Type: topic.GetType(),
		}

		fulldata.InstanceData[0].Data.Sys[0].List = append(fulldata.InstanceData[0].Data.Sys[0].List, t)
	}

	// Send the JSON data to the client
	if err := i.connection.send(fulldata); err != nil {
		return err
	}

	return nil
}

// sendSubscribedTopic sends a message to the client to inform it about the subscribed topic
func (i *Instance) sendSubscribedTopic(topic *Topic) error {
	// Build the JSON data
	fulldata := &connectionData{
		InstanceData: []instanceData{
			{
				ID: i.GetID(),
				Data: eventData{
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
				},
			},
		},
	}

	// Send the JSON data to the client
	if err := i.connection.send(fulldata); err != nil {
		return err
	}

	return nil
}

// sendUnsubscribedTopic sends a message to the client to inform it about the unsubscribed topic
func (i *Instance) sendUnsubscribedTopic(topic *Topic) error {
	// Build the JSON data
	fulldata := &connectionData{
		InstanceData: []instanceData{
			{
				ID: i.GetID(),
				Data: eventData{
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
				},
			},
		},
	}

	// Send the JSON data to the client
	if err := i.connection.send(fulldata); err != nil {
		return err
	}

	return nil
}

// sendInitMSG generates the initial message to send to the client
// It contains all topics and subscribed topics
func (c *Connection) sendInitMSG(onEvent onEventFunc) error {
	fullData := &connectionData{
		InstanceData: make([]instanceData, 0, len(c.instances)),
	}

	// Get all topics and subscribed topics
	for _, i := range c.instances {
		topics := i.GetAllTopics()
		subtopics := i.GetSubscribedTopics()
	
		// Build the JSON data
		instancedata := &eventData{
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
			instancedata.Sys = append(instancedata.Sys, topicData)
		}
	
		// Append subscribed topics data
		if len(subtopics) > 0 {
			subTopicData := eventDataSys{Type: "subscribed"}
			for _, topic := range subtopics {
				subTopicData.List = append(subTopicData.List, eventDataSysList{ID: topic.GetID()})
			}
			instancedata.Sys = append(instancedata.Sys, subTopicData)
		}

		fullData.InstanceData = append(fullData.InstanceData, instanceData{
			ID:   i.GetID(),
			Data: *instancedata,
		})
	}

	// Marshal the data
	jsonData, err := json.Marshal(fullData)
	if err != nil {
		return err
	}

	onEvent("data: " + string(jsonData) + "\n\n")

	// Send JSON data to the client
	return err
}

func (i *Instance) Start(ctx context.Context, onEvent onEventFunc) error {
	return i.connection.Start(ctx, onEvent)
}

// Start the client
// 0. Check if client is already receiving
// 1. Set status to Receving and create stop channel
// 2. Send init message to client
// 3. Keep the connection open
// 4. Send message to client if new data is published over the stream
// 5. Stop the client if the stop channel is closed
func (c *Connection) Start(ctx context.Context, onEvent onEventFunc) error {
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
