package pubsubsse

import (
	"sync"

	"github.com/apex/log"
	"github.com/google/uuid"
)

// Group is a collection of topics.
type Group struct {
	id string

	lock *sync.Mutex

	// Topics is a map of topic IDs to topics.
	topics map[string]*Topic

	// Instances is a map of instance IDs to instances.
	instances map[string]*Instance

	// Events:
	OnNewInstance      *eventManager[*Instance]
	OnNewGroupTopic    *eventManager[*Topic]
	OnRemoveInstance   *eventManager[*Instance]
	OnRemoveGroupTopic *eventManager[*Topic]
}

// GroupTopic
type GroupTopic struct {
	Topic *Topic
	Group *Group
}

func newGroup() *Group {
	return &Group{
		id: "G-" + uuid.New().String(),

		lock: &sync.Mutex{},

		topics:    map[string]*Topic{},
		instances: map[string]*Instance{},

		OnNewInstance:      newEventManager[*Instance](),
		OnNewGroupTopic:    newEventManager[*Topic](),
		OnRemoveInstance:   newEventManager[*Instance](),
		OnRemoveGroupTopic: newEventManager[*Topic](),
	}
}

// GetID returns the ID of the group.
func (g *Group) GetID() string {
	g.lock.Lock()
	defer g.lock.Unlock()

	return g.id
}

// GetTopics returns a map of topics.
func (g *Group) GetTopics() map[string]*Topic {
	g.lock.Lock()
	defer g.lock.Unlock()

	// Create a copy of the topics map
	newmap := make(map[string]*Topic)
	for k, v := range g.topics {
		newmap[k] = v
	}

	// Return the topics
	return newmap
}

// Get topic by id
func (g *Group) GetTopicByID(id string) (*Topic, bool) {
	g.lock.Lock()
	defer g.lock.Unlock()

	t, ok := g.topics[id]
	return t, ok
}

// GetInstances returns a map of instances.
func (g *Group) GetInstances() map[string]*Instance {
	g.lock.Lock()
	defer g.lock.Unlock()

	// Create a copy of the instances map
	newmap := make(map[string]*Instance)
	for k, v := range g.instances {
		newmap[k] = v
	}

	// Return the instances
	return newmap
}

// Get instance by ID
func (g *Group) GetInstanceByID(id string) (*Instance, bool) {
	g.lock.Lock()
	defer g.lock.Unlock()

	c, ok := g.instances[id]
	return c, ok
}

// AddTopic adds a topic to the group.
// 1. Check if topic already exists, return it if it does
// 2. Add the topic to the group
// 3. Inform all instances about the new topic
func (g *Group) NewTopic() *Topic {

	// Create the topic
	t := newTopic(TGroup)
	g.lock.Lock()
	g.topics[t.GetID()] = t
	g.lock.Unlock()

	// Inform all instances about the new topic
	for _, c := range g.GetInstances() {
		if err := c.sendTopicList(); err != nil {
			log.Warnf("[C:%s]: Warning sending new topic to instance: %s", c.id, err)
		}

		// Event:
		c.OnNewTopic.Emit(t)

		gt := &GroupTopic{
			Group: g,
			Topic: t,
		}
		c.OnNewGroupTopic.Emit(gt)
		t.OnNewInstance.Emit(c)
	}

	// Event:
	g.OnNewGroupTopic.Emit(t)

	return t
}

// RemoveTopic removes a topic from the group.
// 0. Check if topic is a group topic
// 1. Check if topic exists in the group
// 2. Unsuscribe all instances from the topic
// 3. Remove topic from the group
// 4. Inform all instances about the removed topic
func (g *Group) RemoveTopic(t *Topic) {
	// Check if topic is a group topic
	if t.GetType() != string(TGroup) {
		log.Errorf("topic is not a group topic")
	}

	// Check if topic exists in sSEPubSubService
	checkIfExist := func() bool {
		if top, ok := g.GetTopicByID(t.GetID()); ok {
			if top == t {
				return true
			}
		}
		return false
	}
	if !checkIfExist() {
		log.Errorf("Topic %s does not exist in group", t.GetID())
		return
	}

	// Unsuscribe all instances from the topic
	for _, c := range t.GetInstances() {
		if err := c.Unsub(t); err != nil {
			log.Warnf("[C:%s]: Warning unsuscribing instance from topic: %s", c.id, err)
		}
	}

	// Remove topic from the group
	g.lock.Lock()
	delete(g.topics, t.GetID())
	g.lock.Unlock()

	// Inform all instances about the removed topic
	for _, c := range g.GetInstances() {
		if err := c.sendTopicList(); err != nil {
			log.Warnf("[C:%s]: Warning sending new topic to instance: %s", c.id, err)
		}

		// Event:
		c.OnRemoveTopic.Emit(t)
		gt := &GroupTopic{
			Group: g,
			Topic: t,
		}
		c.OnRemoveGroupTopic.Emit(gt)
		t.OnRemoveInstance.Emit(c)
	}

	// Event:
	g.OnRemoveGroupTopic.Emit(t)
}

// AddInstance adds a instance to the group.
// 0. Check if instance already exists in the group
// 1. Add instance to the group
// 2. Add group to instance
func (g *Group) AddInstance(c *Instance) {
	// Check if instance already exists in the group
	if _, ok := g.GetInstanceByID(c.GetID()); ok {
		log.Errorf("Instance already exists in group")
		return
	}

	// Add instance to the group
	g.lock.Lock()
	g.instances[c.GetID()] = c
	g.lock.Unlock()

	// Add group to instance. This will inform the instance of the new topics
	c.addGroup(g)

	// Inform instance about the new topic
	if err := c.sendTopicList(); err != nil {
		log.Warnf("[C:%s]: Warning sending new topic to instance: %s", c.id, err)
	}

	// Event:
	g.OnNewInstance.Emit(c)
	for _, t := range g.GetTopics() {
		t.OnNewInstance.Emit(c)
	}
}

// RemoveInstance removes a instance from the group.
// 0. Check if instance exists in the group
// 1. Unsubscribe instance from all group topics
// 2. Remove instance from the group
// 3. Remove group from instance
// 4. Inform instance about the removed topic
func (g *Group) RemoveInstance(c *Instance) {
	// Check if instance exists in the group
	if _, ok := g.GetInstanceByID(c.GetID()); !ok {
		log.Errorf("Instance does not exist in group")
		return
	}

	// Unsubscribe instance from all group topics
	for _, t := range g.GetTopics() {
		if err := c.Unsub(t); err != nil {
			log.Warnf("[C:%s]: Warning unsuscribing instance from topic: %s", c.id, err)
		}

		// Event:
		t.OnRemoveInstance.Emit(c)
	}

	// Remove instance from the group
	g.lock.Lock()
	delete(g.instances, c.GetID())
	g.lock.Unlock()

	// Remove group from instance
	c.removeGroup(g)

	// Inform instance about the removed topic
	if err := c.sendTopicList(); err != nil {
		log.Warnf("[C:%s]: Warning sending new topic to instance: %s", c.id, err)
	}

	// Event:
	g.OnRemoveInstance.Emit(c)
}
