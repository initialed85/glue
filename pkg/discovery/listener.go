package discovery

import (
	"log"
	"net"
	"time"

	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/serialization"
	"github.com/initialed85/glue/pkg/types"
)

type Listener struct {
	networkID              int64
	discoveryListenAddress *net.UDPAddr
	interfaceName          string
	networkManager         *network.Manager
	onReceive              func(*types.Container)
}

func NewListener(
	networkID int64,
	discoveryListenAddress *net.UDPAddr,
	interfaceName string,
	networkManager *network.Manager,
	onReceive func(*types.Container),
) *Listener {
	return &Listener{
		networkID:              networkID,
		discoveryListenAddress: discoveryListenAddress,
		interfaceName:          interfaceName,
		networkManager:         networkManager,
		onReceive:              onReceive,
	}
}

func (l *Listener) callback(
	srcAddr *net.UDPAddr,
	dstAddr *net.UDPAddr,
	data []byte,
) {
	receivedTimestamp := time.Now()

	container, err := serialization.Deserialize(data)
	if err != nil {
		log.Printf("error: attempt to deserialize %#+v failed stating: %v", data, err)
		return
	}

	if container.NetworkID != l.networkID {
		log.Printf("error: expected networkID %#+v but got %#+v for %v", l.networkID, container.NetworkID, container.String())
		return
	}

	if container.Announcement == nil {
		log.Printf("error: unexpectedly received non-announcement %#+v", container)
		return
	}

	// log.Printf("%v -> %v; receive announcement for %v", srcAddr.String(), dstAddr, container.SourceEndpointName)

	container.ReceivedTimestamp = receivedTimestamp
	container.ReceivedFrom = srcAddr.String()
	container.ReceivedBy = dstAddr.String()

	// TODO: maybe some sort of background worker pool vs unbounded amount of goroutines
	go l.onReceive(container)
}

func (l *Listener) Start() {
	err := l.networkManager.RegisterCallback(l.discoveryListenAddress, l.interfaceName, l.callback)
	if err != nil {
		log.Printf("warning: attempt to register callback failed stating: %v", err)
	}
}

func (l *Listener) Stop() {
	err := l.networkManager.UnregisterCallback(l.discoveryListenAddress, l.interfaceName, l.callback)
	if err != nil {
		log.Printf("warning: attempt to unregister callback failed stating: %v", err)
	}
}
