package pubsubsse

import (
	"testing"
	"time"

	"github.com/apex/log"
)

// Tests for events:
// SSEPubSubService.
//     OnNewClient
//     OnNewPublicTopic
//     OnNewGroup
//     OnRemoveClient
//     OnRemovePublicTopic
//     OnRemoveGroup

// client.
//     OnStatusChange
//     OnNewTopic
//     OnNewPublicTopic
//     OnNewPrivateTopic
//     OnNewGroupTopic
//     OnNewGroup
//     OnSubToTopic
//     OnRemoveTopic
//     OnRemovePublicTopic
//     OnRemovePrivateTopic
//     OnRemoveGroupTopic
//     OnRemoveGroup
//     OnUnsubFromTopic

// group.
//     OnNewClient
//     OnNewGroupTopic
//     OnRemoveClient
//     OnRemoveGroupTopic

// topic.
//     OnNewClient
//     OnNewSubOfClient
//     OnPub
//     OnRemoveClient
//     OnUnsubOfClient

// HELPER FUNCTIONS:
func contains(topics []*Topic, topic *Topic) bool {
	for _, t := range topics {
		if t == topic {
			return true
		}
	}
	return false
}

//---------------------------------------------------------------------
// SSEPubSubService.
//---------------------------------------------------------------------

// OnNewClient
// Can be triggert wirth:
// - NewClient
func TestSSEPubSubService_OnNewClient(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Counter of how many times the event was triggered.
	counter := 0

	// client
	var client *Client

	// Event
	s.OnNewClient.Listen(func(c *Client) {
		counter++
		if client != c {
			t.Errorf("Expected %v, got %v", client, c)
		}
	})

	// Create a new client.
	client = s.NewClient()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnNewPublicTopic
// Can be triggert wirth:
// - NewPublicTopic
func TestSSEPubSubService_OnNewPublicTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Counter of how many times the event was triggered.
	counter := 0

	// topic
	var topic *Topic

	// Event
	s.OnNewPublicTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Create a new public topic.
	topic = s.NewPublicTopic()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnNewGroup
// Can be triggert wirth:
// - NewGroup
func TestSSEPubSubService_OnNewGroup(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Counter of how many times the event was triggered.
	counter := 0

	// group
	var group *Group

	// Event
	s.OnNewGroup.Listen(func(g *Group) {
		counter++
		if group != g {
			t.Errorf("Expected %v, got %v", group, g)
		}
	})

	// Create a new group.
	group = s.NewGroup()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnRemoveClient
// Can be triggert wirth:
// - RemoveClient
func TestSSEPubSubService_OnRemoveClient(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Counter of how many times the event was triggered.
	counter := 0

	// client
	var client *Client

	// Event
	s.OnRemoveClient.Listen(func(c *Client) {
		counter++
		if client != c {
			t.Errorf("Expected %v, got %v", client, c)
		}
	})

	// Create a new client.
	client = s.NewClient()

	// Remove the client.
	s.RemoveClient(client)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnRemovePublicTopic
// Can be triggert wirth:
// - RemovePublicTopic
func TestSSEPubSubService_OnRemovePublicTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Counter of how many times the event was triggered.
	counter := 0

	// topic
	var topic *Topic

	// Event
	s.OnRemovePublicTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Create a new public topic.
	topic = s.NewPublicTopic()

	// Remove the topic.
	s.RemovePublicTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnRemoveGroup
// Can be triggert wirth:
// - RemoveGroup
func TestSSEPubSubService_OnRemoveGroup(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Counter of how many times the event was triggered.
	counter := 0

	// group
	var group *Group

	// Event
	s.OnRemoveGroup.Listen(func(g *Group) {
		counter++
		if group != g {
			t.Errorf("Expected %v, got %v", group, g)
		}
	})

	// Create a new group.
	group = s.NewGroup()

	// Remove the group.
	s.RemoveGroup(group)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

//---------------------------------------------------------------------
// client.
//---------------------------------------------------------------------

// OnStatusChange
// Can be triggert wirth:
// - SetStatus
func TestClient_OnStatusChange(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	client := ssePubSub.NewClient()

	expected_status := Waiting

	counter := 0

	client.OnStatusChange.Listen(func(status Status) {
		counter++
		if status != expected_status {
			t.Errorf("Expected %d, got %d", expected_status, status)
		}
	})

	// Start /event
	startEventServer(ssePubSub, t, 8083)

	// Start the client
	expected_status = Receving
	done := make(chan bool)
	connected := make(chan bool)
	httpToEvent(t, client, 8083, connected, done, &[]eventData{})
	<-connected
	expected_status = Waiting
	<-done
	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}
}

// OnNewTopic
// Can be triggert wirth:
// - NewTopic
// - NewPublicTopic
// - NewPrivateTopic
// - NewGroupTopic
func TestClient_OnNewTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// mutex
	mutex := make(chan bool, 5)

	// Counter of how many times the event was triggered.
	counter := 0

	// topic
	var topic *Topic

	// Event
	c.OnNewTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}

		mutex <- true
	})

	// Create a new public topic.
	topic = s.NewPublicTopic()

	<-mutex
	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	// Create a new private topic.
	topic = c.NewPrivateTopic()

	<-mutex
	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	// Create a new group topic.
	group1 := s.NewGroup()
	topic = group1.NewTopic()

	// Add the client to the group.
	group1.AddClient(c)

	<-mutex
	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	// Create a new group topic.
	group2 := s.NewGroup()
	group2.AddClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	// Create a new group topic.
	topic = group2.NewTopic()

	<-mutex
	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}

}

// OnNewPublicTopic
// Can be triggert wirth:
// - NewPublicTopic
func TestClient_OnNewPublicTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Counter of how many times the event was triggered.
	counter := 0

	// topic
	var topic *Topic

	// Event
	c.OnNewPublicTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Create a new public topic.
	topic = s.NewPublicTopic()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnNewPrivateTopic
// Can be triggert wirth:
// - NewPrivateTopic
func TestClient_OnNewPrivateTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Counter of how many times the event was triggered.
	counter := 0

	// topic
	var topic *Topic

	// Event
	c.OnNewPrivateTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Create a new private topic.
	topic = c.NewPrivateTopic()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnNewGroupTopic
// Can be triggert wirth:
// - NewGroupTopic
// - group.AddClient
func TestClient_OnNewGroupTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Create a new group.
	group := s.NewGroup()

	// Add the client to the group.
	group.AddClient(c)

	// Counter of how many times the event was triggered.
	counter := 0

	// topic and group
	var topic *Topic

	// Event
	c.OnNewGroupTopic.Listen(func(gt *GroupTopic) {
		counter++
		if topic != gt.Topic {
			t.Errorf("Expected %v, got %v", topic, gt.Topic)
		}
		if group != gt.Group {
			t.Errorf("Expected %v, got %v", group, gt.Group)
		}
	})

	// Create a new group topic.
	topic = group.NewTopic()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	topic = group.NewTopic()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}
}

// OnNewGroup
// Can be triggert wirth:
// - NewGroup
func TestClient_OnNewGroup(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Counter of how many times the event was triggered.
	counter := 0

	// group
	var group *Group

	// Event
	c.OnNewGroup.Listen(func(g *Group) {
		counter++
		if group != g {
			t.Errorf("Expected %v, got %v", group, g)
		}
	})

	// Create a new group.
	group = s.NewGroup()

	// Add the client to the group.
	group.AddClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnSubToTopic
// Can be triggert wirth:
// - Sub
func TestClient_OnSubToTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	c.OnSubToTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Subscribe to the topic.
	c.Sub(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	topic = c.NewPrivateTopic()
	c.Sub(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	group := s.NewGroup()
	topic = group.NewTopic()
	group.AddClient(c)

	c.Sub(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}
}

// OnRemoveTopic
// Can be triggert wirth:
// - RemovePublicTopic
// - RemovePrivateTopic
// - group.RemoveTopic
// - group.RemoveClient
// - sse.RemoveClient
// - sse.RemoveGroup
func TestClient_OnRemoveTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	id := c.OnRemoveTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Remove the topic.
	s.RemovePublicTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	topic = c.NewPrivateTopic()
	c.RemovePrivateTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	group := s.NewGroup()
	topic = group.NewTopic()
	group.AddClient(c)
	group.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	topic = group.NewTopic()
	group.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}

	c.OnRemoveTopic.Remove(id)
	counter = 0

	var topics []*Topic
	topics = append(topics, s.NewPublicTopic())
	topics = append(topics, c.NewPrivateTopic())
	group = s.NewGroup()
	group.AddClient(c)
	topics = append(topics, group.NewTopic())

	// Event
	c.OnRemoveTopic.Listen(func(top *Topic) {
		counter++
		// if not top in topics
		if !contains(topics, top) {
			t.Errorf("Expected %v, got %v", topics, top)
		}
	})

	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}
}

// OnRemovePublicTopic
// Can be triggert wirth:
// - RemovePublicTopic
// - sse.RemoveClient
func TestClient_OnRemovePublicTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	c.OnRemovePublicTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Remove the topic.
	s.RemovePublicTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	topic = s.NewPublicTopic()
	c.NewPrivateTopic()
	group := s.NewGroup()
	group.AddClient(c)
	group.NewTopic()

	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

}

// OnRemovePrivateTopic
// Can be triggert wirth:
// - RemovePrivateTopic
// - sse.RemoveClient
func TestClient_OnRemovePrivateTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Create a new private topic.
	topic := c.NewPrivateTopic()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	c.OnRemovePrivateTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Remove the topic.
	c.RemovePrivateTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	s.NewPublicTopic()
	topic = c.NewPrivateTopic()
	group := s.NewGroup()
	group.AddClient(c)
	group.NewTopic()

	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}
}

// OnRemoveGroupTopic
// Can be triggert wirth:
// - group.RemoveTopic
// - group.RemoveClient
// - sse.RemoveGroup
// - sse.RemoveClient
func TestClient_OnRemoveGroupTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Create a new group.
	group := s.NewGroup()

	// Add the client to the group.
	group.AddClient(c)

	// Create a new group topic.
	topic := group.NewTopic()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	c.OnRemoveGroupTopic.Listen(func(gt *GroupTopic) {
		counter++
		if topic != gt.Topic {
			t.Errorf("Expected %v, got %v", topic, gt.Topic)
		}
		if group != gt.Group {
			t.Errorf("Expected %v, got %v", group, gt.Group)
		}
	})

	// Remove the topic.
	group.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	topic = group.NewTopic()
	group.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	group = s.NewGroup()
	group.AddClient(c)
	topic = group.NewTopic()

	s.RemoveGroup(group)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	s.NewPublicTopic()
	c.NewPrivateTopic()
	group = s.NewGroup()
	group.AddClient(c)
	topic = group.NewTopic()

	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}
}

// OnRemoveGroup
// Can be triggert wirth:
// - group.RemoveClient
// - sse.RemoveGroup
// - sse.RemoveClient
func TestClient_OnRemoveGroup(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Create a new group.
	group := s.NewGroup()

	// Add the client to the group.
	group.AddClient(c)

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	c.OnRemoveGroup.Listen(func(g *Group) {
		counter++
		if group != g {
			t.Errorf("Expected %v, got %v", group, g)
		}
	})

	// Remove the group.
	s.RemoveGroup(group)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	group = s.NewGroup()
	group.AddClient(c)

	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	group = s.NewGroup()
	group.AddClient(c)

	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}
}

// OnUnsubFromTopic
// Can be triggert wirth:
// - Unsub
// - sse.RemovePublicTopic
// - RemovePrivateTopic
// - group.RemoveTopic
// - sse.RemoveGroup
// - group.RemoveClient
// - sse.RemoveClient
func TestClient_OnUnsubFromTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new client.
	c := s.NewClient()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Subscribe to the topic.
	c.Sub(topic)

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	id := c.OnUnsubFromTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Unsubscribe from the topic.
	c.Unsub(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	c.Sub(topic)
	s.RemovePublicTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	topic = c.NewPrivateTopic()

	c.Sub(topic)
	c.RemovePrivateTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	group := s.NewGroup()
	topic = group.NewTopic()
	group.AddClient(c)

	c.Sub(topic)
	group.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}

	group = s.NewGroup()
	topic = group.NewTopic()
	group.AddClient(c)
	c.Sub(topic)

	group.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 5 {
		t.Errorf("Expected 5, got %d", counter)
	}

	group = s.NewGroup()
	topic = group.NewTopic()
	group.AddClient(c)
	c.Sub(topic)

	s.RemoveGroup(group)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 6 {
		t.Errorf("Expected 6, got %d", counter)
	}

	tops := c.GetAllTopics()
	log.Infof("topics: %d", len(tops))

	c.OnUnsubFromTopic.Remove(id)
	counter = 0

	var topics []*Topic
	topics = append(topics, s.NewPublicTopic())
	topics = append(topics, c.NewPrivateTopic())
	group = s.NewGroup()
	group.AddClient(c)
	topics = append(topics, group.NewTopic())

	// Sub to all topics
	for _, t := range topics {
		if err := c.Sub(t); err != nil {
			log.Errorf("Error subscribing to topic: %s", err.Error())
		}
	}

	log.Infof("subscribed topics: %d", len(c.GetSubscribedTopics()))

	// Event
	c.OnUnsubFromTopic.Listen(func(top *Topic) {
		counter++
		// if not top in topics
		if !contains(topics, top) {
			t.Errorf("Expected %v, got %v", topics, top)
		}
	})

	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

}

//---------------------------------------------------------------------
// group.
//---------------------------------------------------------------------

// OnNewClient
// Can be triggert wirth:
// - AddClient
func TestGroup_OnNewClient(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new group.
	g := s.NewGroup()

	// Create a new client.
	c := s.NewClient()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	g.OnNewClient.Listen(func(cl *Client) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	// Add the client to the group.
	g.AddClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnNewGroupTopic
// Can be triggert wirth:
// - NewTopic
func TestGroup_OnNewGroupTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new group.
	g := s.NewGroup()

	var topic *Topic

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	g.OnNewGroupTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Create a new group topic.
	topic = g.NewTopic()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnRemoveClient
// Can be triggert wirth:
// - RemoveClient
// - sse.RemoveClient
// - sse.RemoveGroup
func TestGroup_OnRemoveClient(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new group.
	g := s.NewGroup()

	// Create a new client.
	c := s.NewClient()

	// Add the client to the group.
	g.AddClient(c)

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	g.OnRemoveClient.Listen(func(cl *Client) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	// Remove the client from the group.
	g.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	g.AddClient(c)

	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	c = s.NewClient()
	g.AddClient(c)

	s.RemoveGroup(g)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}
}

// OnRemoveGroupTopic
// Can be triggert wirth:
// - RemoveTopic
// - sse.RemoveGroup
func TestGroup_OnRemoveGroupTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new group.
	g := s.NewGroup()

	// Create a new group topic.
	topic := g.NewTopic()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	g.OnRemoveGroupTopic.Listen(func(top *Topic) {
		counter++
		if topic != top {
			t.Errorf("Expected %v, got %v", topic, top)
		}
	})

	// Remove the topic.
	g.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	topic = g.NewTopic()

	s.RemoveGroup(g)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}
}

//---------------------------------------------------------------------
// topic.
//---------------------------------------------------------------------

// OnNewClient
// Can be triggert wirth:
// - NewClient
// for public and group topics
func TestTopic_OnNewClient(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	var c *Client

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	id := topic.OnNewClient.Listen(func(cl *Client) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	// Create a new client.
	c = s.NewClient()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	// Clear
	topic.OnNewClient.Remove(id)
	counter = 0

	// Create a new group.
	group := s.NewGroup()
	topic = group.NewTopic()

	// Event
	topic.OnNewClient.Listen(func(cl *Client) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	// Create a new client.
	c = s.NewClient()
	group.AddClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnNewSubOfClient
// Can be triggert wirth:
// - Sub
// for public, private and group topics
func TestTopic_OnNewSubOfClient(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Create a new client.
	c := s.NewClient()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	id := topic.OnNewSubOfClient.Listen(func(cl *Client) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	c.Sub(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	// Clear
	topic.OnNewSubOfClient.Remove(id)
	counter = 0

	c = s.NewClient()

	// Create a new private topic.
	topic = c.NewPrivateTopic()

	// Event
	topic.OnNewSubOfClient.Listen(func(cl *Client) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	c.Sub(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	// Clear
	topic.OnNewSubOfClient.Remove(id)
	counter = 0

	// Create a new group.
	group := s.NewGroup()
	topic = group.NewTopic()

	c = s.NewClient()
	group.AddClient(c)

	// Event
	topic.OnNewSubOfClient.Listen(func(cl *Client) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	c.Sub(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnPub
// Can be triggert wirth:
// - Pub
func TestTopic_OnPub(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Create a new client.
	s.NewClient()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	topic.OnPub.Listen(func(data interface{}) {
		// interface to eventData
		if data.(*eventData).Updates[0].Data != "test" {
			t.Errorf("Expected test, got %s", data.(*eventData).Updates[0].Data)
		}
		counter++
	})

	// Publish to the topic.
	topic.Pub("test")

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnRemoveClient
// Can be triggert wirth:
// - sse.RemovePublicTopic
// - RemovePrivateTopic
// - group.RemoveTopic
// - group.RemoveClient
// - sse.RemoveGroup
// - sse.RemoveClient
func TestTopic_OnRemoveClient(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Create a new client.
	c := s.NewClient()

	// Counter of how many times the event was triggered.
	counter := 0

	helperfunc := func(top *Topic) string {
		// Event
		return top.OnRemoveClient.Listen(func(cl *Client) {
			counter++
			if c != cl {
				t.Errorf("Expected %v, got %v", c, cl)
			}
		})
	}

	id := helperfunc(topic)

	// Remove public topic
	s.RemovePublicTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	topic.OnRemoveClient.Remove(id)

	topic = c.NewPrivateTopic()
	id = helperfunc(topic)
	c.RemovePrivateTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	topic.OnRemoveClient.Remove(id)

	group := s.NewGroup()
	topic = group.NewTopic()
	id = helperfunc(topic)
	group.AddClient(c)

	group.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	topic.OnRemoveClient.Remove(id)

	topic = group.NewTopic()
	id = helperfunc(topic)
	group.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}

	group.AddClient(c)
	s.RemoveGroup(group)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 5 {
		t.Errorf("Expected 5, got %d", counter)
	}

	topic.OnRemoveClient.Remove(id)

	topic = s.NewPublicTopic()
	helperfunc(topic)

	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 6 {
		t.Errorf("Expected 6, got %d", counter)
	}
}

// OnUnsubOfClient
// Can be triggert wirth:
// - Unsub
// - sse.RemovePublicTopic
// - RemovePrivateTopic
// - group.RemoveTopic
// - group.RemoveClient
// - sse.RemoveGroup
// - sse.RemoveClient
func TestTopic_OnUnsubOfClient(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Create a new client.
	c := s.NewClient()

	// Subscribe to the topic.
	c.Sub(topic)

	// Counter of how many times the event was triggered.
	counter := 0

	helperfunc := func(top *Topic) string {
		// Event
		return top.OnUnsubOfClient.Listen(func(cl *Client) {
			counter++
			if c != cl {
				t.Errorf("Expected %v, got %v", c, cl)
			}
		})
	}

	id := helperfunc(topic)

	// Unsubscribe from the topic.
	c.Unsub(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	topic.OnUnsubOfClient.Remove(id)

	topic = s.NewPublicTopic()
	id = helperfunc(topic)
	c.Sub(topic)
	s.RemovePublicTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	topic.OnUnsubOfClient.Remove(id)

	topic = c.NewPrivateTopic()
	id = helperfunc(topic)
	c.Sub(topic)
	c.RemovePrivateTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	topic.OnUnsubOfClient.Remove(id)

	group := s.NewGroup()
	topic = group.NewTopic()
	id = helperfunc(topic)
	group.AddClient(c)

	c.Sub(topic)
	group.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}

	topic.OnUnsubOfClient.Remove(id)

	topic = group.NewTopic()
	c.Sub(topic)
	id = helperfunc(topic)
	group.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 5 {
		t.Errorf("Expected 5, got %d", counter)
	}

	group.AddClient(c)
	c.Sub(topic)
	s.RemoveGroup(group)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 6 {
		t.Errorf("Expected 6, got %d", counter)
	}

	topic.OnUnsubOfClient.Remove(id)

	topic = s.NewPublicTopic()
	helperfunc(topic)

	c.Sub(topic)
	s.RemoveClient(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 7 {
		t.Errorf("Expected 7, got %d", counter)
	}
}
