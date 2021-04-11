package transport

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/serialization"
	"github.com/initialed85/glue/pkg/types"
)

type Receiver struct {
	networkID      int64
	listenPort     int
	interfaceName  string
	networkManager *network.Manager
	sender         *Sender
	onReceive      func(types.Container)
}

func NewReceiver(
	networkID int64,
	listenPort int,
	interfaceName string,
	networkManager *network.Manager,
	sender *Sender,
	onReceive func(types.Container),
) *Receiver {
	r := Receiver{
		networkID:      networkID,
		listenPort:     listenPort,
		interfaceName:  interfaceName,
		networkManager: networkManager,
		sender:         sender,
		onReceive:      onReceive,
	}

	return &r
}

func (r *Receiver) handleReceive(srcAddr *net.UDPAddr, data []byte) {
	var err error

	receivedTimestamp := time.Now()
	receivedAddress := srcAddr.String()

	container, err := serialization.Deserialize(data)
	if err != nil {
		log.Printf("warning: attempt to deserialize returned %#+v for %#+v from %#+v", err, string(data), srcAddr)
		return
	}

	container.ReceivedTimestamp = receivedTimestamp
	container.ReceivedAddress = receivedAddress

	// in all cases sendAck an markAck so that the sender stops trying to sendAck it (even if it's not for us,
	// no amount of resending it to us will fix that)
	if container.Frame.NeedsAck {
		err = r.sender.SendAck(container)
		if err != nil {
			log.Printf("warning: attempt to send ack returned %v for %#+v", err, container)
		}
	}

	if container.NetworkID != r.networkID {
		log.Printf("warning: ignoring container because NetworkID %#v unknown in %#+v", r.networkID, container)
		return
	}

	if container.Frame.IsAck {
		// TODO: maybe some sort of background worker pool vs unbounded amount of goroutines
		go r.sender.MarkAck(container)
		return
	}

	r.onReceive(container)
}

func (r *Receiver) Start() {
	err := r.networkManager.RegisterCallback(
		fmt.Sprintf("0.0.0.0:%v", r.listenPort),
		r.interfaceName,
		r.handleReceive,
	)
	if err != nil {
		log.Printf("warning: attempt to register callback failed stating: %v", err)
	}
}

func (r *Receiver) Stop() {
	err := r.networkManager.UnregisterCallback(
		fmt.Sprintf("0.0.0.0:%v", r.listenPort),
		r.interfaceName,
		r.handleReceive,
	)
	if err != nil {
		log.Printf("warning: attempt to unregister callback failed stating: %v", err)
	}
}
