package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/initialed85/glue/pkg/endpoint"
	"github.com/initialed85/glue/pkg/topics"
	"github.com/initialed85/glue/pkg/utils"
	"github.com/initialed85/glue/pkg/worker"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	sendMessages := flag.Bool("sendMessages", false, "")
	topicName := flag.String("topicName", "some_topic", "")

	flag.Parse()

	endpointManager, err := endpoint.NewManagerSimple()
	if err != nil {
		log.Fatal(err)
	}

	endpointManager.Start()

	var sequence int64 = 0

	scheduleWorker := worker.NewScheduledWorker(
		func() {},
		func() {
			err := endpointManager.Publish(
				*topicName,
				"some_type",
				time.Second,
				[]byte(fmt.Sprintf("%v", sequence)),
			)
			if err != nil {
				log.Printf("warning: %#+v", err)
			}

			sequence++
		},
		func() {},
		time.Millisecond*1000,
	)

	var lastSequence int64 = 0
	lastTimestamp := time.Now()

	if *sendMessages {
		scheduleWorker.Start()
	} else {

		err := endpointManager.Subscribe(
			*topicName,
			"some_type",
			func(message topics.Message) {
				sequence = lastSequence

				log.Printf(
					"from=%#+v, age=%v, seq=%v, seq_diff=%v",
					message.EndpointName,
					message.Timestamp.Sub(lastTimestamp),
					sequence,
					sequence-lastSequence,
				)

				lastSequence, err = strconv.ParseInt(string(message.Payload), 10, 64)
				if err != nil {
					log.Print(err)
				}

				lastTimestamp = message.Timestamp
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

	endpointManager.Stop()
}
