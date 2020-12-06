package topics

import (
	"log"
	"sync"
	"time"

	"github.com/segmentio/ksuid"

	"glue/pkg/transport"
	"glue/pkg/worker"
)

type Subscription struct {
	scheduleWorker             *worker.ScheduledWorker
	mu                         sync.Mutex
	messageByMessageIdentifier map[MessageIdentifier]Message
	endpointID                 ksuid.KSUID
	endpointName               string
	topicName                  string
	topicType                  string
	transportManager           *transport.Manager
	onReceive                  func(Message)
}

func NewSubscription(
	endpointID ksuid.KSUID,
	endpointName string,
	topicName string,
	topicType string,
	transportManager *transport.Manager,
	onReceive func(Message),
) *Subscription {
	s := Subscription{
		messageByMessageIdentifier: make(map[MessageIdentifier]Message),
		endpointID:                 endpointID,
		endpointName:               endpointName,
		topicName:                  topicName,
		topicType:                  topicType,
		transportManager:           transportManager,
		onReceive:                  onReceive,
	}

	s.scheduleWorker = worker.NewScheduledWorker(
		func() {},
		s.work,
		func() {},
		scheduledWorkerRate,
	)

	return &s
}

func (s *Subscription) work() {
	now := time.Now()

	toRemove := make([]MessageIdentifier, 0)

	s.mu.Lock()

	for messageIdentifier, message := range s.messageByMessageIdentifier {
		expireTimestamp := message.Timestamp.Add(message.Expiry)
		if now.Before(expireTimestamp) {
			continue
		}

		toRemove = append(toRemove, messageIdentifier)
	}

	for _, messageIdentifier := range toRemove {
		delete(s.messageByMessageIdentifier, messageIdentifier)
	}

	s.mu.Unlock()
}

func (s *Subscription) HandleReceive(message Message) {
	messages := make([]Message, 0)

	messages = append(messages, message)

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, message := range messages {
		if message.MessageType != StandardMessageType && message.MessageType != ForwardedMessageType {
			log.Printf("warning: unknown MessageType %v in %#+v", message.MessageType, message)
			continue
		}

		s.onReceive(message)

		s.messageByMessageIdentifier[MessageIdentifier{
			EndpointID:     message.EndpointID,
			SequenceNumber: message.SequenceNumber,
		}] = message
	}
}

func (s *Subscription) Start() {
	s.scheduleWorker.Start()
	log.Printf("subscription started: name=%#+v, type=%#+v", s.topicName, s.topicType)
}

func (s *Subscription) Stop() {
	s.scheduleWorker.Stop()
	log.Printf("subscription stopped: name=%#+v, type=%#+v", s.topicName, s.topicType)
}
