package discovery

import (
	"log"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/serialization"
	"github.com/initialed85/glue/pkg/types"
	"github.com/initialed85/glue/pkg/worker"
)

type Announcer struct {
	scheduledWorker  *worker.ScheduledWorker
	networkID        int64
	endpointID       ksuid.KSUID
	endpointName     string
	listenPort       int
	multicastAddress string
	interfaceName    string
	rate             time.Duration
	networkManager   *network.Manager
	onSend           func(types.Container)
}

func NewAnnouncer(
	networkID int64,
	endpointID ksuid.KSUID,
	endpointName string,
	listenPort int,
	multicastAddress string,
	interfaceName string,
	rate time.Duration,
	networkManager *network.Manager,
	onSend func(types.Container),
) *Announcer {
	a := Announcer{
		networkID:        networkID,
		endpointID:       endpointID,
		endpointName:     endpointName,
		listenPort:       listenPort,
		multicastAddress: multicastAddress,
		interfaceName:    interfaceName,
		rate:             rate,
		networkManager:   networkManager,
		onSend:           onSend,
	}

	a.scheduledWorker = worker.NewScheduledWorker(
		func() {},
		a.work,
		func() {},
		rate,
	)

	return &a
}

func (a *Announcer) work() {
	rawSrcAddr, err := a.networkManager.GetRawSrcAddr(a.multicastAddress)
	if err != nil {
		log.Printf("warning: %v", err)
		return
	}

	container := types.GetAnnouncementContainer(
		time.Now(),
		rawSrcAddr,
		a.networkID,
		a.endpointID,
		a.endpointName,
		a.rate,
		a.listenPort,
	)

	data, err := serialization.Serialize(container)
	if err != nil {
		log.Printf("warning: %v", err)
		return
	}

	err = a.networkManager.Send(a.multicastAddress, data)
	if err != nil {
		log.Printf("warning: %v", err)
		return
	}
}

func (a *Announcer) Start() {
	a.scheduledWorker.Start()
}

func (a *Announcer) Stop() {
	a.scheduledWorker.Stop()
}
