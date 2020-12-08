package topics

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/transport"
	"github.com/initialed85/glue/pkg/types"
)

type Subscriber struct {
	mu                      sync.Mutex
	subscriptionByTopicName map[string]*Subscription
	endpointID              ksuid.KSUID
	endpointName            string
	transportManager        *transport.Manager
	publisher               *Publisher
}

func NewSubscriber(
	endpointID ksuid.KSUID,
	endpointName string,
	transportManager *transport.Manager,
	publisher *Publisher,
) *Subscriber {
	s := Subscriber{
		subscriptionByTopicName: make(map[string]*Subscription),
		endpointID:              endpointID,
		endpointName:            endpointName,
		transportManager:        transportManager,
		publisher:               publisher,
	}

	return &s
}

func (s *Subscriber) HandleReceive(container types.Container) {
	var message Message

	err := json.Unmarshal(container.Frame.Payload, &message)
	if err != nil {
		log.Printf("warning: attempt to unmarshal returned %#+v for %#+v from %#+v", err, string(container.Frame.Payload), container.ReceivedAddress)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	subscription, ok := s.subscriptionByTopicName[message.TopicName]
	if !ok {
		// TODO: flesh this out- it's for unrelated late joiners
		err = s.subscribe(
			message.TopicName,
			message.TopicType,
			func(Message) {
				// noop
			},
		)
		return
	}

	if subscription.topicType != message.TopicType {
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

// be sure you're holding the mutex before calling this
func (s *Subscriber) subscribe(
	topicName string,
	topicType string,
	onReceive func(Message),
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
		// TODO: accessing property of other struct
		subscription.onReceive = onReceive
	}

	return nil
}

func (s *Subscriber) Subscribe(
	topicName string,
	topicType string,
	onReceive func(Message),
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.subscribe(
		topicName,
		topicType,
		onReceive,
	)
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
