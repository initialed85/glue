package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/initialed85/glue/pkg/endpoint"
	"github.com/initialed85/glue/pkg/helpers"
	"github.com/initialed85/glue/pkg/topics"
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
			inputBuf := []byte(fmt.Sprintf("%v;", sequence))

			inputSize := 65536

			for len(inputBuf) < inputSize {
				inputBuf = append(inputBuf, byte(rand.Intn(255)))
			}

			outputBuf := bytes.NewBuffer(inputBuf)

			err := endpointManager.Publish(
				*topicName,
				"some_type",
				time.Millisecond*50,
				outputBuf.Bytes(),
			)
			if err != nil {
				log.Printf("warning: %v", err)
				return
			}

			sequence++
		},
		func() {},
		time.Millisecond*50,
	)

	lastTimestamp := time.Now()
	var lastSequence int64 = 0
	var lostSyncCount int64 = 0

	if *sendMessages {
		scheduleWorker.Start()
	} else {

		err := endpointManager.Subscribe(
			*topicName,
			"some_type",
			func(message *topics.Message) {
				rawPayload := bytes.NewBuffer(message.Payload)

				payload := strings.Split(rawPayload.String(), ";")[0]

				sequence, err = strconv.ParseInt(payload, 10, 64)
				if err != nil {
					log.Print(err)
					return
				}

				if sequence%20 == 0 {
					log.Printf(
						"from=%#+v, age=%v, seq=%v, seq_diff=%v, lost_sync_count=%v",
						message.EndpointName,
						message.Timestamp.Sub(lastTimestamp),
						sequence,
						sequence-lastSequence,
						lostSyncCount,
					)
				}

				if sequence-lastSequence > 1 {
					lostSyncCount += sequence - lastSequence - 1
				}

				lastTimestamp = message.Timestamp
				lastSequence = sequence
			},
		)
		if err != nil {
			log.Printf("warning: %#+v", err)
		}
	}

	log.Print("press Ctrl + C to exit...")
	helpers.WaitForCtrlC()

	if *sendMessages {
		scheduleWorker.Stop()
	}

	endpointManager.Stop()
}
