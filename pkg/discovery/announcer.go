package discovery

import (
	"log"
	"net"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/serialization"
	"github.com/initialed85/glue/pkg/types"
	"github.com/initialed85/glue/pkg/worker"
)

type Announcer struct {
	scheduledWorker        *worker.ScheduledWorker
	networkID              int64
	endpointID             ksuid.KSUID
	endpointName           string
	listenAddress          *net.UDPAddr
	discoveryListenAddress *net.UDPAddr
	discoveryTargetAddress *net.UDPAddr
	interfaceName          string
	rate                   time.Duration
	networkManager         *network.Manager
	onSend                 func(*types.Container)
}

func NewAnnouncer(
	networkID int64,
	endpointID ksuid.KSUID,
	endpointName string,
	listenAddress *net.UDPAddr,
	discoveryListenAddress *net.UDPAddr,
	discoveryTargetAddress *net.UDPAddr,
	interfaceName string,
	rate time.Duration,
	networkManager *network.Manager,
	onSend func(*types.Container),
) *Announcer {
	a := Announcer{
		networkID:              networkID,
		endpointID:             endpointID,
		endpointName:           endpointName,
		listenAddress:          listenAddress,
		discoveryListenAddress: discoveryListenAddress,
		discoveryTargetAddress: discoveryTargetAddress,
		interfaceName:          interfaceName,
		rate:                   rate,
		networkManager:         networkManager,
		onSend:                 onSend,
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
	if a.discoveryTargetAddress == nil {
		return
	}

	srcAddr, err := a.networkManager.GetRawSrcAddr(a.discoveryTargetAddress)
	if err != nil {
		log.Printf("warning: announcer failed to get src addr for %v: %v", a.discoveryTargetAddress.String(), err)
		return
	}

	discoveryListenAddr := &net.UDPAddr{
		IP:   srcAddr.IP,
		Port: a.discoveryListenAddress.Port,
		Zone: srcAddr.Zone,
	}

	listenAddr := &net.UDPAddr{
		IP:   srcAddr.IP,
		Port: a.listenAddress.Port,
		Zone: srcAddr.Zone,
	}

	container := types.GetAnnouncementContainer(
		time.Now(),
		srcAddr.String(),
		a.networkID,
		a.endpointID,
		a.endpointName,
		a.rate,
		discoveryListenAddr,
		a.discoveryTargetAddress,
		listenAddr,
	)

	container.SentTo = a.discoveryTargetAddress.String()

	data, err := serialization.Serialize(container)
	if err != nil {
		log.Printf("warning: %v", err)
		return
	}

	// log.Printf("%v -> %v; send announcement for %v", srcAddr.String(), a.discoveryTargetAddress.String(), container.SourceEndpointName)

	err = a.networkManager.Send(a.discoveryTargetAddress, data)
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
