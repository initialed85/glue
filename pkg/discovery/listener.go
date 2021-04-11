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
	networkID        int64
	multicastAddress string
	interfaceName    string
	networkManager   *network.Manager
	onReceive        func(types.Container)
}

func NewListener(
	networkID int64,
	multicastAddress string,
	interfaceName string,
	networkManager *network.Manager,
	onReceive func(types.Container),
) *Listener {
	return &Listener{
		networkID:        networkID,
		multicastAddress: multicastAddress,
		interfaceName:    interfaceName,
		networkManager:   networkManager,
		onReceive:        onReceive,
	}
}

func (l *Listener) callback(
	srcAddr *net.UDPAddr,
	data []byte,
) {
	receivedTimestamp := time.Now()

	container, err := serialization.Deserialize(data)
	if err != nil {
		log.Printf("error: attempt to deserialize %#+v failed stating: %v", data, err)
		return
	}

	if container.NetworkID != l.networkID {
		log.Printf("error: expected networkID %#+v but got %#+v", l.networkID, container.NetworkID)
		return
	}

	container.ReceivedTimestamp = receivedTimestamp
	container.ReceivedAddress = srcAddr.String()

	// TODO: maybe some sort of background worker pool vs unbounded amount of goroutines
	go l.onReceive(container)
}

func (l *Listener) Start() {
	err := l.networkManager.RegisterCallback(l.multicastAddress, l.interfaceName, l.callback)
	if err != nil {
		log.Printf("warning: attempt to register callback failed stating: %v", err)
	}
}

func (l *Listener) Stop() {
	err := l.networkManager.UnregisterCallback(l.multicastAddress, l.interfaceName, l.callback)
	if err != nil {
		log.Printf("warning: attempt to unregister callback failed stating: %v", err)
	}
}
