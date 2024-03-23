package topics

import (
	"time"

	"github.com/segmentio/ksuid"
)

const scheduledWorkerRate = time.Second * 1

type MessageType int

// not using iota as a means of being explicit
const (
	// a normal, directly delivered message
	StandardMessageType MessageType = 1

	// a message that came via the forwarded / late joiner path
	ForwardedMessageType MessageType = 2

	// a later joiner sends this when a new endpoint is discovered
	LateJoinerMessagesRequestType MessageType = 3

	// the new endpoint sends this in response
	LateJoinerMessagesResponseType MessageType = 4
)

type Message struct {
	// trust the sent timestamp as it's the source of truth for the state in this message
	Timestamp time.Time `json:"timestamp"`

	// discard the message after it's this much older
	Expiry time.Duration `json:"expiry"`

	// original source
	EndpointID   ksuid.KSUID `json:"endpoint_id"`
	EndpointName string      `json:"endpoint_name"`

	// incremented by sending endpoint
	SequenceNumber int64 `json:"sequence_number"`

	// topic
	TopicName string `json:"topic_name"`
	TopicType string `json:"topic_type"`

	// to identify and route the payload
	MessageType MessageType `json:"message_type"`

	// the actual user payload / control message content
	Payload []byte `json:"payload"`
}

type MessageIdentifier struct {
	EndpointID     ksuid.KSUID
	SequenceNumber int64
}

type LateJoinerMessagesRequest struct {
	// messages the late joiner already knows about
	KnownMessages []MessageIdentifier `json:"known_messages"`
}

type LateJoinerMessagesResponse struct {
	// messages that the endpoint is holding that the late joiner doesn't have already
	HeldMessages []Message `json:"held_messages"`
}
