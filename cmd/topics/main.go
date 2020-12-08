package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/ksuid"

	"glue/pkg/discovery"
	"glue/pkg/network"
	"glue/pkg/topics"
	"glue/pkg/transport"
	"glue/pkg/types"
	"glue/pkg/utils"
	"glue/pkg/worker"
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
	topicName := flag.String("topicName", "some_topic", "")

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

	var topicsManager *topics.Manager

	transportManager := transport.NewManager(
		*networkID,
		endpointID,
		*endpointName,
		*listenPort,
		*interfaceName,
		discoveryManager,
		networkManager,
		func(container types.Container) {
			topicsManager.HandleReceive(container)
		},
	)

	topicsManager = topics.NewManager(
		endpointID,
		*endpointName,
		transportManager,
	)

	networkManager.Start()
	discoveryManager.Start()
	transportManager.Start()
	topicsManager.Start()

	scheduleWorker := worker.NewScheduledWorker(
		func() {},
		func() {
			err := topicsManager.Publish(
				*topicName,
				"some_type",
				time.Second,
				[]byte(fmt.Sprintf("This is a payload as at %v", time.Now().String())),
			)
			if err != nil {
				log.Printf("warning: %#+v", err)
			}
		},
		func() {},
		time.Millisecond*200,
	)

	if *sendMessages {
		scheduleWorker.Start()
	} else {
		err := topicsManager.Subscribe(
			*topicName,
			"some_type",
			func(message topics.Message) {
				log.Printf(
					"received: %v sent %v", message.EndpointName, string(message.Payload),
				)
			},
		)
		if err != nil {
			log.Printf("warning: %#+v", err)
		}
	}

	log.Print("press Ctrl + C to exit...")
	utils.WaitForCtrlC()

	if *sendMessages {
		scheduleWorker.Stop()
	}

	topicsManager.Stop()
	transportManager.Stop()
	discoveryManager.Stop()
	networkManager.Stop()

}
