package main

import (
	"fmt"
	"sync"
)

// Topic Types
type topicType string

const (
	Public topicType = "public"
	Private topicType = "private"
	Group topicType = "group"
)

// Topic represents a messaging topic in the SSE pub-sub system.
type topic struct {
	Name    string
	Type    topicType
	Clients map[string]*client
	lock    sync.Mutex
}

// Add new public topic
func (s *sSEPubSubHandler) NewPublicTopic(name string) error {
	// if topic already exists, return error
	if _, exists := s.publicTopics[name]; exists {
		return fmt.Errorf("topic %s already exists", name)
	}

	top := &topic{
		Name:    name,
		Type:    Public,
		Clients: make(map[string]*client),
		lock:    sync.Mutex{},
	}

	// Add to list of topics
	s.lock.Lock()
	s.publicTopics[name] = top
	s.lock.Unlock()

	return nil
}

// Add new private private topic
func (s *sSEPubSubHandler) NewPrivateTopic(name string, client *client) error {
	return client.NewPrivateTopic(name)
}

// Remove public topic
func (s *sSEPubSubHandler) RemovePublicTopic(name string) error {
	// if topic does not exists, return error
	if _, exists := s.publicTopics[name]; !exists {
		return fmt.Errorf("topic %s does not exists", name)
	}

	// Remove from list of topics. This automaticly removes all clients which have this topic subscribed
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.publicTopics, name)

	return nil
}

// Remove private private topic
func (s *sSEPubSubHandler) RemovePrivateTopic(name string, client *client) error {
	return client.RemovePrivateTopic(name)
}

// Remove client from topic
func (t *topic) removeClient(id string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.Clients != nil {
		delete(t.Clients, id)
	}
}

func (s *sSEPubSubHandler) GetTopics() map[string]*topic {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.publicTopics
}
