package types

import (
	"net"
	"time"

	"github.com/segmentio/ksuid"
)

func GetAnnouncementContainer(
	sentTimestamp time.Time,
	sentBy string,
	networkID int64,
	sourceEndpointID ksuid.KSUID,
	sourceEndpointName string,
	sentRate time.Duration,
	discoveryListenAddress *net.UDPAddr,
	discoveryTargetAddress *net.UDPAddr,
	listenAddress *net.UDPAddr,
) *Container {
	return &Container{
		SentTimestamp:      sentTimestamp,
		SentBy:             sentBy,
		NetworkID:          networkID,
		SourceEndpointID:   sourceEndpointID,
		SourceEndpointName: sourceEndpointName,
		Announcement: &Announcement{
			SentRate:               sentRate,
			ListenPort:             listenAddress.Port,
			ListenAddr:             listenAddress,
			DiscoveryListenAddress: discoveryListenAddress.String(),
			DiscoveryListenAddr:    discoveryListenAddress,
			DiscoveryTargetAddress: discoveryTargetAddress.String(),
			DiscoveryTargetAddr:    discoveryTargetAddress,
		},
	}
}

func GetFrameContainer(
	resendPeriod time.Duration,
	resendExpiry time.Duration,
	networkID int64,
	sourceEndpointID ksuid.KSUID,
	sourceEndpointName string,
	correlationID ksuid.KSUID,
	fragmentCount int64,
	fragmentIndex int64,
	destinationEndpointID ksuid.KSUID,
	destinationEndpointName string,
	needsAck bool,
	isAck bool,
	payload []byte,
) *Container {
	return &Container{
		NetworkID:          networkID,
		SourceEndpointID:   sourceEndpointID,
		SourceEndpointName: sourceEndpointName,
		Frame: &Frame{
			ResendPeriod:            resendPeriod,
			ResendExpiry:            resendExpiry,
			FrameID:                 ksuid.New(),
			CorrelationID:           correlationID,
			FragmentCount:           fragmentCount,
			FragmentIndex:           fragmentIndex,
			DestinationEndpointID:   destinationEndpointID,
			DestinationEndpointName: destinationEndpointName,
			NeedsAck:                needsAck,
			IsAck:                   isAck,
			Payload:                 payload,
		},
	}
}

func GetFrameAckContainer(
	networkID int64,
	sourceEndpointID ksuid.KSUID,
	sourceEndpointName string,
	container *Container,
) *Container {
	return &Container{
		NetworkID:          networkID,
		SourceEndpointID:   sourceEndpointID,
		SourceEndpointName: sourceEndpointName,
		Frame: &Frame{
			ResendPeriod:            time.Duration(0), // doesn't matter
			ResendExpiry:            time.Duration(0), // doesn't matter
			FrameID:                 container.Frame.FrameID,
			CorrelationID:           container.Frame.CorrelationID,
			FragmentCount:           1, // fixed
			FragmentIndex:           0, // fixed
			DestinationEndpointID:   container.SourceEndpointID,
			DestinationEndpointName: container.SourceEndpointName,
			NeedsAck:                false,
			IsAck:                   true,
			Payload:                 []byte{},
		},
	}
}
