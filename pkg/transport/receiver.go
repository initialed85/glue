package transport

import (
	"log"
	"net"
	"time"

	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/serialization"
	"github.com/initialed85/glue/pkg/types"
)

type Receiver struct {
	networkID      int64
	listenAddress  *net.UDPAddr
	interfaceName  string
	networkManager *network.Manager
	sender         *Sender
	onReceive      func(*types.Container)
}

func NewReceiver(
	networkID int64,
	listenAddress *net.UDPAddr,
	interfaceName string,
	networkManager *network.Manager,
	sender *Sender,
	onReceive func(*types.Container),
) *Receiver {
	r := Receiver{
		networkID:      networkID,
		listenAddress:  listenAddress,
		interfaceName:  interfaceName,
		networkManager: networkManager,
		sender:         sender,
		onReceive:      onReceive,
	}

	return &r
}

func (r *Receiver) handleReceive(srcAddr *net.UDPAddr, dstAddr *net.UDPAddr, data []byte) {
	var err error

	receivedTimestamp := time.Now()

	container, err := serialization.Deserialize(data)
	if err != nil {
		log.Printf("warning: attempt to deserialize returned %#+v for %#+v from %#+v", err, string(data), srcAddr)
		return
	}

	container.ReceivedTimestamp = receivedTimestamp
	container.ReceivedFrom = srcAddr.String()
	container.ReceivedBy = dstAddr.String()

	if container.NetworkID != r.networkID {
		log.Printf("warning: ignoring container because NetworkID %v unknown in %v", r.networkID, container.String())
		return
	}

	if container.Frame == nil {
		log.Printf("error: unexpectedly received non-frame %#+v", container)
	}

	// in all cases sendAck marks a NeedsAck so that the sender stops trying to sendAck it (even if it's not for us,
	// no amount of resending it to us will fix that)
	if container.Frame.NeedsAck {
		err = r.sender.SendAck(container)
		if err != nil {
			log.Printf("warning: attempt to send ack returned %v for %v", err, container.String())
		}
	}

	if container.Frame.IsAck {
		// log.Printf("%v -> %v; receive ack frame from %v", srcAddr.String(), dstAddr, container.SourceEndpointName)

		// TODO: maybe some sort of background worker pool vs unbounded amount of goroutines
		go r.sender.MarkAck(container)
		return
	}

	// log.Printf("%v -> %v; receive data frame from %v", srcAddr.String(), dstAddr, container.SourceEndpointName)

	r.onReceive(container)
}

func (r *Receiver) Start() {
	err := r.networkManager.RegisterCallback(
		r.listenAddress,
		r.interfaceName,
		r.handleReceive,
	)
	if err != nil {
		log.Printf("warning: attempt to register callback failed stating: %v", err)
	}
}

func (r *Receiver) Stop() {
	err := r.networkManager.UnregisterCallback(
		r.listenAddress,
		r.interfaceName,
		r.handleReceive,
	)
	if err != nil {
		log.Printf("warning: attempt to unregister callback failed stating: %v", err)
	}
}
