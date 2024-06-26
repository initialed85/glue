package topics

import (
	"log"
	"slices"
	"sync"

	"github.com/segmentio/ksuid"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/maps"

	"github.com/initialed85/glue/pkg/fragmentation"
	"github.com/initialed85/glue/pkg/transport"
	"github.com/initialed85/glue/pkg/types"
)

type Subscriber struct {
	mu                                      sync.Mutex
	subscriptionByTopicName                 map[string]*Subscription
	containerByFragmentIndexByCorrelationID map[ksuid.KSUID]map[int64]*types.Container
	endpointID                              ksuid.KSUID
	endpointName                            string
	transportManager                        *transport.Manager
	publisher                               **Publisher
}

func NewSubscriber(
	endpointID ksuid.KSUID,
	endpointName string,
	transportManager *transport.Manager,
	publisher **Publisher,
) *Subscriber {
	s := Subscriber{
		subscriptionByTopicName:                 make(map[string]*Subscription),
		containerByFragmentIndexByCorrelationID: make(map[ksuid.KSUID]map[int64]*types.Container),
		endpointID:                              endpointID,
		endpointName:                            endpointName,
		transportManager:                        transportManager,
		publisher:                               publisher,
	}

	return &s
}

func (s *Subscriber) handleInternalReceive(message *Message) {
	if message == nil {
		log.Printf("warning: subscriber had message unexpectedly nil")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	subscription, ok := s.subscriptionByTopicName[message.TopicName]

	// if we don't have a subscription, see if we're listening for the wildcard topic
	usingWildcard := false
	if !ok {
		subscription, ok = s.subscriptionByTopicName["#"]
		usingWildcard = ok
	}

	// TODO: here's where we'd put the late joiner / persistence stuff
	if !ok {
		return
	}

	// TODO: not sure how to handle type safety and wildcard topics
	// TODO: fix hack usage for the bridge
	if !usingWildcard && message.TopicType != "__mqtt_to_glue_bridge__" && message.TopicType != subscription.topicType {
		log.Printf(
			"warning: expected type %#v for topic %#v but got %#v; message was %#+v",
			subscription.topicType,
			subscription.topicName,
			message.TopicType,
			message,
		)
		return
	}

	subscription.HandleReceive(message)
}

func (s *Subscriber) HandleReceive(container *types.Container) {
	if container == nil || container.Frame == nil || container.Frame.Payload == nil {
		log.Printf("warning: subscriber had Container / Container.Frame / Container.Frame.Payload unexpectedly nil")
		return
	}

	if container.Frame.FragmentCount > 0 {
		skip := true

		s.mu.Lock()
		containerByFragmentIndex, ok := s.containerByFragmentIndexByCorrelationID[container.Frame.CorrelationID]
		if !ok {
			containerByFragmentIndex = make(map[int64]*types.Container)
		}
		containerByFragmentIndex[container.Frame.FragmentIndex] = container

		if len(containerByFragmentIndex) == int(container.Frame.FragmentCount) {
			containers := maps.Values(containerByFragmentIndex)
			slices.SortFunc(containers, func(a *types.Container, b *types.Container) int {
				if a.Frame.FragmentIndex < b.Frame.FragmentIndex {
					return -1
				} else if a.Frame.FragmentIndex > b.Frame.FragmentIndex {
					return 1
				}

				return 0
			})

			fragments := make([][]byte, 0)
			for _, thisContainer := range containers {
				fragments = append(fragments, thisContainer.Frame.Payload)
			}

			payload, err := fragmentation.Defragment(fragments)
			if err != nil {
				log.Printf("warning: failed to defragment %v fragments for %v: %v", len(fragments), container, err)
			} else {
				container = containers[len(containers)-1]
				container.Frame.Payload = payload

				skip = false
			}

			delete(s.containerByFragmentIndexByCorrelationID, container.Frame.CorrelationID)
		} else {
			s.containerByFragmentIndexByCorrelationID[container.Frame.CorrelationID] = containerByFragmentIndex
		}

		s.mu.Unlock()

		if skip {
			return
		}
	}

	var message *Message

	err := msgpack.Unmarshal(container.Frame.Payload, &message)
	if err != nil {
		log.Printf("warning: attempt to unmarshal returned %#+v for %#+v from %#+v", err, string(container.Frame.Payload), container.ReceivedFrom)
	}

	s.handleInternalReceive(message)
}

// be sure you're holding the mutex before calling this
func (s *Subscriber) subscribe(
	topicName string,
	topicType string,
	onReceive func(*Message),
) error {
	subscription, ok := s.subscriptionByTopicName[topicName]
	if !ok {
		subscription := NewSubscription(
			s.endpointID,
			s.endpointName,
			topicName,
			topicType,
			s.transportManager,
			onReceive,
		)
		subscription.Start()
		s.subscriptionByTopicName[topicName] = subscription
	} else {
		subscription.OnReceive(onReceive)
	}

	return nil
}

func (s *Subscriber) Subscribe(
	topicName string,
	topicType string,
	onReceive func(*Message),
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.subscribe(
		topicName,
		topicType,
		onReceive,
	)
}

func (s *Subscriber) Unsubscribe(
	topicName string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscription, ok := s.subscriptionByTopicName[topicName]
	if !ok {
		return nil
	}

	subscription.Stop()

	s.subscriptionByTopicName[topicName] = nil

	return nil
}

func (s *Subscriber) Start() {
	// noop
}

func (s *Subscriber) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, subscription := range s.subscriptionByTopicName {
		subscription.Stop()
	}
}
