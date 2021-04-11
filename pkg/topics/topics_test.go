package topics

import (
	"log"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"

	"github.com/initialed85/glue/pkg/discovery"
	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/transport"
	"github.com/initialed85/glue/pkg/types"
)

func getThings(endpointName string, listenPort int) (*network.Manager, ksuid.KSUID, *discovery.Manager, chan types.Container, chan types.Container, *transport.Manager, *Manager) {
	networkManager := network.NewManager()

	added := make(chan types.Container, 65536)
	removed := make(chan types.Container, 65536)

	endpointID := ksuid.New()

	discoveryManager := discovery.NewManager(
		1,
		endpointID,
		endpointName,
		listenPort,
		"239.192.137.1:27320",
		"en0",
		time.Millisecond*100,
		3,
		networkManager,
		func(container types.Container) {
			added <- container
		},
		func(container types.Container) {
			removed <- container
		},
	)

	var topicsManager *Manager

	transportManager := transport.NewManager(
		1,
		endpointID,
		endpointName,
		listenPort,
		"en0",
		discoveryManager,
		networkManager,
		func(container types.Container) {
			topicsManager.HandleReceive(container)
		},
	)

	topicsManager = NewManager(
		endpointID,
		endpointName,
		transportManager,
	)

	return networkManager, endpointID, discoveryManager, added, removed, transportManager, topicsManager
}

func startThings(networkManager *network.Manager, discoveryManager *discovery.Manager, transportManager *transport.Manager, topicsManager *Manager) {
	networkManager.Start()
	discoveryManager.Start()
	transportManager.Start()
	topicsManager.Start()
}

func stopThings(networkManager *network.Manager, discoveryManager *discovery.Manager, transportManager *transport.Manager, topicsManager *Manager) {
	networkManager.Stop()
	discoveryManager.Stop()
	transportManager.Stop()
	topicsManager.Stop()
}

func TestIntegration_Manager(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	networkManager1, _, discoveryManager1, added1, removed1, transportManager1, topicsManager1 := getThings("A", 27321)
	startThings(networkManager1, discoveryManager1, transportManager1, topicsManager1)

	networkManager2, _, discoveryManager2, added2, _, transportManager2, topicsManager2 := getThings("B", 27322)
	startThings(networkManager2, discoveryManager2, transportManager2, topicsManager2)

	select {
	case added := <-added1:
		assert.Equal(t, "B", added.SourceEndpointName)
	case <-time.After(time.Second):
		log.Fatal("timed out waiting for A to see B")
	}

	select {
	case added := <-added2:
		assert.Equal(t, "A", added.SourceEndpointName)
	case <-time.After(time.Second):
		log.Fatal("timed out waiting for B to see A")
	}

	consumed1 := make(chan []byte, 65536)

	err := topicsManager1.Subscribe(
		"some_topic",
		"some_type",
		func(message Message) {
			consumed1 <- message.Payload
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = topicsManager2.Publish(
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

	stopThings(networkManager2, discoveryManager2, transportManager2, topicsManager2)

	select {
	case removed := <-removed1:
		assert.Equal(t, "B", removed.SourceEndpointName)
	case <-time.After(time.Second):
		log.Fatal("timed out waiting for A to stop seeing B")
	}

	select {
	case _ = <-removed1:
		log.Fatal("B should not have seen A")
	case <-time.After(time.Second):
		// noop
	}

	stopThings(networkManager1, discoveryManager1, transportManager1, topicsManager1)
}
