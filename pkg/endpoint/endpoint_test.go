package endpoint

import (
	"log"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"

	"glue/pkg/topics"
	"glue/pkg/types"
)

func getThings(endpointName string, listenPort int) *Manager {
	return NewManager(
		1,
		ksuid.New(),
		endpointName,
		listenPort,
		"239.192.137.1:27320",
		"en0",
		time.Millisecond*100,
		3,
		func(container types.Container) {},
		func(container types.Container) {},
	)
}

func startThings(endpointManager *Manager) {
	endpointManager.Start()
}

func stopThings(endpointManager *Manager) {
	endpointManager.Stop()
}

func TestIntegration_Manager(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	endpointManager1 := getThings("A", 27321)
	startThings(endpointManager1)

	endpointManager2 := getThings("B", 27322)
	startThings(endpointManager2)

	time.Sleep(time.Millisecond * 100)

	consumed1 := make(chan []byte, 65536)

	err := endpointManager1.Subscribe(
		"some_topic",
		"some_type",
		func(message topics.Message) {
			consumed1 <- message.Payload
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	err = endpointManager2.Publish(
		"some_topic",
		"some_type",
		time.Second,
		[]byte("Some payload"),
	)
	if err != nil {
		log.Fatal(err)
	}

	select {
	case consumed := <-consumed1:
		assert.Equal(t, []byte("Some payload"), consumed)
	case <-time.After(time.Second):
		log.Fatal("timed out waiting for A to receive a publication from B")
	}

	stopThings(endpointManager2)

	stopThings(endpointManager1)
}
