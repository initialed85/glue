package topics

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/transport"
	"github.com/initialed85/glue/pkg/worker"
)

// TODO: revise? configurable?
const MessageTimeout = time.Second * 1
const MessageExpiry = time.Second * 5

type Publication struct {
	scheduleWorker             *worker.ScheduledWorker
	mu                         sync.Mutex
	messageByMessageIdentifier map[MessageIdentifier]Message
	sequenceNumber             int64
	endpointID                 ksuid.KSUID
	endpointName               string
	topicName                  string
	topicType                  string
	transportManager           *transport.Manager
}

func NewPublication(
	endpointID ksuid.KSUID,
	endpointName string,
	topicName string,
	topicType string,
	transportManager *transport.Manager,

) *Publication {
	p := Publication{
		messageByMessageIdentifier: make(map[MessageIdentifier]Message),
		sequenceNumber:             1,
		endpointID:                 endpointID,
		endpointName:               endpointName,
		topicName:                  topicName,
		topicType:                  topicType,
		transportManager:           transportManager,
	}

	p.scheduleWorker = worker.NewScheduledWorker(
		func() {},
		p.work,
		func() {},
		scheduledWorkerRate,
	)

	return &p
}

func (p *Publication) work() {
	now := time.Now()

	toRemove := make([]MessageIdentifier, 0)

	p.mu.Lock()

	for messageIdentifier, message := range p.messageByMessageIdentifier {
		expireTimestamp := message.Timestamp.Add(message.Expiry)
		if now.Before(expireTimestamp) {
			continue
		}

		toRemove = append(toRemove, messageIdentifier)
	}

	for _, messageIdentifier := range toRemove {
		delete(p.messageByMessageIdentifier, messageIdentifier)
	}

	p.mu.Unlock()
}

func (p *Publication) Publish(
	expiry time.Duration,
	payload []byte,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	message := Message{
		Timestamp:      time.Now(),
		Expiry:         expiry,
		SequenceNumber: p.sequenceNumber,
		TopicName:      p.topicName,
		TopicType:      p.topicType,
		MessageType:    StandardMessageType,
		Payload:        payload,
	}

	// TODO: proper serialization
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	p.transportManager.Broadcast(
		MessageTimeout,
		MessageExpiry,
		ksuid.New(),
		1, // TODO: implement fragmentation
		0, // TODO: implement fragmentation
		true,
		payload,
	)

	p.messageByMessageIdentifier[MessageIdentifier{
		EndpointID:     p.endpointID,
		SequenceNumber: p.sequenceNumber,
	}] = message

	p.sequenceNumber += 1

	return nil
}

func (p *Publication) Start() {
	p.scheduleWorker.Start()
	log.Printf("publication started: name=%#+v, type=%#+v", p.topicName, p.topicType)
}

func (p *Publication) Stop() {
	p.scheduleWorker.Stop()
	log.Printf("publication stopped: name=%#+v, type=%#+v", p.topicName, p.topicType)
}
