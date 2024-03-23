package transport

import (
	"log"
	"sync"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/discovery"
	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/serialization"
	"github.com/initialed85/glue/pkg/types"
	"github.com/initialed85/glue/pkg/worker"
)

const scheduledWorkerRate = time.Second * 1

type Sender struct {
	scheduledWorker        *worker.ScheduledWorker
	mu                     sync.Mutex
	sentContainerByFrameID map[ksuid.KSUID]*types.Container
	networkID              int64
	endpointID             ksuid.KSUID
	endpointName           string
	discoveryManager       *discovery.Manager
	networkManager         *network.Manager
}

func NewSender(
	networkID int64,
	endpointID ksuid.KSUID,
	endpointName string,
	discoveryManager *discovery.Manager,
	networkManager *network.Manager,
) *Sender {
	s := Sender{
		sentContainerByFrameID: make(map[ksuid.KSUID]*types.Container),
		networkID:              networkID,
		endpointID:             endpointID,
		endpointName:           endpointName,
		discoveryManager:       discoveryManager,
		networkManager:         networkManager,
	}

	s.scheduledWorker = worker.NewScheduledWorker(
		func() {},
		s.work,
		func() {},
		scheduledWorkerRate,
	)

	return &s
}

func (s *Sender) work() {
	now := time.Now()

	toResend := make([]*types.Container, 0)
	toDelete := make([]*types.Container, 0)

	s.mu.Lock()

	for _, frame := range s.sentContainerByFrameID {
		if now.After(frame.SentTimestamp.Add(frame.Frame.ResendExpiry)) {
			toDelete = append(toDelete, frame)
		} else if now.After(frame.LastSentTimestamp.Add(frame.Frame.ResendPeriod)) {
			toResend = append(toResend, frame)
		}
	}

	for _, container := range toDelete {
		delete(s.sentContainerByFrameID, container.Frame.FrameID)
	}

	s.mu.Unlock()

	var err error

	for _, container := range toResend {
		// TODO: backoff multiplier if we get failures here?
		err = s.send(container, true, true)
		if err != nil {
			log.Printf("warning: failed resend of frame trying to send for %v because of %v", container.String(), err)
			continue
		}
	}
}

func (s *Sender) send(container *types.Container, permitResend bool, isResend bool) error {
	announcementContainer, err := s.discoveryManager.GetLastAnnouncementContainerByEndpointName(container.Frame.DestinationEndpointName)
	if err != nil {
		return err
	}

	rawSrcAddr, err := s.networkManager.GetRawSrcAddr(announcementContainer.Announcement.ListenAddr)
	if err != nil {
		return err
	}

	container.SentBy = rawSrcAddr.String()
	container.SentTo = announcementContainer.Announcement.ListenAddr.String()

	now := time.Now()

	if !isResend {
		container.SentTimestamp = now
		container.LastSentTimestamp = now
	} else {
		container.LastSentTimestamp = now
	}

	if permitResend {
		s.mu.Lock()
		s.sentContainerByFrameID[container.Frame.FrameID] = container
		s.mu.Unlock()
	}

	data, err := serialization.Serialize(container)
	if err != nil {
		return err
	}

	// attemptType := "sent"
	// if isResend {
	// 	attemptType = "resent"
	// }

	// frameType := "data"
	// if container.Frame.IsAck {
	// 	frameType = "ack"
	// }

	// log.Printf("%v -> %v; %v %v frame to %v", attemptType, frameType, rawSrcAddr.String(), announcementContainer.Announcement.ListenAddr.String(), container.SourceEndpointName)

	return s.networkManager.Send(announcementContainer.Announcement.ListenAddr, data)
}

func (s *Sender) Send(
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
	frame := types.GetFrameContainer(
		resendPeriod,
		resendExpiry,
		s.networkID,
		s.endpointID,
		s.endpointName,
		correlationID,
		fragmentCount,
		fragmentIndex,
		destinationEndpointID,
		destinationEndpointName,
		needsAck,
		isAck,
		payload,
	)

	return s.send(frame, needsAck && !isAck, false)
}

func (s *Sender) Broadcast(
	resendTimeout time.Duration,
	resendExpiry time.Duration,
	correlationID ksuid.KSUID,
	fragmentCount int64,
	fragmentIndex int64,
	needsAck bool,
	payload []byte,
) {
	var err error
	for _, container := range s.discoveryManager.GetAllAnnouncementContainers() {
		if container.SourceEndpointID == s.endpointID {
			continue
		}

		err = s.Send(
			resendTimeout,
			resendExpiry,
			correlationID,
			fragmentCount,
			fragmentIndex,
			container.SourceEndpointID,
			container.SourceEndpointName,
			needsAck,
			false, // doesn't make sense to broadcast an ack
			payload,
		)
		if err != nil {
			log.Printf("warning: failed to broadcast %v bytes because %v for %v", len(payload), err, container.String())
		}
	}
}

func (s *Sender) SendAck(container *types.Container) error {
	ackContainer := types.GetFrameAckContainer(
		s.networkID,
		s.endpointID,
		s.endpointName,
		container,
	)

	return s.send(ackContainer, false, false)
}

func (s *Sender) MarkAck(container *types.Container) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.sentContainerByFrameID[container.Frame.FrameID]
	if !ok {
		log.Printf("warning: failed marking messge as ack'd because it's unknown: %v", container.String())
		return
	}

	delete(s.sentContainerByFrameID, container.Frame.FrameID)
}

func (s *Sender) Start() {
	s.scheduledWorker.Start()
}

func (s *Sender) Stop() {
	s.scheduledWorker.Stop()
}
