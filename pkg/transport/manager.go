package transport

import (
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/discovery"
	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/types"
)

type Manager struct {
	sender           *Sender
	receiver         *Receiver
	networkID        int64
	endpointID       ksuid.KSUID
	endpointName     string
	listenPort       int
	interfaceName    string
	discoveryManager *discovery.Manager
	networkManager   *network.Manager
	onReceive        func(types.Container)
}

func NewManager(
	networkID int64,
	endpointID ksuid.KSUID,
	endpointName string,
	listenPort int,
	interfaceName string,
	discoveryManager *discovery.Manager,
	networkManager *network.Manager,
	onReceive func(types.Container),
) *Manager {
	m := Manager{
		networkID:        networkID,
		endpointID:       endpointID,
		endpointName:     endpointName,
		listenPort:       listenPort,
		interfaceName:    interfaceName,
		discoveryManager: discoveryManager,
		networkManager:   networkManager,
		onReceive:        onReceive,
	}

	m.sender = NewSender(
		m.networkID,
		m.endpointID,
		m.endpointName,
		m.discoveryManager,
		m.networkManager,
	)

	m.receiver = NewReceiver(
		m.networkID,
		m.listenPort,
		m.interfaceName,
		m.networkManager,
		m.sender,
		m.onReceive,
	)

	return &m
}

func (m *Manager) Send(
	resendPeriod time.Duration,
	resendExpiry time.Duration,
	correlationID ksuid.KSUID,
	fragmentCount int64,
	fragmentIndex int64,
	destinationEndpointID ksuid.KSUID,
	destinationEndpointName string,
	needsAck bool,
	isAck bool,
	payload []byte,
) error {
	return m.sender.Send(
		resendPeriod,
		resendExpiry,
		correlationID,
		fragmentCount,
		fragmentIndex,
		destinationEndpointID,
		destinationEndpointName,
		needsAck,
		isAck,
		payload,
	)
}

func (m *Manager) Broadcast(
	resendTimeout time.Duration,
	resendExpiry time.Duration,
	correlationID ksuid.KSUID,
	fragmentCount int64,
	fragmentIndex int64,
	needsAck bool,
	payload []byte,
) {
	m.sender.Broadcast(
		resendTimeout,
		resendExpiry,
		correlationID,
		fragmentCount,
		fragmentIndex,
		needsAck,
		payload,
	)
}

func (m *Manager) Start() {
	m.sender.Start()
	m.receiver.Start()
}

func (m *Manager) Stop() {
	m.receiver.Stop()
	m.sender.Stop()
}
