package types

import (
	"time"

	"github.com/segmentio/ksuid"
)

type Announcement struct {
	// used to determine when to expire an announcement
	SentRate time.Duration `json:"sent_rate"`

	// UDP port the announced endpoint is listening on for unicast communications
	ListenPort int `json:"listen_port"`
}

type Frame struct {
	// don't attempt to resend until the frame is this much older
	ResendPeriod time.Duration `json:"-"`

	// don't attempt to resend the frame after it's this much older
	ResendExpiry time.Duration `json:"-"`

	// unique for this frame
	FrameID ksuid.KSUID `json:"frame_id"`

	// unique for a related group of frames
	CorrelationID ksuid.KSUID `json:"correlation_id"`

	// how many fragments there are (1 = single frame)
	FragmentCount int64 `json:"fragment_count"`

	// which one this is
	FragmentIndex int64 `json:"fragment_index"`

	// a distant SourceEndpointID
	DestinationEndpointID ksuid.KSUID `json:"destination_endpoint_id"`

	// a distance SourceEndpointName
	DestinationEndpointName string `json:"destination_endpoint_name"`

	// is a markAck needed?
	NeedsAck bool `json:"needs_ack"`

	// is this a markAck?
	IsAck bool `json:"is_ack"`

	// the actual user payload (or fragment thereof)
	Payload []byte `json:"payload"`
}

type Container struct {
	// timestamp the sender originally sent the container
	SentTimestamp time.Time `json:"sent_timestamp"`

	// when then sender most recently sent the container (according to the sender)
	LastSentTimestamp time.Time `json:"last_sent_timestamp"`

	// local address of the sender's socket
	SentAddress string `json:"sent_address"`

	// timestamp the receiver received the container
	ReceivedTimestamp time.Time `json:"-"`

	// remote address of the receiver's socket
	ReceivedAddress string `json:"-"`

	// used to identify a group of endpoints
	NetworkID int64 `json:"network_id"`

	// automatically generated per endpoint lifecycle
	SourceEndpointID ksuid.KSUID `json:"endpoint_id"`

	// set by user- cannot appear twice in the same network
	SourceEndpointName string `json:"endpoint_name"`

	// content for an announcement
	Announcement *Announcement `json:"announcement"`

	// content for a frame
	Frame *Frame `json:"frame"`
}
