package pubsubsse

import (
	"sync"

	"github.com/apex/log"
)

// SSEPubSubService represents the SSE publisher and subscriber system.
type SSEPubSubService struct {
	clients      map[string]*Client
	publicTopics map[string]*Topic
	groups       map[string]*Group

	lock sync.Mutex

	// Events:
	OnNewClient *eventManager[Client]
}

// NewSSEPubSub creates a new sSEPubSubService instance.
func NewSSEPubSubService() *SSEPubSubService {
	return &SSEPubSubService{
		clients:      make(map[string]*Client),
		publicTopics: make(map[string]*Topic),
		groups:       make(map[string]*Group),

		lock: sync.Mutex{},

		OnNewClient: newEventManager[Client](),
	}
}

// Create new client
func (s *SSEPubSubService) NewClient() *Client {
	// Lock the sSEPubSubService
	s.lock.Lock()

	c := newClient(s)
	s.clients[c.GetID()] = c
	s.lock.Unlock()

	// Emit event
	s.OnNewClient.Emit(c)

	return c
}

// Remove client
// 1. Unsubscribe from all topics
// 2. Remove all private topics
// 3. Stop the client
// 4. Remove client from sSEPubSubService
func (s *SSEPubSubService) RemoveClient(c *Client) {
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
func (s *SSEPubSubService) NewGroup(name string) *Group {
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
func (s *SSEPubSubService) RemoveGroup(g *Group) {
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
func (s *SSEPubSubService) GetGroups() map[string]*Group {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Create a copy of the map
	newmap := make(map[string]*Group)
	for k, v := range s.groups {
		newmap[k] = v
	}

	return newmap
}

// Get group by name
func (s *SSEPubSubService) GetGroupByName(name string) (*Group, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	g, ok := s.groups[name]
	return g, ok
}

// Get clients
func (s *SSEPubSubService) GetClients() map[string]*Client {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Create a copy of the map
	newmap := make(map[string]*Client)
	for k, v := range s.clients {
		newmap[k] = v
	}

	return newmap
}

// Get client by ID
func (s *SSEPubSubService) GetClientByID(id string) (*Client, bool) {
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
func (s *SSEPubSubService) NewPublicTopic(name string) *Topic {
	// Check if topic already exists, return it if it does
	if t, ok := s.GetPublicTopicByName(name); ok {
		return t
	}

	// Create a new public topic
	t := newTopic(name, TPublic)
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
func (s *SSEPubSubService) RemovePublicTopic(t *Topic) {
	// Check if topic is public
	if t.GetType() != string(TPublic) {
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
func (s *SSEPubSubService) GetPublicTopics() map[string]*Topic {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Create a copy of the map
	newmap := make(map[string]*Topic)
	for k, v := range s.publicTopics {
		newmap[k] = v
	}

	return newmap
}

// Get public topic by name
func (s *SSEPubSubService) GetPublicTopicByName(name string) (*Topic, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	t, ok := s.publicTopics[name]
	return t, ok
}
