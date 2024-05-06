package pubsubsse

import (
	"context"
	"fmt"
	"sync"

	"github.com/apex/log"
)

// Instance represents a subscriber with a channel to send messages.
type Instance struct {
	lock sync.Mutex

	id string

	connection *Connection

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

// Get ID
func (i *Instance) GetID() string {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.id
}

// Get connection id
func (i *Instance) GetConnectionID() string {
	return i.getConnection().GetID()
}

// Get connection
func (i *Instance) getConnection() *Connection {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.connection
}

func (i *Instance) GetStatus() Status {
	return i.getConnection().GetStatus()
}


// Get public topics
func (i *Instance) GetPublicTopics() map[string]*Topic {
	return i.getConnection().getSSEPubSubService().GetPublicTopics()
}

// Get public topic by id
func (i *Instance) GetPublicTopicByID(id string) (*Topic, bool) {
	return i.getConnection().getSSEPubSubService().GetPublicTopicByID(id)
}

// New private topic
// 0. Check if topic already exists, return it if it does
// 1. Create a new private topic
// 2. Add the topic to the instance
// 3. Inform the instance about the new topic
func (i *Instance) NewPrivateTopic() *Topic {
	t := newTopic(TPrivate)

	i.lock.Lock()
	i.privateTopics[t.GetID()] = t
	i.lock.Unlock()

	// Inform the instance about the new topic
	if err := i.sendTopicList(); err != nil {
		log.Warnf("[C:%s]: Warning sending new topic to instance: %s", i.GetID(), err)
	}

	// Emit event
	i.OnNewTopic.Emit(t)
	i.OnNewPrivateTopic.Emit(t)
	t.OnNewInstance.Emit(i)

	return t
}

// Remove private topic
// 0. Check if topic exists, return error if it does not
// 1. Unsubscribe from the topic
// 2. Remove the topic from the instance
// 3. Inform the instance about the removed topic by sending the new topic list
func (i *Instance) RemovePrivateTopic(t *Topic) {
	// if topic does not exist, return
	if _, ok := i.GetPrivateTopicByID(t.GetID()); !ok {
		log.Errorf("[C:%s]: topic %s does not exist", i.GetID(), t.GetID())
		return
	}

	// Remove this topic from all instances
	for _, i := range t.GetInstances() {
		i.Unsub(t) // Try to unsubscribe from the topic
	}

	// Remove topic from instance
	i.lock.Lock()
	delete(i.privateTopics, t.GetID())
	i.lock.Unlock()

	// Inform the instance about the removed topic by sending the new topic list
	if err := i.sendTopicList(); err != nil {
		log.Warnf("[C:%s]: Warning sending new topic to instance: %s", i.GetID(), err)
	}

	// Emit event
	i.OnRemoveTopic.Emit(t)
	i.OnRemovePrivateTopic.Emit(t)
	t.OnRemoveInstance.Emit(i)
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

// Subscribe to a topic
// 1. If instance can subscribe to this topic, add instance to topic and return nil
// 2. Inform the instance about the new topic by sending this topic as subscribed
func (i *Instance) Sub(topic *Topic) error {
	// if topic exists, add instance to topic and return nil
	if t, ok := i.GetTopicByID(topic.GetID()); ok {
		if topic == t {
			t.addInstance(i)

			// Inform the instance about the new topic by sending this topic as subscribed
			if err := i.sendSubscribedTopic(t); err != nil {
				log.Warnf("[C:%s]: Warning sending new topic to instance: %s", i.GetID(), err)
			}

			// Emit event
			i.OnSubToTopic.Emit(t)

			return nil
		}
	}

	return fmt.Errorf("[C:%s]: topic %s does not exist or instance can not subscribe to it", i.GetID(), topic.GetID())
}

// Unsubscribe from a topic
// 1. If instance is subscribed to this topic, remove instance from topic and return nil
// 2. Inform the instance about the new topic by sending this topic as unsubscribed
func (i *Instance) Unsub(topic *Topic) error {
	// if topic exists and instance is subscribed to it, remove instance from topic and return nil
	if t, ok := i.GetTopicByID(topic.GetID()); ok {
		if topic == t {
			if !t.IsSubscribed(i) {
				return fmt.Errorf("[C:%s]: instance is not subscribed to topic %s", i.GetID(), topic.GetID())
			}
			t.removeInstance(i)

			// Inform the instance about the new topic by sending this topic as unsubscribed
			if err := i.sendUnsubscribedTopic(t); err != nil {
				log.Warnf("[C:%s]: Warning sending new topic to instance: %s", i.GetID(), err)
			}

			// Emit event
			i.OnUnsubFromTopic.Emit(t)

			return nil
		}
	}

	return fmt.Errorf("[C:%s]: topic %s does not exist or instance can not unsubscribe from it", i.GetID(), topic.GetID())
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

// Get all topics
func (i *Instance) GetAllTopics() map[string]*Topic {
	newmap := make(map[string]*Topic)
	for k, v := range i.GetPublicTopics() {
		newmap[k] = v
	}
	for _, v := range i.GetGroups() {
		for k, v := range v.GetTopics() {
			newmap[k] = v
		}
	}
	for k, v := range i.GetPrivateTopics() {
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

func (i *Instance) send(msg interface{}) error {
	return i.getConnection().send(msg)
}

// sendTopicList sends a message to the instance to inform it about the topics
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

	// Send the JSON data to the instance
	if err := i.getConnection().send(fulldata); err != nil {
		return err
	}

	return nil
}

// sendSubscribedTopic sends a message to the instance to inform it about the subscribed topic
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

	// Send the JSON data to the instance
	if err := i.getConnection().send(fulldata); err != nil {
		return err
	}

	return nil
}

// sendUnsubscribedTopic sends a message to the instance to inform it about the unsubscribed topic
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

	// Send the JSON data to the instance
	if err := i.getConnection().send(fulldata); err != nil {
		return err
	}

	return nil
}

func (i *Instance) Start(ctx context.Context, onEvent onEventFunc) error {
	return i.getConnection().Start(ctx, onEvent)
}