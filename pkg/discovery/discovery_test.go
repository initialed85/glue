package discovery

import (
	"log"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"

	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/types"
)

func getThings(endpointName string, listenPort int) (*network.Manager, *Manager, chan types.Container, chan types.Container) {
	networkManager := network.NewManager()

	added := make(chan types.Container, 65536)
	removed := make(chan types.Container, 65536)

	endpointID := ksuid.New()

	discoveryManager := NewManager(
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

	return networkManager, discoveryManager, added, removed
}

func startThings(networkManager *network.Manager, discoveryManager *Manager) {
	networkManager.Start()
	discoveryManager.Start()
}

func stopThings(networkManager *network.Manager, discoveryManager *Manager) {
	discoveryManager.Stop()
	networkManager.Stop()
}

func TestIntegration_Manager(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	networkManager1, discoveryManager1, added1, removed1 := getThings("A", 27321)
	startThings(networkManager1, discoveryManager1)

	networkManager2, discoveryManager2, added2, _ := getThings("B", 27322)
	startThings(networkManager2, discoveryManager2)

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

	stopThings(networkManager2, discoveryManager2)

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

	stopThings(networkManager1, discoveryManager1)
}
