package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/initialed85/glue/pkg/discovery"
	"github.com/initialed85/glue/pkg/network"
	"github.com/initialed85/glue/pkg/transport"
	"github.com/initialed85/glue/pkg/types"
	"github.com/initialed85/glue/pkg/utils"
	"github.com/initialed85/glue/pkg/worker"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	networkID := flag.Int64("networkID", 1, "")
	endpointName := flag.String("endpointName", "", "")
	interfaceName := flag.String("interfaceName", "", "")
	listenPort := flag.Int("listenPort", -1, "")
	multicastAddress := flag.String("multicastAddress", "239.192.137.1:27320", "")
	rateMillis := flag.Int64("rateMillis", 1000, "")
	timeoutMultiplier := flag.Float64("timeoutMultiplier", 3, "")
	sendMessages := flag.Bool("sendMessages", false, "")

	flag.Parse()

	endpointID := ksuid.New()

	if *endpointName == "" {
		*endpointName = fmt.Sprintf("Endpoint_%v", endpointID.String())
	}

	if *interfaceName == "" {
		log.Fatal("interfaceName not specified")
	}

	if *listenPort < 1 {
		log.Fatal("listenPort not specified (or less than 1)")
	}

	networkManager := network.NewManager()

	discoveryManager := discovery.NewManager(
		*networkID,
		endpointID,
		*endpointName,
		*listenPort,
		*multicastAddress,
		*interfaceName,
		time.Millisecond*time.Duration(*rateMillis),
		*timeoutMultiplier,
		networkManager,
		func(container types.Container) {
			// noop
		},
		func(container types.Container) {
			// noop
		},
	)

	transportManager := transport.NewManager(
		*networkID,
		endpointID,
		*endpointName,
		*listenPort,
		*interfaceName,
		discoveryManager,
		networkManager,
		func(container types.Container) {
			log.Printf(
				"endpointID=%#v, endpointName=%#v, sentAddress=%#v, latency=%#+v, receivedTimestamp=%#v, payload=%#v",
				container.SourceEndpointID.String(),
				container.SourceEndpointName,
				container.SentAddress,
				container.ReceivedAddress,
				container.ReceivedTimestamp.Sub(container.SentTimestamp).String(),
				string(container.Frame.Payload),
			)
		},
	)

	networkManager.Start()
	discoveryManager.Start()
	transportManager.Start()

	scheduleWorker := worker.NewScheduledWorker(
		func() {},
		func() {
			transportManager.Broadcast(
				time.Millisecond*100,
				time.Second,
				ksuid.New(),
				1,
				0,
				true,
				[]byte(fmt.Sprintf(
					"Some payload as of %v",
					time.Now().String(),
				)),
			)
		},
		func() {},
		time.Millisecond*200,
	)

	if *sendMessages {
		scheduleWorker.Start()
	}

	log.Print("press Ctrl + C to exit...")
	utils.WaitForCtrlC()

	if *sendMessages {
		scheduleWorker.Stop()
	}

	transportManager.Stop()
	discoveryManager.Stop()
	networkManager.Stop()

}
