package transport

import (
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/initialed85/glue/pkg/discovery"
	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/types"
	"log"
	"testing"
	"time"
)

func getThings(endpointName string, listenPort int) (*network.Manager, ksuid.KSUID, *discovery.Manager, chan types.Container, chan types.Container, *Manager, chan types.Container) {
	networkManager := network.NewManager()

	added := make(chan types.Container, 65536)
	removed := make(chan types.Container, 65536)
	received := make(chan types.Container, 65536)

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

	transportManager := NewManager(
		1,
		endpointID,
		endpointName,
		listenPort,
		"en0",
		discoveryManager,
		networkManager,
		func(container types.Container) {
			received <- container
		},
	)

	return networkManager, endpointID, discoveryManager, added, removed, transportManager, received
}

func startThings(networkManager *network.Manager, discoveryManager *discovery.Manager, transportManager *Manager) {
	networkManager.Start()
	discoveryManager.Start()
	transportManager.Start()
}

func stopThings(networkManager *network.Manager, discoveryManager *discovery.Manager, transportManager *Manager) {
	transportManager.Stop()
	discoveryManager.Stop()
	networkManager.Stop()
}

func TestIntegration_Manager(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	networkManager1, _, discoveryManager1, added1, removed1, transportManager1, _ := getThings("A", 27321)
	startThings(networkManager1, discoveryManager1, transportManager1)

	networkManager2, endpointID2, discoveryManager2, added2, _, transportManager2, received2 := getThings("B", 27322)
	startThings(networkManager2, discoveryManager2, transportManager2)

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

	err := transportManager1.Send(
		time.Millisecond*100,
		time.Second,
		ksuid.New(),
		1,
		0,
		endpointID2,
		"B",
		true,
		false,
		[]byte("Some payload"),
	)
	if err != nil {
		log.Fatal(err)
	}

	select {
	case received := <-received2:
		assert.Equal(t, "A", received.SourceEndpointName)
	case <-time.After(time.Second):
		log.Fatal("timed out waiting for B to receive from A")
	}

	stopThings(networkManager2, discoveryManager2, transportManager2)

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

	stopThings(networkManager1, discoveryManager1, transportManager1)
}
