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

type Connection struct {
	lock sync.Mutex

	id     string
	status Status

	sSEPubSubService *SSEPubSubService

	stream chan string

	stopchan chan struct{}

	connectionTimout time.Duration

	// Instances
	instances map[string]*Instance
}

// New connection
func (s *SSEPubSubService) newConnection() *Connection {
	return &Connection{
		id:     "C-" + uuid.New().String(),
		status: Created,

		stream: make(chan string, 100),

		sSEPubSubService: s,

		lock: sync.Mutex{},

		connectionTimout: s.GetInstanceTimeout(),

		stopchan: nil,

		instances: make(map[string]*Instance),
	}
}

// Create a new instance
func (c *Connection) newInstance() *Instance {
	in := &Instance{
		lock: sync.Mutex{},

		id: "I-" + uuid.New().String(),

		sSEPubSubService: c.getSSEPubSubService(),

		connection: c,

		privateTopics: make(map[string]*Topic),

		groups: make(map[string]*Group),

		// Events:
		OnStatusChange:       newEventManager[Status](),
		OnNewTopic:           newEventManager[*Topic](),
		OnNewPublicTopic:     c.getSSEPubSubService().OnNewPublicTopic,
		OnNewPrivateTopic:    newEventManager[*Topic](),
		OnNewGroupTopic:      newEventManager[*GroupTopic](),
		OnNewGroup:           newEventManager[*Group](),
		OnSubToTopic:         newEventManager[*Topic](),
		OnRemoveTopic:        newEventManager[*Topic](),
		OnRemovePublicTopic:  c.getSSEPubSubService().OnRemovePublicTopic,
		OnRemovePrivateTopic: newEventManager[*Topic](),
		OnRemoveGroupTopic:   newEventManager[*GroupTopic](),
		OnRemoveGroup:        newEventManager[*Group](),
		OnUnsubFromTopic:     newEventManager[*Topic](),
	}

	// Add instance to the connection
	id := in.GetID()
	c.lock.Lock()
	c.instances[id] = in
	c.lock.Unlock()

	return in
}

// Add instance
func (c *Connection) addInstance(i *Instance) {
	// Add instance to the connection
	id := i.GetID()
	c.lock.Lock()
	c.instances[id] = i
	c.lock.Unlock()
}

// Remove instance
// 1. Unsubscribe from all topics
// 2. Remove all private topics
// 3. Stop the instance
// 4. Remove instance from sSEPubSubService
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
		t.OnRemoveInstance.Emit(i)
	}

	// Remove all private topics
	for _, t := range i.GetPrivateTopics() {
		i.RemovePrivateTopic(t)
	}

	// Remove all groups
	groups := i.GetGroups()
	for _, g := range groups {
		g.RemoveInstance(i)
	}

	// Remove instance from sSEPubSubService
	id := i.GetID()
	c.lock.Lock()
	delete(c.instances, id)
	c.lock.Unlock()

	// Emit event
	c.getSSEPubSubService().OnRemoveInstance.Emit(i)
}

// Get connection id
func (c *Connection) GetID() string {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.id
}

// Get Status
func (c *Connection) GetStatus() Status {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.status
}

// Get getSSEPubSubService
func (c *Connection) getSSEPubSubService() *SSEPubSubService {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.sSEPubSubService
}

// GetInstances
func (c *Connection) GetInstances() map[string]*Instance {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.instances
}

// Change status
func (c *Connection) changeStatus(s Status) {
	c.lock.Lock()
	c.status = s
	c.lock.Unlock()

	// Start TimeoutCheck
	if s == Waiting {
		c.timeoutCheck()
	}

	log.Infof("[C:%s]: Status changed to %s", c.GetID(), StatusToString(s))

	// Emit event
	for _, i := range c.GetInstances() {
		i.OnStatusChange.Emit(c.GetStatus())
	}

	if s == Stopped || s == Timeout {
		// Remove all instances
		for _, i := range c.GetInstances() {
			c.getSSEPubSubService().RemoveInstance(i)
		}
	}
}

// Timeout check
func (c *Connection) timeoutCheck() {
	go func() {
		c.lock.Lock()
		timeout := c.connectionTimout
		c.lock.Unlock()

		time.Sleep(timeout)
		if c.GetStatus() != Receiving && c.GetStatus() != Stopped {
			c.changeStatus(Timeout)
		}
	}()
}

// send a message to the instance
// 1. Marshal the data
// 2. Put the data into the stream to send it to the instance
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
		// handle the case where the channel is full or the instance is not receiving
		return fmt.Errorf("[C:%s]: stream is full", c.GetID())
	}
	return fmt.Errorf("[C:%s]: instance is not receiving", c.GetID())
}

// sendInitMSG generates the initial message to send to the instance
// It contains all topics and subscribed topics
func (c *Connection) sendInitMSG(onEvent onEventFunc) error {
	fullData := &connectionData{
		InstanceData: make([]instanceData, 0, len(c.GetInstances())),
	}

	// Get all topics and subscribed topics
	for _, i := range c.GetInstances() {
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

	// Send JSON data to the instance
	return err
}

// Start the instance
// 0. Check if instance is already receiving
// 1. Set status to Receving and create stop channel
// 2. Send init message to instance
// 3. Keep the connection open
// 4. Send message to instance if new data is published over the stream
// 5. Stop the instance if the stop channel is closed
func (c *Connection) Start(ctx context.Context, onEvent onEventFunc) error {
	// Set status to Receving and create stop channel
	if c.GetStatus() == Receiving {
		return fmt.Errorf("[C:%s]: Instance is already receiving", c.GetID())
	}

	// Set status to Receving and create stop channel
	c.changeStatus(Receiving)
	c.lock.Lock()
	c.stopchan = make(chan struct{})
	c.stream = make(chan string)
	c.lock.Unlock()

	// Stop the instance at the end
	defer func() {
		c.stop(Waiting)
	}()

	if err := c.sendInitMSG(onEvent); err != nil {
		log.Warnf("[C:%s]: Warning sending init message to instance: %s", c.GetID(), err)
		return err
	}

	// Keep the connection open until it's closed by the instance
loop:
	for {
		select {
		case msg, ok := <-c.stream:
			if !ok {
				log.Infof("[C:%s] Connection stopped receiving", c.GetID())
				break loop
			}
			log.Infof("[C:%s] Sending message to connection: %s", c.GetID(), msg)
			onEvent(msg)
		case <-ctx.Done():
			log.Infof("[C:%s] Connection stopped receiving", c.GetID())
			break loop
		case <-c.stopchan:
			log.Infof("[C:%s] Connection stopped receiving", c.GetID())
			break loop
		}
	}
	return nil
}

// Stop the instance from receiving messages over the event stream
func (c *Connection) stop(status ...Status) {
	st := c.GetStatus()
	if st == Waiting || st == Timeout || st == Stopped {
		return
	}

	if len(status) > 0 {
		c.changeStatus(status[0])
	} else {
		// Stop the instance
		c.changeStatus(Stopped)
	}

	// Lock the instance
	c.lock.Lock()
	defer c.lock.Unlock()

	// Close the stream
	if c.stream != nil {
		close(c.stream)
	}
}