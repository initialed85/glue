package discovery

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"

	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/types"
)

func getThings(endpointName string, listenPort int, address string) (*network.Manager, *Manager, chan *types.Container, chan *types.Container) {
	networkManager := network.NewManager()

	added := make(chan *types.Container, 65536)
	removed := make(chan *types.Container, 65536)

	endpointID := ksuid.New()

	listenAddress, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%v", listenPort))
	discoveryListenAddress, _ := net.ResolveUDPAddr("udp4", address)
	discoveryTargetAddress, _ := net.ResolveUDPAddr("udp4", address)

	discoveryManager := NewManager(
		1,
		endpointID,
		endpointName,
		listenAddress,
		discoveryListenAddress,
		discoveryTargetAddress,
		"en0",
		time.Millisecond*100,
		3,
		networkManager,
		func(container *types.Container) {
			added <- container
		},
		func(container *types.Container) {
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
	time.Sleep(time.Second * 1)
}

func TestIntegration_Manager(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	t.Run("Multicast", func(t *testing.T) {
		networkManager1, discoveryManager1, added1, removed1 := getThings("A", 27321, "239.192.137.1:27320")
		startThings(networkManager1, discoveryManager1)

		networkManager2, discoveryManager2, added2, _ := getThings("B", 27322, "239.192.137.1:27320")
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
	})

	t.Run("Unicast", func(t *testing.T) {
		t.Skip("probably can't get this working without Docker for network isolation")

		networkManager0, discoveryManager0, _, _ := getThings("Z", 27319, "239.192.137.1:27320")
		startThings(networkManager0, discoveryManager0)

		networkManager1, discoveryManager1, added1, removed1 := getThings("A", 27321, "127.0.0.1:27320")
		startThings(networkManager1, discoveryManager1)

		networkManager2, discoveryManager2, added2, _ := getThings("B", 27322, "127.0.0.1:27320")
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

		stopThings(networkManager0, discoveryManager0)
	})
}
