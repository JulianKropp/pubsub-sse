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

	// Send new topics to all clients
	for _, client := range s.clients {
		client.sendNewTopicsList()
	}
	s.lock.Unlock()

	return nil
}

// Remove public topic
func (s *sSEPubSubHandler) RemovePublicTopic(name string) error {
	// if topic does not exists, return error
	if _, exists := s.publicTopics[name]; !exists {
		return fmt.Errorf("topic %s does not exists", name)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	// Unsubscribe all clients from this topic
	for _, client := range s.publicTopics[name].Clients {
		client.Unsub(name)
	}

	// Remove from list of topics. This automaticly removes all clients which have this topic subscribed
	delete(s.publicTopics, name)

	return nil
}

func (s *sSEPubSubHandler) GetTopics() map[string]*topic {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.publicTopics
}
