package discovery

import (
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/types"
	"github.com/initialed85/glue/pkg/worker"
	"log"
	"sync"
	"time"
)

const scheduledWorkerRate = time.Millisecond * 100

type Manager struct {
	scheduledWorker                       *worker.ScheduledWorker
	announcer                             *Announcer
	listener                              *Listener
	mu                                    sync.Mutex
	lastAnnouncementContainerByEndpointID map[ksuid.KSUID]types.Container
	networkID                             int64
	endpointID                            ksuid.KSUID
	endpointName                          string
	listenPort                            int
	multicastAddress                      string
	interfaceName                         string
	rate                                  time.Duration
	rateTimeoutMultiplier                 float64
	networkManager                        *network.Manager
	onAdded                               func(types.Container)
	onRemoved                             func(types.Container)
}

func NewManager(
	networkID int64,
	endpointID ksuid.KSUID,
	endpointName string,
	listenPort int,
	multicastAddress string,
	interfaceName string,
	rate time.Duration,
	rateTimeoutMultiplier float64,
	networkManager *network.Manager,
	onAdded func(types.Container),
	onRemoved func(types.Container),
) *Manager {
	m := Manager{
		lastAnnouncementContainerByEndpointID: make(map[ksuid.KSUID]types.Container),
		networkID:                             networkID,
		endpointID:                            endpointID,
		endpointName:                          endpointName,
		listenPort:                            listenPort,
		multicastAddress:                      multicastAddress,
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
		m.listenPort,
		m.multicastAddress,
		m.interfaceName,
		m.rate,
		m.networkManager,
		m.onSend,
	)

	m.listener = NewListener(
		m.networkID,
		m.multicastAddress,
		m.interfaceName,
		m.networkManager,
		m.onReceive,
	)

	return &m
}

func (m *Manager) work() {
	now := time.Now()

	toRemove := make([]types.Container, 0)

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
		log.Printf(
			"removed: endpointID=%#+v, endpointName=%#+v, receivedAddress=%#+v, listenPort=%#+v",
			container.SourceEndpointID.String(),
			container.SourceEndpointName,
			container.ReceivedAddress,
			container.Announcement.ListenPort,
		)

		// TODO: fix unbounded goroutine use
		go m.onRemoved(container)
	}

	m.mu.Unlock()
}

func (m *Manager) onSend(container types.Container) {
	_ = container // noop
}

func (m *Manager) onReceive(container types.Container) {
	if container.SourceEndpointID == m.endpointID {
		return // ignore own announcements
	}

	if container.SourceEndpointName == m.endpointName {
		log.Printf("warning: ignoring annoumcement because of clash with us: %#+v", container)
		return // warn and ignore if somebody is pretending to us
	}

	possiblyClashingAnnouncementContainer, err := m.GetLastAnnouncementContainerByEndpointName(container.SourceEndpointName)
	if err == nil {
		if possiblyClashingAnnouncementContainer.SourceEndpointID != container.SourceEndpointID {
			log.Printf("warning: ignoring container because of clash with existing endpoint: %#+v", container)
			return // warn and ignore if somebody is pretending to be somebody else
		}
	}

	m.mu.Lock()

	_, endpointExists := m.lastAnnouncementContainerByEndpointID[container.SourceEndpointID]
	m.lastAnnouncementContainerByEndpointID[container.SourceEndpointID] = container

	m.mu.Unlock()

	if !endpointExists {
		log.Printf(
			"added: endpointID=%#+v, endpointName=%#+v, receivedAddress=%#+v, listenPort=%#+v",
			container.SourceEndpointID.String(),
			container.SourceEndpointName,
			container.ReceivedAddress,
			container.Announcement.ListenPort,
		)

		// TODO: maybe some sort of background worker pool vs unbounded amount of goroutines
		go m.onAdded(container)
	}
}

func (m *Manager) GetLastAnnouncementContainerByEndpointName(endpointName string) (types.Container, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, container := range m.lastAnnouncementContainerByEndpointID {
		if container.SourceEndpointName != endpointName {
			continue
		}

		return container, nil
	}

	return types.Container{}, fmt.Errorf("no announcements for endpointName %#v", endpointName)
}

func (m *Manager) GetAllAnnouncementContainers() []types.Container {
	m.mu.Lock()
	defer m.mu.Unlock()

	containers := make([]types.Container, 0)

	for _, container := range m.lastAnnouncementContainerByEndpointID {
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
