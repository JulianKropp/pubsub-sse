package pubsubsse

import (
	"testing"
	"time"

	"github.com/apex/log"
)

// Tests for events:
// SSEPubSubService.
//     OnNewInstance
//     OnNewPublicTopic
//     OnNewGroup
//     OnRemoveInstance
//     OnRemovePublicTopic
//     OnRemoveGroup

// instance.
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
//     OnNewInstance
//     OnNewGroupTopic
//     OnRemoveInstance
//     OnRemoveGroupTopic

// topic.
//     OnNewInstance
//     OnNewSubOfInstance
//     OnPub
//     OnRemoveInstance
//     OnUnsubOfInstance

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

// OnNewInstance
// Can be triggert wirth:
// - NewInstance
func TestSSEPubSubService_OnNewInstance(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Counter of how many times the event was triggered.
	counter := 0

	// instance
	var instance *Instance

	// Event
	s.OnNewInstance.Listen(func(c *Instance) {
		counter++
		if instance != c {
			t.Errorf("Expected %v, got %v", instance, c)
		}
	})

	// Create a new instance.
	instance = s.NewInstance()

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

// OnRemoveInstance
// Can be triggert wirth:
// - RemoveInstance
func TestSSEPubSubService_OnRemoveInstance(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Counter of how many times the event was triggered.
	counter := 0

	// instance
	var instance *Instance

	// Event
	s.OnRemoveInstance.Listen(func(c *Instance) {
		counter++
		if instance != c {
			t.Errorf("Expected %v, got %v", instance, c)
		}
	})

	// Create a new instance.
	instance = s.NewInstance()

	// Remove the instance.
	s.RemoveInstance(instance)

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
// instance.
//---------------------------------------------------------------------

// OnStatusChange
// Can be triggert wirth:
// - SetStatus
// - Wait for Tiemout
// - RemoveInstance
func TestInstance_OnStatusChange(t *testing.T) {
	ssePubSub := NewSSEPubSubService()
	ssePubSub.SetInstanceTimeout(1 * time.Second)
	instance := ssePubSub.NewInstance()

	expectedStatuses := make(chan Status, 4) // Channel to synchronize expected statuses
	expectedStatuses <- Receiving            // Queue the expected statuses in order
	expectedStatuses <- Waiting
	expectedStatuses <- Timeout

	counter := 0

	instance.OnStatusChange.Listen(func(status Status) {
		counter++
		expectedStatus := <-expectedStatuses // Receive the next expected status
		t.Log("Status:", status)
		if status != expectedStatus {
			t.Errorf("Expected %d, got %d", expectedStatus, status)
		}
	})

	// Start /event
	startEventServer(ssePubSub, t, 8083)

	// Start the instance
	done := make(chan bool)
	connected := make(chan bool)
	httpToEvent(t, instance, 8083, connected, done, &[]eventData{})
	<-connected
	<-done
	time.Sleep(2 * time.Second) // Wait for the timeout to potentially trigger

	// Check if the event was triggered correctly.
	if counter != 3 {
		t.Errorf("Expected 3 status changes, got %d", counter)
	}

	instance2 := ssePubSub.NewInstance()
	counter = 0

	instance2.OnStatusChange.Listen(func(status Status) {
		counter++
		if status != Stopped {
			t.Errorf("Expected %d, got %d", Stopped, status)
		}
	})

	ssePubSub.RemoveInstance(instance2)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered correctly.
	if counter != 1 {
		t.Errorf("Expected 1 status changes, got %d", counter)
	}
}

// OnNewTopic
// Can be triggert wirth:
// - NewTopic
// - NewPublicTopic
// - NewPrivateTopic
// - NewGroupTopic
func TestInstance_OnNewTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

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

	// Add the instance to the group.
	group1.AddInstance(c)

	<-mutex
	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	// Create a new group topic.
	group2 := s.NewGroup()
	group2.AddInstance(c)

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
func TestInstance_OnNewPublicTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

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
func TestInstance_OnNewPrivateTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

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
// - group.AddInstance
func TestInstance_OnNewGroupTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

	// Create a new group.
	group := s.NewGroup()

	// Add the instance to the group.
	group.AddInstance(c)

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
func TestInstance_OnNewGroup(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

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

	// Add the instance to the group.
	group.AddInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnSubToTopic
// Can be triggert wirth:
// - Sub
func TestInstance_OnSubToTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

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
	group.AddInstance(c)

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
// - group.RemoveInstance
// - sse.RemoveInstance
// - sse.RemoveGroup
func TestInstance_OnRemoveTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

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
	group.AddInstance(c)
	group.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	topic = group.NewTopic()
	group.RemoveInstance(c)

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
	group.AddInstance(c)
	topics = append(topics, group.NewTopic())

	// Event
	c.OnRemoveTopic.Listen(func(top *Topic) {
		counter++
		// if not top in topics
		if !contains(topics, top) {
			t.Errorf("Expected %v, got %v", topics, top)
		}
	})

	s.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}
}

// OnRemovePublicTopic
// Can be triggert wirth:
// - RemovePublicTopic
// - sse.RemoveInstance
func TestInstance_OnRemovePublicTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

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
	group.AddInstance(c)
	group.NewTopic()

	s.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

}

// OnRemovePrivateTopic
// Can be triggert wirth:
// - RemovePrivateTopic
// - sse.RemoveInstance
func TestInstance_OnRemovePrivateTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

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
	group.AddInstance(c)
	group.NewTopic()

	s.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}
}

// OnRemoveGroupTopic
// Can be triggert wirth:
// - group.RemoveTopic
// - group.RemoveInstance
// - sse.RemoveGroup
// - sse.RemoveInstance
func TestInstance_OnRemoveGroupTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

	// Create a new group.
	group := s.NewGroup()

	// Add the instance to the group.
	group.AddInstance(c)

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
	group.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	group = s.NewGroup()
	group.AddInstance(c)
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
	group.AddInstance(c)
	topic = group.NewTopic()

	s.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}
}

// OnRemoveGroup
// Can be triggert wirth:
// - group.RemoveInstance
// - sse.RemoveGroup
// - sse.RemoveInstance
func TestInstance_OnRemoveGroup(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

	// Create a new group.
	group := s.NewGroup()

	// Add the instance to the group.
	group.AddInstance(c)

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
	group.AddInstance(c)

	s.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	group = s.NewGroup()
	group.AddInstance(c)

	s.RemoveInstance(c)

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
// - group.RemoveInstance
// - sse.RemoveInstance
func TestInstance_OnUnsubFromTopic(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new instance.
	c := s.NewInstance()

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
	group.AddInstance(c)

	c.Sub(topic)
	group.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}

	group = s.NewGroup()
	topic = group.NewTopic()
	group.AddInstance(c)
	c.Sub(topic)

	group.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 5 {
		t.Errorf("Expected 5, got %d", counter)
	}

	group = s.NewGroup()
	topic = group.NewTopic()
	group.AddInstance(c)
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
	group.AddInstance(c)
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

	s.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

}

//---------------------------------------------------------------------
// group.
//---------------------------------------------------------------------

// OnNewInstance
// Can be triggert wirth:
// - AddInstance
func TestGroup_OnNewInstance(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new group.
	g := s.NewGroup()

	// Create a new instance.
	c := s.NewInstance()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	g.OnNewInstance.Listen(func(cl *Instance) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	// Add the instance to the group.
	g.AddInstance(c)

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

// OnRemoveInstance
// Can be triggert wirth:
// - RemoveInstance
// - sse.RemoveInstance
// - sse.RemoveGroup
func TestGroup_OnRemoveInstance(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new group.
	g := s.NewGroup()

	// Create a new instance.
	c := s.NewInstance()

	// Add the instance to the group.
	g.AddInstance(c)

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	g.OnRemoveInstance.Listen(func(cl *Instance) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	// Remove the instance from the group.
	g.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	g.AddInstance(c)

	s.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	c = s.NewInstance()
	g.AddInstance(c)

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

// OnNewInstance
// Can be triggert wirth:
// - NewInstance
// for public and group topics
func TestTopic_OnNewInstance(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	var c *Instance

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	id := topic.OnNewInstance.Listen(func(cl *Instance) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	// Create a new instance.
	c = s.NewInstance()

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}

	// Clear
	topic.OnNewInstance.Remove(id)
	counter = 0

	// Create a new group.
	group := s.NewGroup()
	topic = group.NewTopic()

	// Event
	topic.OnNewInstance.Listen(func(cl *Instance) {
		counter++
		if c != cl {
			t.Errorf("Expected %v, got %v", c, cl)
		}
	})

	// Create a new instance.
	c = s.NewInstance()
	group.AddInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 1 {
		t.Errorf("Expected 1, got %d", counter)
	}
}

// OnNewSubOfInstance
// Can be triggert wirth:
// - Sub
// for public, private and group topics
func TestTopic_OnNewSubOfInstance(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Create a new instance.
	c := s.NewInstance()

	// Counter of how many times the event was triggered.
	counter := 0

	// Event
	id := topic.OnNewSubOfInstance.Listen(func(cl *Instance) {
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
	topic.OnNewSubOfInstance.Remove(id)
	counter = 0

	c = s.NewInstance()

	// Create a new private topic.
	topic = c.NewPrivateTopic()

	// Event
	topic.OnNewSubOfInstance.Listen(func(cl *Instance) {
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
	topic.OnNewSubOfInstance.Remove(id)
	counter = 0

	// Create a new group.
	group := s.NewGroup()
	topic = group.NewTopic()

	c = s.NewInstance()
	group.AddInstance(c)

	// Event
	topic.OnNewSubOfInstance.Listen(func(cl *Instance) {
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

	// Create a new instance.
	s.NewInstance()

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

// OnRemoveInstance
// Can be triggert wirth:
// - sse.RemovePublicTopic
// - RemovePrivateTopic
// - group.RemoveTopic
// - group.RemoveInstance
// - sse.RemoveGroup
// - sse.RemoveInstance
func TestTopic_OnRemoveInstance(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Create a new instance.
	c := s.NewInstance()

	// Counter of how many times the event was triggered.
	counter := 0

	helperfunc := func(top *Topic) string {
		// Event
		return top.OnRemoveInstance.Listen(func(cl *Instance) {
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

	topic.OnRemoveInstance.Remove(id)

	topic = c.NewPrivateTopic()
	id = helperfunc(topic)
	c.RemovePrivateTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	topic.OnRemoveInstance.Remove(id)

	group := s.NewGroup()
	topic = group.NewTopic()
	id = helperfunc(topic)
	group.AddInstance(c)

	group.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	topic.OnRemoveInstance.Remove(id)

	topic = group.NewTopic()
	id = helperfunc(topic)
	group.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}

	group.AddInstance(c)
	s.RemoveGroup(group)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 5 {
		t.Errorf("Expected 5, got %d", counter)
	}

	topic.OnRemoveInstance.Remove(id)

	topic = s.NewPublicTopic()
	helperfunc(topic)

	s.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 6 {
		t.Errorf("Expected 6, got %d", counter)
	}
}

// OnUnsubOfInstance
// Can be triggert wirth:
// - Unsub
// - sse.RemovePublicTopic
// - RemovePrivateTopic
// - group.RemoveTopic
// - group.RemoveInstance
// - sse.RemoveGroup
// - sse.RemoveInstance
func TestTopic_OnUnsubOfInstance(t *testing.T) {
	// Create a new SSEPubSubService.
	s := NewSSEPubSubService()

	// Create a new public topic.
	topic := s.NewPublicTopic()

	// Create a new instance.
	c := s.NewInstance()

	// Subscribe to the topic.
	c.Sub(topic)

	// Counter of how many times the event was triggered.
	counter := 0

	helperfunc := func(top *Topic) string {
		// Event
		return top.OnUnsubOfInstance.Listen(func(cl *Instance) {
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

	topic.OnUnsubOfInstance.Remove(id)

	topic = s.NewPublicTopic()
	id = helperfunc(topic)
	c.Sub(topic)
	s.RemovePublicTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 2 {
		t.Errorf("Expected 2, got %d", counter)
	}

	topic.OnUnsubOfInstance.Remove(id)

	topic = c.NewPrivateTopic()
	id = helperfunc(topic)
	c.Sub(topic)
	c.RemovePrivateTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 3 {
		t.Errorf("Expected 3, got %d", counter)
	}

	topic.OnUnsubOfInstance.Remove(id)

	group := s.NewGroup()
	topic = group.NewTopic()
	id = helperfunc(topic)
	group.AddInstance(c)

	c.Sub(topic)
	group.RemoveTopic(topic)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 4 {
		t.Errorf("Expected 4, got %d", counter)
	}

	topic.OnUnsubOfInstance.Remove(id)

	topic = group.NewTopic()
	c.Sub(topic)
	id = helperfunc(topic)
	group.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 5 {
		t.Errorf("Expected 5, got %d", counter)
	}

	group.AddInstance(c)
	c.Sub(topic)
	s.RemoveGroup(group)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 6 {
		t.Errorf("Expected 6, got %d", counter)
	}

	topic.OnUnsubOfInstance.Remove(id)

	topic = s.NewPublicTopic()
	helperfunc(topic)

	c.Sub(topic)
	s.RemoveInstance(c)

	time.Sleep(100 * time.Millisecond)

	// Check if the event was triggered.
	if counter != 7 {
		t.Errorf("Expected 7, got %d", counter)
	}
}
