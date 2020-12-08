package topics

import (
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/transport"
	"github.com/initialed85/glue/pkg/types"
)

type Manager struct {
	publisher        *Publisher
	subscriber       *Subscriber
	endpointID       ksuid.KSUID
	endpointName     string
	transportManager *transport.Manager
}

func NewManager(
	endpointID ksuid.KSUID,
	endpointName string,
	transportManager *transport.Manager,
) *Manager {
	m := Manager{
		endpointID:       endpointID,
		endpointName:     endpointName,
		transportManager: transportManager,
	}

	m.publisher = NewPublisher(
		m.endpointID,
		m.endpointName,
		m.transportManager,
	)

	m.subscriber = NewSubscriber(
		m.endpointID,
		m.endpointName,
		m.transportManager,
		m.publisher,
	)

	return &m
}

func (m *Manager) HandleReceive(container types.Container) {
	m.subscriber.HandleReceive(container)
}

func (m *Manager) Publish(
	topicName string,
	topicType string,
	expiry time.Duration,
	payload []byte,
) error {
	return m.publisher.Publish(
		topicName,
		topicType,
		expiry,
		payload,
	)
}

func (m *Manager) Subscribe(
	topicName string,
	topicType string,
	onReceive func(Message),
) error {
	return m.subscriber.Subscribe(
		topicName,
		topicType,
		onReceive,
	)
}

func (m *Manager) Start() {
	m.publisher.Start()
	m.subscriber.Start()
}

func (m *Manager) Stop() {
	m.subscriber.Stop()
	m.publisher.Stop()
}
