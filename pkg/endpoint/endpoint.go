package endpoint

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/discovery"
	"github.com/initialed85/glue/pkg/helpers"
	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/topics"
	"github.com/initialed85/glue/pkg/transport"
	"github.com/initialed85/glue/pkg/types"
)

type Manager struct {
	networkID                      int64
	endpointID                     ksuid.KSUID
	endpointName                   string
	listenAddress                  *net.UDPAddr
	discoveryListenAddress         *net.UDPAddr
	discoveryTargetAddress         *net.UDPAddr
	listenInterface                string
	discoveryRate                  time.Duration
	discoveryRateTimeoutMultiplier float64
	onAdded                        func(*types.Container)
	onRemoved                      func(*types.Container)
	networkManager                 *network.Manager
	discoveryManager               *discovery.Manager
	transportManager               *transport.Manager
	topicsManager                  *topics.Manager
}

func NewManager(
	networkID int64,
	endpointID ksuid.KSUID,
	endpointName string,
	listenAddress *net.UDPAddr,
	discoveryListenAddress *net.UDPAddr,
	discoveryTargetAddress *net.UDPAddr,
	listenInterface string,
	discoveryRate time.Duration,
	discoveryRateTimeoutMultiplier float64,
	onAdded func(*types.Container),
	onRemoved func(*types.Container),
) *Manager {
	log.Printf("endpoint; networkID: %v", networkID)
	log.Printf("endpoint; endpointID: %v", endpointID)
	log.Printf("endpoint; endpointName: %v", endpointName)
	log.Printf("endpoint; listenAddress: %v", listenAddress)
	log.Printf("endpoint; discoveryListenAddress: %v", discoveryListenAddress)
	log.Printf("endpoint; discoveryTargetAddress: %v", discoveryTargetAddress)
	log.Printf("endpoint; listenInterface: %v", listenInterface)
	log.Printf("endpoint; discoveryRate: %v", discoveryRate)
	log.Printf("endpoint; discoveryRateTimeoutMultiplier: %v", discoveryRateTimeoutMultiplier)

	m := Manager{
		networkID:                      networkID,
		endpointID:                     endpointID,
		endpointName:                   endpointName,
		listenAddress:                  listenAddress,
		discoveryListenAddress:         discoveryListenAddress,
		discoveryTargetAddress:         discoveryTargetAddress,
		listenInterface:                listenInterface,
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
		listenAddress,
		discoveryListenAddress,
		discoveryTargetAddress,
		listenInterface,
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
		listenAddress,
		listenInterface,
		m.discoveryManager,
		m.networkManager,
		func(container *types.Container) {
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
	networkID, err := helpers.GetNetworkIDFromEnv()
	if err != nil {
		networkID = 1
	}

	endpointID, err := helpers.GetEndpointIDFromEnv()
	if err != nil {
		endpointID = ksuid.New()
	}

	endpointName, err := helpers.GetEndpointNameFromEnv()
	if err != nil {
		endpointName = fmt.Sprintf("Endpoint_%v", endpointID)
	}

	listenAddress, err := helpers.GetListenAddressFromEnv()
	if err != nil {
		listenPort, err := network.GetFreePort()
		if err != nil {
			return nil, err
		}

		rawListenAddress := fmt.Sprintf("0.0.0.0:%v", listenPort)

		listenAddress, err = net.ResolveUDPAddr(network.GetNetwork(rawListenAddress), rawListenAddress)
		if err != nil {
			return nil, err
		}
	}

	discoveryTargetAddress, err := helpers.GetDiscoveryTargetAddressFromEnv()
	if err != nil {
		discoveryTargetAddress, _ = net.ResolveUDPAddr("udp4", "239.192.137.1:27320")
	}

	discoveryListenAddress, err := helpers.GetDiscoveryListenAddressFromEnv()
	if err != nil {
		discoveryListenAddress, _ = net.ResolveUDPAddr("udp4", "239.192.137.1:27320")
	}

	listenInterface, err := helpers.GetListenInterfaceFromEnv()
	if err != nil {
		listenInterface, err = network.GetDefaultInterfaceName()
		if err != nil {
			return nil, err
		}
	}

	discoveryRate, err := helpers.GetDiscoveryRateFromEnv()
	if err != nil {
		discoveryRate = time.Second * 1
	}

	discoveryRateTimeoutMultiplier, err := helpers.GetDiscoveryRateTimeoutMultiplierFromEnv()
	if err != nil {
		discoveryRateTimeoutMultiplier = 2.0
	}

	return NewManager(
		networkID,
		endpointID,
		endpointName,
		listenAddress,
		discoveryListenAddress,
		discoveryTargetAddress,
		listenInterface,
		discoveryRate,
		discoveryRateTimeoutMultiplier,
		func(container *types.Container) {},
		func(container *types.Container) {},
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
	onReceive func(*topics.Message),
) error {
	return m.topicsManager.Subscribe(
		topicName,
		topicType,
		onReceive,
	)
}

func (m *Manager) Unsubscribe(
	topicName string,
) error {
	return m.topicsManager.Unsubscribe(topicName)
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
