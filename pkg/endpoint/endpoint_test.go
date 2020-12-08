package endpoint

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/initialed85/glue/pkg/topics"
)

func getThings() *Manager {
	m, err := NewManagerSimple()
	if err != nil {
		log.Fatal(err)
	}

	return m
}

func startThings(endpointManager *Manager) {
	endpointManager.Start()
}

func stopThings(endpointManager *Manager) {
	endpointManager.Stop()
}

func TestIntegration_ManagerSimple(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	endpointManager1 := getThings()
	startThings(endpointManager1)

	endpointManager2 := getThings()
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

func TestIntegration_ManagerSimpleSingleEndpointTalkingToItself(t *testing.T) {
	endpointManager := getThings()

	startThings(endpointManager)

	time.Sleep(time.Millisecond * 100)

	consumed := make(chan topics.Message, 65536)

	err := endpointManager.Subscribe(
		"some_topic",
		"some_type",
		func(message topics.Message) {
			consumed <- message
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	err = endpointManager.Publish(
		"some_topic",
		"some_type",
		time.Second,
		[]byte("Some payload"),
	)
	if err != nil {
		log.Fatal(err)
	}

	select {
	case consumed := <-consumed:
		assert.Equal(t, []byte("Some payload"), consumed.Payload)
	case <-time.After(time.Second):
		log.Fatal("timed out waiting to receive a payload from self")
	}

	stopThings(endpointManager)
}

func TestIntegration_ManagerSimpleSingleEndpointTalkingToItselfWithWildcardSubscription(t *testing.T) {
	endpointManager := getThings()

	startThings(endpointManager)

	time.Sleep(time.Millisecond * 100)

	consumed := make(chan topics.Message, 65536)

	err := endpointManager.Subscribe(
		"#",
		"some_type",
		func(message topics.Message) {
			consumed <- message
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	err = endpointManager.Publish(
		"some_topic",
		"some_type",
		time.Second,
		[]byte("Some payload"),
	)
	if err != nil {
		log.Fatal(err)
	}

	select {
	case consumed := <-consumed:
		assert.Equal(t, []byte("Some payload"), consumed.Payload)
	case <-time.After(time.Second):
		log.Fatal("timed out waiting to receive a payload from self")
	}

	stopThings(endpointManager)
}
