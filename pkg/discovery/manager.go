package discovery

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/serialization"
	"github.com/initialed85/glue/pkg/types"
	"github.com/initialed85/glue/pkg/worker"
)

const scheduledWorkerRate = time.Second * 1

type Manager struct {
	scheduledWorker                       *worker.ScheduledWorker
	announcer                             *Announcer
	listener                              *Listener
	mu                                    sync.Mutex
	lastAnnouncementContainerByEndpointID map[ksuid.KSUID]*types.Container
	networkID                             int64
	endpointID                            ksuid.KSUID
	endpointName                          string
	listenAddress                         *net.UDPAddr
	discoveryListenAddress                *net.UDPAddr
	discoveryTargetAddress                *net.UDPAddr
	interfaceName                         string
	rate                                  time.Duration
	rateTimeoutMultiplier                 float64
	networkManager                        *network.Manager
	onAdded                               func(*types.Container)
	onRemoved                             func(*types.Container)
}

func NewManager(
	networkID int64,
	endpointID ksuid.KSUID,
	endpointName string,
	listenAddress *net.UDPAddr,
	discoveryListenAddress *net.UDPAddr,
	discoveryTargetAddress *net.UDPAddr,
	interfaceName string,
	rate time.Duration,
	rateTimeoutMultiplier float64,
	networkManager *network.Manager,
	onAdded func(*types.Container),
	onRemoved func(*types.Container),
) *Manager {
	m := Manager{
		lastAnnouncementContainerByEndpointID: make(map[ksuid.KSUID]*types.Container),
		networkID:                             networkID,
		endpointID:                            endpointID,
		endpointName:                          endpointName,
		listenAddress:                         listenAddress,
		discoveryListenAddress:                discoveryListenAddress,
		discoveryTargetAddress:                discoveryTargetAddress,
		interfaceName:                         interfaceName,
		rate:                                  rate,
		rateTimeoutMultiplier:                 rateTimeoutMultiplier,
		networkManager:                        networkManager,
		onAdded:                               onAdded,
		onRemoved:                             onRemoved,
	}

	m.scheduledWorker = worker.NewScheduledWorker(
		func() {},
		m.work,
		func() {},
		scheduledWorkerRate,
	)

	m.announcer = NewAnnouncer(
		m.networkID,
		m.endpointID,
		m.endpointName,
		m.listenAddress,
		m.discoveryListenAddress,
		m.discoveryTargetAddress,
		m.interfaceName,
		m.rate,
		m.networkManager,
		m.onSend,
	)

	m.listener = NewListener(
		m.networkID,
		m.discoveryListenAddress,
		m.interfaceName,
		m.networkManager,
		m.onReceive,
	)

	return &m
}

func (m *Manager) work() {
	now := time.Now()

	toRemove := make([]*types.Container, 0)

	m.mu.Lock()

	for _, container := range m.lastAnnouncementContainerByEndpointID {
		expireDuration := time.Millisecond * time.Duration(float64(container.Announcement.SentRate.Milliseconds())*m.rateTimeoutMultiplier)
		expireTimestamp := container.ReceivedTimestamp.Add(expireDuration)
		if now.Before(expireTimestamp) {
			continue
		}

		toRemove = append(toRemove, container)
	}

	for _, container := range toRemove {
		delete(m.lastAnnouncementContainerByEndpointID, container.SourceEndpointID)
	}

	for _, container := range toRemove {
		log.Printf("removed: %v", container.String())

		// TODO: fix unbounded goroutine use
		go m.onRemoved(container)
	}

	m.mu.Unlock()
}

func (m *Manager) onSend(container *types.Container) {
	_ = container // noop
}

func (m *Manager) onReceive(container *types.Container) {
	if container.SourceEndpointID != m.endpointID {
		if container.SourceEndpointName == m.endpointName {
			log.Printf("warning: ignoring %v because of EndpointName clash with us (%v)", container.String(), m.endpointName)
			return
		}

		possiblyClashingAnnouncementContainer, err := m.GetLastAnnouncementContainerByEndpointName(container.SourceEndpointName)
		if err == nil {
			if possiblyClashingAnnouncementContainer.SourceEndpointID != container.SourceEndpointID {
				log.Printf("warning: ignoring %v because of EndpointName clash with existing endpoint %v",
					container.String(),
					possiblyClashingAnnouncementContainer.String(),
				)
				return
			}
		}
	}

	m.mu.Lock()
	_, endpointExists := m.lastAnnouncementContainerByEndpointID[container.SourceEndpointID]
	m.lastAnnouncementContainerByEndpointID[container.SourceEndpointID] = container
	m.mu.Unlock()

	if !endpointExists && container.SourceEndpointID != m.endpointID {
		log.Printf("added: %v", container.String())

		// TODO: maybe some sort of background worker pool vs unbounded amount of goroutines
		go m.onAdded(container)
	}

	// an announcement sent directly to us requires us to flush all our known announcements back to it
	if container.Announcement.DiscoveryTargetAddr != nil &&
		!container.Announcement.DiscoveryTargetAddr.IP.IsMulticast() &&
		container.Announcement.DiscoveryListenAddr != nil {

		for _, otherContainer := range m.GetAllAnnouncementContainers(true) {
			if otherContainer.SourceEndpointID == container.SourceEndpointID {
				continue
			}

			if otherContainer.Announcement.Forwarded {
				continue
			}

			copiedOtherContainer := otherContainer.Copy()
			copiedOtherContainer.Announcement.Forwarded = true

			data, err := serialization.Serialize(copiedOtherContainer)
			if err != nil {
				log.Printf("warning: %v", err)
				continue
			}

			err = m.networkManager.Send(container.Announcement.DiscoveryListenAddr, data)
			if err != nil {
				log.Printf("warning: %v", err)
				continue
			}
		}
	}
}

func (m *Manager) GetLastAnnouncementContainerByEndpointName(endpointName string) (*types.Container, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, container := range m.lastAnnouncementContainerByEndpointID {
		if container.SourceEndpointName != endpointName {
			continue
		}

		return container, nil
	}

	return nil, fmt.Errorf("no announcements for endpointName %#v", endpointName)
}

func (m *Manager) GetAllAnnouncementContainers(includeSelf ...bool) []*types.Container {
	m.mu.Lock()
	defer m.mu.Unlock()

	actualIncludeSelf := len(includeSelf) > 1 && includeSelf[0]

	containers := make([]*types.Container, 0)

	for _, container := range m.lastAnnouncementContainerByEndpointID {
		if container.SourceEndpointID == m.endpointID && !actualIncludeSelf {
			continue
		}

		containers = append(containers, container)
	}

	return containers
}

func (m *Manager) Start() {
	m.scheduledWorker.Start()
	m.listener.Start()
	m.announcer.Start()
}

func (m *Manager) Stop() {
	m.scheduledWorker.Stop()
	m.listener.Stop()
	m.announcer.Stop()
}
