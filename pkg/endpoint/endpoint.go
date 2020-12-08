package endpoint

import (
	"fmt"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/discovery"
	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/topics"
	"github.com/initialed85/glue/pkg/transport"
	"github.com/initialed85/glue/pkg/types"
)

type Manager struct {
	networkID                      int64
	endpointID                     ksuid.KSUID
	endpointName                   string
	listenPort                     int
	multicastAddress               string
	interfaceName                  string
	discoveryRate                  time.Duration
	discoveryRateTimeoutMultiplier float64
	onAdded                        func(types.Container)
	onRemoved                      func(types.Container)
	networkManager                 *network.Manager
	discoveryManager               *discovery.Manager
	transportManager               *transport.Manager
	topicsManager                  *topics.Manager
}

func NewManager(
	networkID int64,
	endpointID ksuid.KSUID,
	endpointName string,
	listenPort int,
	multicastAddress string,
	interfaceName string,
	discoveryRate time.Duration,
	discoveryRateTimeoutMultiplier float64,
	onAdded func(types.Container),
	onRemoved func(types.Container),
) *Manager {
	m := Manager{
		networkID:                      networkID,
		endpointID:                     endpointID,
		endpointName:                   endpointName,
		listenPort:                     listenPort,
		multicastAddress:               multicastAddress,
		interfaceName:                  interfaceName,
		discoveryRate:                  discoveryRate,
		discoveryRateTimeoutMultiplier: discoveryRateTimeoutMultiplier,
		onAdded:                        onAdded,
		onRemoved:                      onRemoved,
		networkManager:                 network.NewManager(),
	}

	m.discoveryManager = discovery.NewManager(
		networkID,
		endpointID,
		endpointName,
		listenPort,
		multicastAddress,
		interfaceName,
		discoveryRate,
		discoveryRateTimeoutMultiplier,
		m.networkManager,
		onAdded,
		onRemoved,
	)

	m.transportManager = transport.NewManager(
		networkID,
		endpointID,
		endpointName,
		listenPort,
		interfaceName,
		m.discoveryManager,
		m.networkManager,
		func(container types.Container) {
			m.topicsManager.HandleReceive(container)
		},
	)

	m.topicsManager = topics.NewManager(
		endpointID,
		endpointName,
		m.transportManager,
	)

	return &m
}

func NewManagerSimple() (*Manager, error) {
	endpointID := ksuid.New()

	port, err := network.GetFreePort()
	if err != nil {
		return nil, err
	}

	defaultInterfaceName, err := network.GetDefaultInterfaceName()
	if err != nil {
		return nil, err
	}

	return NewManager(
		1,
		endpointID,
		fmt.Sprintf("Endpoint_%v", endpointID),
		port,
		"239.192.137.1:27320",
		defaultInterfaceName,
		time.Millisecond*100,
		5,
		func(container types.Container) {},
		func(container types.Container) {},
	), nil
}

func (m *Manager) EndpointID() ksuid.KSUID {
	return m.endpointID
}

func (m *Manager) EndpointName() string {
	return m.endpointName
}

func (m *Manager) Publish(
	topicName string,
	topicType string,
	expiry time.Duration,
	payload []byte,
) error {
	return m.topicsManager.Publish(
		topicName,
		topicType,
		expiry,
		payload,
	)
}

func (m *Manager) Subscribe(
	topicName string,
	topicType string,
	onReceive func(topics.Message),
) error {
	return m.topicsManager.Subscribe(
		topicName,
		topicType,
		onReceive,
	)
}

func (m *Manager) Start() {
	m.networkManager.Start()
	m.discoveryManager.Start()
	m.transportManager.Start()
	m.topicsManager.Start()
}

func (m *Manager) Stop() {
	m.networkManager.Stop()
	m.discoveryManager.Stop()
	m.transportManager.Stop()
	m.topicsManager.Stop()
}
