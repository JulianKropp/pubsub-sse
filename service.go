package pubsubsse

import (
	"sync"

	"github.com/apex/log"
	"github.com/google/uuid"
)

// SSEPubSubService represents the SSE publisher and subscriber system.
type SSEPubSubService struct {
	id           string
	clients      map[string]*Client
	publicTopics map[string]*Topic
	groups       map[string]*Group

	lock sync.Mutex

	// Events:
	OnNewClient         *eventManager[*Client]
	OnNewPublicTopic    *eventManager[*Topic]
	OnNewGroup          *eventManager[*Group]
	OnRemoveClient      *eventManager[*Client]
	OnRemovePublicTopic *eventManager[*Topic]
	OnRemoveGroup       *eventManager[*Group]
}

// NewSSEPubSub creates a new sSEPubSubService instance.
func NewSSEPubSubService() *SSEPubSubService {
	return &SSEPubSubService{
		id:           "S-" + uuid.New().String(),
		clients:      make(map[string]*Client),
		publicTopics: make(map[string]*Topic),
		groups:       make(map[string]*Group),

		lock: sync.Mutex{},

		OnNewClient:         newEventManager[*Client](),
		OnNewPublicTopic:    newEventManager[*Topic](),
		OnNewGroup:          newEventManager[*Group](),
		OnRemoveClient:      newEventManager[*Client](),
		OnRemovePublicTopic: newEventManager[*Topic](),
		OnRemoveGroup:       newEventManager[*Group](),
	}
}

// GetID returns the ID of the sSEPubSubService.
func (s *SSEPubSubService) GetID() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.id
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
	for _, t := range s.GetPublicTopics() {
		t.OnNewClient.Emit(c)
	}

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
			log.Warnf("[C:%s]: Warning unsubscribing from topic %s: %s", c.GetID(), t.GetID(), err)
		}
	}

	// Remove public topics
	for _, t := range c.GetPublicTopics() {
		c.OnRemoveTopic.Emit(t)
		c.OnRemovePublicTopic.Emit(t)
		t.OnRemoveClient.Emit(c)
	}

	// Remove all private topics
	for _, t := range c.GetPrivateTopics() {
		c.RemovePrivateTopic(t)
	}

	// Remove all groups
	groups := c.GetGroups()
	for _, g := range groups {
		g.RemoveClient(c)
	}

	// stop the client
	c.stop()

	// Lock the sSEPubSubService
	s.lock.Lock()
	defer s.lock.Unlock()

	// Remove client from sSEPubSubService
	delete(s.clients, c.GetID())

	// Emit event
	s.OnRemoveClient.Emit(c)
}

// Add Group
// 0. Check if group already exists, return it if it does
// 1. Create a new group
// 2. Add the group to the sSEPubSubService
func (s *SSEPubSubService) NewGroup() *Group {
	// Create a new group
	g := newGroup()

	// Add the group to the sSEPubSubService
	s.lock.Lock()
	s.groups[g.GetID()] = g
	s.lock.Unlock()

	// Emit event
	s.OnNewGroup.Emit(g)

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
		if group, ok := s.GetGroupByID(g.GetID()); ok {
			if group == g {
				return true
			}
		}
		return false
	}
	if !checkIfExist() {
		log.Errorf("Group %s does not exist in sSEPubSubService", g.GetID())
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
	delete(s.groups, g.GetID())
	s.lock.Unlock()

	// Emit event
	s.OnRemoveGroup.Emit(g)
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

// Get group by ID
func (s *SSEPubSubService) GetGroupByID(id string) (*Group, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	g, ok := s.groups[id]
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
func (s *SSEPubSubService) NewPublicTopic() *Topic {
	// Create a new public topic
	t := newTopic(TPublic)
	s.lock.Lock()
	s.publicTopics[t.GetID()] = t
	s.lock.Unlock()

	// Inform all clients about the new topic
	for _, c := range s.GetClients() {
		if err := c.sendTopicList(); err != nil {
			log.Warnf("[C:%s]: Warning sending new topic to client: %s", c.id, err)
		}

		// event:
		c.OnNewTopic.Emit(t)
		t.OnNewClient.Emit(c)
	}

	// Emit event
	s.OnNewPublicTopic.Emit(t)

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
		log.Errorf("Topic %s is not public", t.GetID())
		return
	}

	// Check if topic exists in sSEPubSubService
	checkIfExist := func() bool {
		if top, ok := s.GetPublicTopicByID(t.GetID()); ok {
			if top == t {
				return true
			}
		}
		return false
	}
	if !checkIfExist() {
		log.Errorf("Topic %s does not exist in sSEPubSubService", t.GetID())
		return
	}

	// Remove this topic from all clients
	for _, c := range t.GetClients() {
		if err := c.Unsub(t); err != nil {
			log.Warnf("[C:%s]: Warning unsubscribing from topic %s: %s", c.GetID(), t.GetID(), err)
		}
	}

	// Remove topic from sSEPubSubService
	s.lock.Lock()
	delete(s.publicTopics, t.GetID())
	s.lock.Unlock()

	// Inform all clients about the removed topic by sending the new topic list
	for _, c := range s.GetClients() {
		if err := c.sendTopicList(); err != nil {
			log.Warnf("[C:%s]: Warning sending new topic to client: %s", c.id, err)
		}

		// Emit event
		c.OnRemoveTopic.Emit(t)
		t.OnRemoveClient.Emit(c)
	}

	// Emit event
	s.OnRemovePublicTopic.Emit(t)
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

// Get public topic by ID
func (s *SSEPubSubService) GetPublicTopicByID(id string) (*Topic, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	t, ok := s.publicTopics[id]
	return t, ok
}
