package pubsubsse

import (
	"sync"

	"github.com/apex/log"
)

// sSEPubSubService represents the SSE publisher and subscriber system.
type sSEPubSubService struct {
	clients      map[string]*client
	publicTopics map[string]*topic
	groups       map[string]*group

	lock sync.Mutex
}

// NewSSEPubSub creates a new sSEPubSubService instance.
func NewSSEPubSubService() *sSEPubSubService {
	return &sSEPubSubService{
		clients:      make(map[string]*client),
		publicTopics: make(map[string]*topic),
		groups:       make(map[string]*group),

		lock: sync.Mutex{},
	}
}

// Create new client
func (s *sSEPubSubService) NewClient() *client {
	// Lock the sSEPubSubService
	s.lock.Lock()
	defer s.lock.Unlock()

	c := newClient(s)
	s.clients[c.GetID()] = c
	return c
}

// Remove client
// 1. Unsubscribe from all topics
// 2. Remove all private topics
// 3. Stop the client
// 4. Remove client from sSEPubSubService
func (s *sSEPubSubService) RemoveClient(c *client) {
	// Unsubscribe from all topics
	alltopics := c.GetAllTopics()
	for _, t := range alltopics {
		if err := c.Unsub(t); err != nil {
			log.Errorf("[C:%s]: Error unsubscribing from topic %s: %s", c.GetID(), t.GetName(), err)
		}
	}

	// Remove all private topics
	for _, t := range c.GetPrivateTopics() {
		c.RemovePrivateTopic(t)
	}

	// stop the client
	c.stop()

	// Lock the sSEPubSubService
	s.lock.Lock()
	defer s.lock.Unlock()

	// Remove client from sSEPubSubService
	delete(s.clients, c.GetID())
}

// Add Group
// 0. Check if group already exists, return it if it does
// 1. Create a new group
// 2. Add the group to the sSEPubSubService
func (s *sSEPubSubService) NewGroup(name string) *group {
	// Check if group already exists, return it if it does
	if g, ok := s.GetGroupByName(name); ok {
		return g
	}

	// Create a new group
	g := newGroup(name)

	// Add the group to the sSEPubSubService
	s.lock.Lock()
	s.groups[g.GetName()] = g
	s.lock.Unlock()

	return g
}

// Remove group
// 0. Check if group exists in sSEPubSubService
// 1. Remove all topics from the group
// 2. Remove all clients from the group
// 3. Remove group from sSEPubSubService
func (s *sSEPubSubService) RemoveGroup(g *group) {
	// Check if group exists in sSEPubSubService
	checkIfExist := func() bool {
		if group, ok := s.GetGroupByName(g.GetName()); ok {
			if group == g {
				return true
			}
		}
		return false
	}
	if !checkIfExist() {
		log.Errorf("Group %s does not exist in sSEPubSubService", g.GetName())
		return
	}

	// Remove all topics from the group
	for _, t := range g.GetTopics() {
		g.RemoveTopic(t)
	}

	// Remove all clients from the group
	for _, c := range g.GetClients() {
		g.RemoveClient(c)
	}

	// Remove group from sSEPubSubService
	s.lock.Lock()
	delete(s.groups, g.GetName())
	s.lock.Unlock()
}

// Get groups
func (s *sSEPubSubService) GetGroups() map[string]*group {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Create a copy of the map
	newmap := make(map[string]*group)
	for k, v := range s.groups {
		newmap[k] = v
	}

	return newmap
}

// Get group by name
func (s *sSEPubSubService) GetGroupByName(name string) (*group, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	g, ok := s.groups[name]
	return g, ok
}

// Get clients
func (s *sSEPubSubService) GetClients() map[string]*client {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Create a copy of the map
	newmap := make(map[string]*client)
	for k, v := range s.clients {
		newmap[k] = v
	}

	return newmap
}

// Get client by ID
func (s *sSEPubSubService) GetClientByID(id string) (*client, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	c, ok := s.clients[id]
	return c, ok
}

// Create new public topic
// 0. Check if topic already exists, return it if it does
// 1. Create a new public topic
// 2. Add the topic to the sSEPubSubService
// 3. Inform all clients about the new topic
func (s *sSEPubSubService) NewPublicTopic(name string) *topic {
	// Check if topic already exists, return it if it does
	if t, ok := s.GetPublicTopicByName(name); ok {
		return t
	}

	// Create a new public topic
	t := newTopic(name, Public)
	s.lock.Lock()
	s.publicTopics[t.GetName()] = t
	s.lock.Unlock()

	// Inform all clients about the new topic
	for _, c := range s.GetClients() {
		if err := c.sendTopicList(); err != nil {
			log.Errorf("[C:%s]: Error sending new topic to client: %s", c.id, err)
		}
	}

	return t
}

// Remove public topic
// 0. Check if topic is public
// 1. Unsubscribe all clients from the topic
// 2. Check if topic exists in sSEPubSubService
// 3. Remove topic from sSEPubSubService
// 4. Inform all clients about the removed topic by sending the new topic list
func (s *sSEPubSubService) RemovePublicTopic(t *topic) {
	// Check if topic is public
	if t.GetType() != string(Public) {
		log.Errorf("Topic %s is not public", t.GetName())
		return
	}

	// Check if topic exists in sSEPubSubService
	checkIfExist := func() bool {
		if top, ok := s.GetPublicTopicByName(t.GetName()); ok {
			if top == t {
				return true
			}
		}
		return false
	}
	if !checkIfExist() {
		log.Errorf("Topic %s does not exist in sSEPubSubService", t.GetName())
		return
	}

	// Remove this topic from all clients
	for _, c := range t.GetClients() {
		if err := c.Unsub(t); err != nil {
			log.Errorf("[C:%s]: Error unsubscribing from topic %s: %s", c.GetID(), t.GetName(), err)
		}
	}

	// Remove topic from sSEPubSubService
	s.lock.Lock()
	delete(s.publicTopics, t.GetName())
	s.lock.Unlock()

	// Inform all clients about the removed topic by sending the new topic list
	for _, c := range s.GetClients() {
		if err := c.sendTopicList(); err != nil {
			log.Errorf("[C:%s]: Error sending new topic to client: %s", c.id, err)
		}
	}
}

// Get public topics
func (s *sSEPubSubService) GetPublicTopics() map[string]*topic {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Create a copy of the map
	newmap := make(map[string]*topic)
	for k, v := range s.publicTopics {
		newmap[k] = v
	}

	return newmap
}

// Get public topic by name
func (s *sSEPubSubService) GetPublicTopicByName(name string) (*topic, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	t, ok := s.publicTopics[name]
	return t, ok
}
