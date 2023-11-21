package pubsubsse

import (
	"sync"

	"github.com/apex/log"
	"github.com/google/uuid"
)

// Group is a collection of topics.
type group struct {
	// Name is the name of the group.
	name string
	id  string

	lock *sync.Mutex

	// Topics is a map of topic names to topics.
	topics map[string]*topic

	// Clients is a map of client IDs to clients.
	clients map[string]*client
}

func newGroup(name string) *group {
	return &group{
		name: name,
		id:   uuid.New().String(),

		lock: &sync.Mutex{},

		topics:  map[string]*topic{},
		clients: map[string]*client{},
	}
}

// GetName returns the name of the group.
func (g *group) GetName() string {
	g.lock.Lock()
	defer g.lock.Unlock()

	return g.name
}

// GetID returns the ID of the group.
func (g *group) GetID() string {
	g.lock.Lock()
	defer g.lock.Unlock()

	return g.id
}

// GetTopics returns a map of topics.
func (g *group) GetTopics() map[string]*topic {
	g.lock.Lock()
	defer g.lock.Unlock()

	// Create a copy of the topics map
	newmap := make(map[string]*topic)
	for k, v := range g.topics {
		newmap[k] = v
	}

	// Return the topics
	return newmap
}

// Get topic by name
func (g *group) GetTopicByName(name string) (*topic, bool) {
	g.lock.Lock()
	defer g.lock.Unlock()

	t, ok := g.topics[name]
	return t, ok
}

// GetClients returns a map of clients.
func (g *group) GetClients() map[string]*client {
	g.lock.Lock()
	defer g.lock.Unlock()

	// Create a copy of the clients map
	newmap := make(map[string]*client)
	for k, v := range g.clients {
		newmap[k] = v
	}

	// Return the clients
	return newmap
}

// Get client by ID
func (g *group) GetClientByID(id string) (*client, bool) {
	g.lock.Lock()
	defer g.lock.Unlock()

	c, ok := g.clients[id]
	return c, ok
}

// AddTopic adds a topic to the group.
// 1. Check if topic already exists, return it if it does
// 2. Add the topic to the group
// 3. Inform all clients about the new topic
func (g *group) NewTopic(name string) *topic {
	// Check if the topic already exists and return it if it does
	if t, ok := g.GetTopicByName(name); ok {
		return t
	}

	// Create the topic
	t := newTopic(name, Group)
	g.lock.Lock()
	g.topics[name] = t
	g.lock.Unlock()

	// Inform all clients about the new topic
	for _, c := range g.GetClients() {
		if err := c.sendTopicList(); err != nil {
			log.Errorf("[C:%s]: Error sending new topic to client: %s", c.id, err)
		}
	}

	return t
}

// RemoveTopic removes a topic from the group.
// 0. Check if topic is a group topic
// 1. Check if topic exists in the group
// 2. Unsuscribe all clients from the topic
// 3. Remove topic from the group
// 4. Inform all clients about the removed topic
func (g *group) RemoveTopic(t *topic) {
	// Check if topic is a group topic
	if t.GetType() != string(Group) {
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
	}
}

// AddClient adds a client to the group.
// 0. Check if client already exists in the group
// 1. Add client to the group
// 2. Add group to client
func (g *group) AddClient(c *client) {
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
}

// RemoveClient removes a client from the group.
// 0. Check if client exists in the group
// 1. Unsubscribe client from all group topics
// 2. Remove client from the group
// 3. Remove group from client
// 4. Inform client about the removed topic
func (g *group) RemoveClient(c *client) {
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
}