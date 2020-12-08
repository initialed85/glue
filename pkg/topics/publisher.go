package topics

import (
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/transport"
)

type Publisher struct {
	mu                     sync.Mutex
	publicationByTopicName map[string]*Publication
	endpointID             ksuid.KSUID
	endpointName           string
	transportManager       *transport.Manager
	subscriber             **Subscriber
}

func NewPublisher(
	endpointID ksuid.KSUID,
	endpointName string,
	transportManager *transport.Manager,
	subscriber **Subscriber,
) *Publisher {
	p := Publisher{
		publicationByTopicName: make(map[string]*Publication),
		endpointID:             endpointID,
		endpointName:           endpointName,
		transportManager:       transportManager,
		subscriber:             subscriber,
	}

	return &p
}

func (p *Publisher) Publish(
	topicName string,
	topicType string,
	expiry time.Duration,
	payload []byte,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	publication, ok := p.publicationByTopicName[topicName]
	if !ok {
		publication = NewPublication(
			p.endpointID,
			p.endpointName,
			topicName,
			topicType,
			p.transportManager,
			p.subscriber,
		)
		publication.Start()
		p.publicationByTopicName[topicName] = publication
	} else {
		if publication.TopicType() != topicType {
			return fmt.Errorf("publication for topic %#+v already exists with type %#+v; cannot publish with type %#+v",
				publication.topicName,
				publication.topicType,
				topicType,
			)
		}
	}

	return publication.Publish(
		expiry,
		payload,
	)
}

func (p *Publisher) Start() {
	// noop
}

func (p *Publisher) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, publication := range p.publicationByTopicName {
		publication.Stop()
	}
}
