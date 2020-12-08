# glue

This repo contains some Go code as a reference implementation for a brokerless/distributed pub/sub concept I've been toying with that's 
not entirely unlike OMG's DDS or Intel's DPS for IoT.

## Goal

I want two or more endpoints on the same layer 2 network segment to share data via topics without needing to know about IP addresses or 
port numbers, and I don't want to rely on a broker.

Put simply, I want a robust alternative to MQTT without a broker.

## Conceptual architecture

From top to bottom:

- Model (NOT STARTED)
    - this one I have the loosest concept of
    - basically I wanna built on the Topics layer to send around a delta of data and build up a distributed model of state
    - should deal with conflicts, consensus / quorum etc
    - maybe somebody has done a good Paxos library?
- Topics (NOT STARTED)
    - addressing is topic names
    - publish / subscribe
    - handle late joiners (at the publisher level)
    - handle network partitions (any endpoint can cache messages)
- Fragmentation (IN PROGRESS)
    - addressing is endpoint IDs and names
    - send / receive
    - handle fragmentation / defragmentation of messages
- Transport (COMPLETE)
    - addressing is endpoint IDs and names
    - send / receive
    - handle ACKs / resending of messages
- Discovery (COMPLETE)
    - addressing is endpoint IDs and names
    - announce / listen
    - handle add on discovery / remove on expiry
- Serialization (COMPLETE)
    - cross-platform/cross-language format (right now it's just JSON)
- Network (COMPLETE)
    - shared abstraction for low level network interactions

## Example

### Transport

An example usage of the Transport layer is at `cmd/transport/main.go`.

To use; run the following 4 commands in 4 different shells (note: you may need to change the interfaceName flag):

```
go run cmd/transport/main.go -interfaceName en0 -multicastAddress 239.192.137.1:37320 -listenPort 37321 -endpointName A -sendMessages
go run cmd/transport/main.go -interfaceName en0 -multicastAddress 239.192.137.1:37320 -listenPort 37322 -endpointName B -sendMessages
go run cmd/transport/main.go -interfaceName en0 -multicastAddress 239.192.137.1:37320 -listenPort 37323 -endpointName C -sendMessages
go run cmd/transport/main.go -interfaceName en0 -multicastAddress 239.192.137.1:37320 -listenPort 37324 -endpointName D -sendMessages
```

### Topics

An example usage of the Transport layer is at `cmd/transport/main.go`.

To use; run the following 4 commands in 4 different shells (note: you may need to change the interfaceName flag):

```
go run cmd/topics/main.go -interfaceName en0 -listenPort 37321 -sendMessages
go run cmd/topics/main.go -interfaceName en0 -listenPort 37322
go run cmd/topics/main.go -interfaceName en0 -listenPort 37323
go run cmd/topics/main.go -interfaceName en0 -listenPort 37324
```

### Endpoint

```
go run cmd/endpoint/main.go -interfaceName en0 -listenPort 37321 -sendMessages
go run cmd/endpoint/main.go -interfaceName en0 -listenPort 37322
go run cmd/endpoint/main.go -interfaceName en0 -listenPort 37323
go run cmd/endpoint/main.go -interfaceName en0 -listenPort 37324
```
