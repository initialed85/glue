# glue

# status: working (but needs more work) and is in production for some of my personal projects

This repo contains some Go code as a reference implementation for a brokerless/distributed pub/sub concept I've been toying with that's
not entirely unlike OMG's DDS or Intel's DPS for IoT.

## Goal

I want two or more endpoints on the same layer 2 network segment to share data via topics without needing to know about IP addresses or
port numbers, and I don't want to rely on a broker.

Put simply, I want a robust alternative to MQTT without a broker.

## Conceptual architecture

From top to bottom:

-   Model (NOT STARTED)
    -   this one I have the loosest concept of
    -   basically I wanna built on the Topics layer to send around a delta of data and build up a distributed model of state
    -   should deal with conflicts, consensus / quorum etc
    -   maybe somebody has done a good Paxos library?
-   Topics (IN PROGRESS)
    -   addressing is topic names
    -   publish / subscribe
    -   handle late joiners (at the publisher level)
    -   handle network partitions (any endpoint can cache messages)
-   Fragmentation (IN PROGRESS)
    -   addressing is endpoint IDs and names
    -   send / receive
    -   handle fragmentation / defragmentation of messages
-   Transport (COMPLETE)
    -   addressing is endpoint IDs and names
    -   send / receive
    -   handle ACKs / resending of messages
-   Discovery (COMPLETE)
    -   addressing is endpoint IDs and names
    -   announce / listen
    -   handle add on discovery / remove on expiry
-   Serialization (COMPLETE)
    -   cross-platform/cross-language format (right now it's just JSON)
-   Network (COMPLETE)
    -   shared abstraction for low level network interactions

## Usage

You can tap into the various layers of abstraction as suits your needs, but the simple Endpoint will probably cater to most of your needs (it's geared around a single Go program participating as a single Endpoint).

A subscriber looks something like this:

```go
endpointManager, err := endpoint.NewManagerSimple()
if err != nil {
		log.Fatal(err)
}

endpointManager.Start()
defer endpointManager.Stop()

err := endpointManager.Subscribe(
    "some_topic",
    "some_type",
    func(message *topics.Message) {
        log.Printf("%v said %#+v", message.EndpointName, string(message.Payload))
    },
)
if err != nil {
    log.Printf("warning: %#+v", err)
}

for {
    time.Sleep(time.Second*1)
}
```

And a publisher looks something like this:

```go
endpointManager, err := endpoint.NewManagerSimple()
if err != nil {
    log.Fatal(err)
}

endpointManager.Start()
defer endpointManager.Stop()

for {
    err := endpointManager.Publish(
        "some_topic",
        "some_type",
        time.Millisecond*100, // message expiry
        []byte("Hello, world."),
    )
    if err != nil {
        log.Printf("warning: %#+v", err)
    }

time.Sleep(time.Second*1)
}
```

Given the focus around a single Go program being a single Endpoint, you can inject a bunch of config for the Endpoint at runtime using environment variables:

-   `GLUE_NETWORK_ID`
    -   An identifier that must be common between all endpoints that wish to communicate
    -   Can be used to establish separate logical partitions within the same physical network
-   `GLUE_ENDPOINT_ID: int`
    -   A unique identifier for an endpoint that is expected to remain constant across reboots of the endpoint
    -   Intended to be used to ensure the uniqueness of a specific service role (e.g. TimeSyncProducer)
-   `GLUE_ENDPOINT_NAME: string`
    -   A unique identifier for an endpoint that is expected to change across reboots of the endpoint
    -   Intended to be used to ensure the uniqueness of a specific instance of a service role (e.g. TimeSyncProducer_123abc)
-   `GLUE_LISTEN_ADDRESS`
    -   A UDP address (typically unicast) to listen for Glue data packets on
-   `GLUE_LISTEN_INTERFACE`
    -   An interface name to bind to while listening for Glue data packets
-   `GLUE_DISCOVERY_TARGET_ADDRESS`
    -   A UDP address (typically multicast, can be unicast or broadcast) to send Glue discovery announcement packets to
-   `GLUE_DISCOVERY_LISTEN_ADDRESS`
    -   A UDP address (typically multicast, can be unicast or broadcast) to listen for Glue discovery announcement packets on
-   `GLUE_DISCOVERY_RATE_MILLISECONDS`
    -   The rate (in milliseconds) at which to produce Glue discovery announcements packets
    -   e.g. 1000 = 1 second = 1 Hz
-   `GLUE_DISCOVERY_RATE_TIMEOUT_MULTIPLIER`
    -   A multiplier that describes the Glue discovery adjacency timeout when applied to the discovery rate
    -   e.g. 5 = 5 (x 1 second)

## Examples

### Spin up 3 subscribers and 1 producer with multicast discovery

This is mode is most like the typical DDS rollout for a layer 2 broadcast domain:

```shell
# shell 1
go run ./cmd/simple_endpoint/

# shell 2
go run ./cmd/simple_endpoint/

# shell 3
go run ./cmd/simple_endpoint/

# shell 4
go run ./cmd/simple_endpoint/ -sendMessages
```

### Spin up 3 subscribers and 1 producer with unicast discovery

This mode has utility for layer 3 / routed networks (e.g. Kubernetes) that aren't constrained by NAT:

```shell
# shell 1 - this is essentially our rendesvous point / discovery service
GLUE_DISCOVERY_TARGET_ADDRESS=0 go run ./cmd/simple_endpoint/

# shell 2
GLUE_DISCOVERY_TARGET_ADDRESS=127.0.0.1:27320 go run ./cmd/simple_endpoint/

# shell 3
GLUE_DISCOVERY_TARGET_ADDRESS=127.0.0.1:27320 go run ./cmd/simple_endpoint/

# shell 4
GLUE_DISCOVERY_TARGET_ADDRESS=127.0.0.1:27320 go run ./cmd/simple_endpoint/ -sendMessages
```
