package serialization

import (
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/initialed85/glue/pkg/types"
	"log"
	"testing"
	"time"
)

func getAnnouncementContainer() types.Container {
	return types.GetAnnouncementContainer(
		time.Now(),
		"1.2.3.4:1234",
		1,
		ksuid.New(),
		"some-endpoint-1",
		time.Second,
		5678,
	)
}

func getFrameContainer() types.Container {
	return types.GetFrameContainer(
		time.Millisecond*100,
		time.Second,
		1,
		ksuid.New(),
		"some-endpoint-1",
		ksuid.New(),
		1,
		0,
		ksuid.New(),
		"other-endpoint-1",
		true,
		false,
		[]byte("Some payload"),
	)
}

func testSerializeAndDeserializeContainer(t *testing.T, expected types.Container) {
	data, err := Serialize(expected)
	if err != nil {
		log.Fatal(err)
	}

	actual, err := Deserialize(data)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, expected.SentTimestamp.Format(time.RFC3339), actual.SentTimestamp.Format(time.RFC3339))

	if expected.Frame != nil {
		expected.Frame.ResendPeriod = time.Duration(0)
		expected.Frame.ResendExpiry = time.Duration(0)
	}

	expected.SentTimestamp = time.Time{}
	actual.SentTimestamp = time.Time{}

	assert.Equal(t, expected, actual)

	if expected.Announcement == nil {
		assert.Nil(t, actual.Announcement)
	}

	if expected.Frame == nil {
		assert.Nil(t, actual.Frame)
	}
}

func TestSerializeAndDeserializeAnnouncement(t *testing.T) {
	testSerializeAndDeserializeContainer(t, getAnnouncementContainer())
}

func TestSerializeAndDeserializeFrame(t *testing.T) {
	testSerializeAndDeserializeContainer(t, getFrameContainer())
}
