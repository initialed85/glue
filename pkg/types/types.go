package types

import (
	"fmt"
	"net"
	"time"

	"github.com/segmentio/ksuid"
)

type Announcement struct {
	// used to determine when to expire an announcement
	SentRate time.Duration `json:"sent_rate"`

	// UDP port the announced endpoint is listening on for unicast data communications
	ListenPort int `json:"listen_port"`
	ListenAddr *net.UDPAddr

	// UDP address the announced endpoint is sending discovery announcements
	DiscoveryListenAddress string `json:"discovery_listen_address"`
	DiscoveryListenAddr    *net.UDPAddr

	// UDP address the announced endpoint is listening on for discovery announcements
	DiscoveryTargetAddress string `json:"discovery_target_address"`
	DiscoveryTargetAddr    *net.UDPAddr

	// used to avoid announcement forwarding loops
	Forwarded bool
}

func (a *Announcement) String() string {
	return fmt.Sprintf(
		"Announcement[%v / %v for %v on port %v]",
		a.DiscoveryListenAddress,
		a.DiscoveryTargetAddress,
		a.SentRate,
		a.ListenPort,
	)
}

func (a *Announcement) Copy() *Announcement {
	return &Announcement{
		SentRate:               a.SentRate,
		ListenPort:             a.ListenPort,
		ListenAddr:             a.ListenAddr,
		DiscoveryListenAddress: a.DiscoveryListenAddress,
		DiscoveryListenAddr:    a.DiscoveryListenAddr,
		DiscoveryTargetAddress: a.DiscoveryTargetAddress,
		DiscoveryTargetAddr:    a.DiscoveryTargetAddr,
		Forwarded:              a.Forwarded,
	}
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

func (f *Frame) String() string {
	return fmt.Sprintf(
		"Frame[%v / %v (%v / %v) for %v (%v); %v]",
		f.FrameID.String(),
		f.CorrelationID.String(),
		f.FragmentIndex+1,
		f.FragmentCount,
		f.DestinationEndpointName,
		f.DestinationEndpointID,
		len(f.Payload),
	)
}

func (f *Frame) Copy() *Frame {
	return &Frame{
		ResendPeriod:            f.ResendPeriod,
		ResendExpiry:            f.ResendExpiry,
		FrameID:                 f.FrameID,
		CorrelationID:           f.CorrelationID,
		FragmentCount:           f.FragmentCount,
		FragmentIndex:           f.FragmentIndex,
		DestinationEndpointID:   f.DestinationEndpointID,
		DestinationEndpointName: f.DestinationEndpointName,
		NeedsAck:                f.NeedsAck,
		IsAck:                   f.IsAck,
		Payload:                 f.Payload,
	}
}

type Container struct {
	// timestamp the sender originally sent the container
	SentTimestamp time.Time `json:"sent_timestamp"`

	// when then sender most recently sent the container (according to the sender)
	LastSentTimestamp time.Time `json:"last_sent_timestamp"`

	// source address according to the sender
	SentBy     string `json:"sent_by"`
	SentByAddr *net.UDPAddr

	// destination address according to the sender
	SentTo     string `json:"sent_to"`
	SentToAddr *net.UDPAddr

	// timestamp the receiver received the container
	ReceivedTimestamp time.Time `json:"-"`

	// destination address according to the receiver
	ReceivedFrom     string `json:"-"`
	ReceivedFromAddr *net.UDPAddr

	// source address according to the receiver
	ReceivedBy     string `json:"-"`
	ReceivedByAddr *net.UDPAddr

	// used to identify a group of endpoints
	NetworkID int64 `json:"network_id"`

	// automatically generated per endpoint lifecycle
	SourceEndpointID ksuid.KSUID `json:"source_endpoint_id"`

	// set by user- cannot appear twice in the same network
	SourceEndpointName string `json:"source_endpoint_name"`

	// content for an announcement
	Announcement *Announcement `json:"announcement"`

	// content for a frame
	Frame *Frame `json:"frame"`
}

func (c *Container) String() string {
	content := "(null)"

	if c.Announcement != nil {
		content = c.Announcement.String()
	} else if c.Frame != nil {
		content = c.Frame.String()
	}

	return fmt.Sprintf(
		"Container[%v: %v (%v) @ %v (%v) -> %v (%v); %v]",
		c.NetworkID,
		c.SourceEndpointName,
		c.SourceEndpointID,
		c.SentBy,
		c.ReceivedFrom,
		c.SentTo,
		c.ReceivedBy,
		content,
	)
}

func (c *Container) Copy() *Container {
	var announcement *Announcement
	if c.Announcement != nil {
		announcement = c.Announcement.Copy()
	}

	var frame *Frame
	if c.Frame != nil {
		frame = c.Frame.Copy()
	}

	return &Container{
		SentTimestamp:      c.SentTimestamp,
		LastSentTimestamp:  c.LastSentTimestamp,
		SentBy:             c.SentBy,
		SentByAddr:         c.SentByAddr,
		SentTo:             c.SentTo,
		SentToAddr:         c.SentToAddr,
		ReceivedTimestamp:  c.ReceivedTimestamp,
		ReceivedFrom:       c.ReceivedFrom,
		ReceivedFromAddr:   c.ReceivedFromAddr,
		ReceivedBy:         c.ReceivedBy,
		ReceivedByAddr:     c.ReceivedByAddr,
		NetworkID:          c.NetworkID,
		SourceEndpointID:   c.SourceEndpointID,
		SourceEndpointName: c.SourceEndpointName,
		Announcement:       announcement,
		Frame:              frame,
	}
}
