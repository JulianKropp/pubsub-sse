package pubsubsse

import (
	"sync"

	"github.com/apex/log"
	"github.com/google/uuid"
)

// Group is a collection of topics.
type Group struct {
	// Name is the name of the group.
	name string
	id   string

	lock *sync.Mutex

	// Topics is a map of topic names to topics.
	topics map[string]*Topic

	// Clients is a map of client IDs to clients.
	clients map[string]*Client

	// Events:
	OnNewClient *eventManager[*Client]
	OnNewGroupTopic *eventManager[*Topic]
	OnRemoveClient *eventManager[*Client]
	OnRemoveGroupTopic *eventManager[*Topic]
}

// GroupTopic
type GroupTopic struct {
	Topic *Topic
	Group *Group
}

func newGroup(name string) *Group {
	return &Group{
		name: name,
		id:   uuid.New().String(),

		lock: &sync.Mutex{},

		topics:  map[string]*Topic{},
		clients: map[string]*Client{},

		OnNewClient: newEventManager[*Client](),
		OnNewGroupTopic: newEventManager[*Topic](),
		OnRemoveClient: newEventManager[*Client](),	
		OnRemoveGroupTopic: newEventManager[*Topic](),
	}
}

// GetName returns the name of the group.
func (g *Group) GetName() string {
	g.lock.Lock()
	defer g.lock.Unlock()

	return g.name
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

// Get topic by name
func (g *Group) GetTopicByName(name string) (*Topic, bool) {
	g.lock.Lock()
	defer g.lock.Unlock()

	t, ok := g.topics[name]
	return t, ok
}

// GetClients returns a map of clients.
func (g *Group) GetClients() map[string]*Client {
	g.lock.Lock()
	defer g.lock.Unlock()

	// Create a copy of the clients map
	newmap := make(map[string]*Client)
	for k, v := range g.clients {
		newmap[k] = v
	}

	// Return the clients
	return newmap
}

// Get client by ID
func (g *Group) GetClientByID(id string) (*Client, bool) {
	g.lock.Lock()
	defer g.lock.Unlock()

	c, ok := g.clients[id]
	return c, ok
}

// AddTopic adds a topic to the group.
// 1. Check if topic already exists, return it if it does
// 2. Add the topic to the group
// 3. Inform all clients about the new topic
func (g *Group) NewTopic(name string) *Topic {
	// Check if the topic already exists and return it if it does
	if t, ok := g.GetTopicByName(name); ok {
		return t
	}

	// Create the topic
	t := newTopic(name, TGroup)
	g.lock.Lock()
	g.topics[name] = t
	g.lock.Unlock()

	// Inform all clients about the new topic
	for _, c := range g.GetClients() {
		if err := c.sendTopicList(); err != nil {
			log.Errorf("[C:%s]: Error sending new topic to client: %s", c.id, err)
		}

		// Event:
		c.OnNewTopic.Emit(t)

		gt := &GroupTopic{
			Group: g,
			Topic: t,
		}
		c.OnNewGroupTopic.Emit(gt)
		t.OnNewClient.Emit(c)
	}

	// Event:
	g.OnNewGroupTopic.Emit(t)

	return t
}

// RemoveTopic removes a topic from the group.
// 0. Check if topic is a group topic
// 1. Check if topic exists in the group
// 2. Unsuscribe all clients from the topic
// 3. Remove topic from the group
// 4. Inform all clients about the removed topic
func (g *Group) RemoveTopic(t *Topic) {
	// Check if topic is a group topic
	if t.GetType() != string(TGroup) {
		log.Errorf("topic is not a group topic")
	}

	// Check if topic exists in sSEPubSubService
	checkIfExist := func() bool {
		if top, ok := g.GetTopicByName(t.GetName()); ok {
			if top == t {
				return true
			}
		}
		return false
	}
	if !checkIfExist() {
		log.Errorf("Topic %s does not exist in group", t.GetName())
		return
	}

	// Unsuscribe all clients from the topic
	for _, c := range t.GetClients() {
		if err := c.Unsub(t); err != nil {
			log.Errorf("[C:%s]: Error unsuscribing client from topic: %s", c.id, err)
		}
	}

	// Remove topic from the group
	g.lock.Lock()
	delete(g.topics, t.GetName())
	g.lock.Unlock()

	// Inform all clients about the removed topic
	for _, c := range g.GetClients() {
		if err := c.sendTopicList(); err != nil {
			log.Errorf("[C:%s]: Error sending new topic to client: %s", c.id, err)
		}

		// Event:
		c.OnRemoveTopic.Emit(t)
		gt := &GroupTopic{
			Group: g,
			Topic: t,
		}
		c.OnRemoveGroupTopic.Emit(gt)
		t.OnRemoveClient.Emit(c)
	}

	// Event:
	g.OnRemoveGroupTopic.Emit(t)
}

// AddClient adds a client to the group.
// 0. Check if client already exists in the group
// 1. Add client to the group
// 2. Add group to client
func (g *Group) AddClient(c *Client) {
	// Check if client already exists in the group
	if _, ok := g.GetClientByID(c.GetID()); ok {
		log.Errorf("Client already exists in group")
		return
	}

	// Add client to the group
	g.lock.Lock()
	g.clients[c.GetID()] = c
	g.lock.Unlock()

	// Add group to client. This will inform the client of the new topics
	c.addGroup(g)

	// Inform client about the new topic
	if err := c.sendTopicList(); err != nil {
		log.Errorf("[C:%s]: Error sending new topic to client: %s", c.id, err)
	}

	// Event:
	g.OnNewClient.Emit(c)
}

// RemoveClient removes a client from the group.
// 0. Check if client exists in the group
// 1. Unsubscribe client from all group topics
// 2. Remove client from the group
// 3. Remove group from client
// 4. Inform client about the removed topic
func (g *Group) RemoveClient(c *Client) {
	// Check if client exists in the group
	if _, ok := g.GetClientByID(c.GetID()); !ok {
		log.Errorf("Client does not exist in group")
		return
	}

	// Unsubscribe client from all group topics
	for _, t := range g.GetTopics() {
		if err := c.Unsub(t); err != nil {
			log.Errorf("[C:%s]: Error unsuscribing client from topic: %s", c.id, err)
		}
	}

	// Remove client from the group
	g.lock.Lock()
	delete(g.clients, c.GetID())
	g.lock.Unlock()

	// Remove group from client
	c.removeGroup(g)

	// Inform client about the removed topic
	if err := c.sendTopicList(); err != nil {
		log.Errorf("[C:%s]: Error sending new topic to client: %s", c.id, err)
	}

	// Event:
	g.OnRemoveClient.Emit(c)
}
