# glue

This repo contains some Go code as a reference implementation for a brokerless/distributed pub/sub concept 
I've been toying with that's not entirely unlike OMG's DDS.

## Goal

I want two or more endpoints on the same layer 2 network segment to share data via topics without needing to
know about IP addresses or port numbers and I don't want to rely on a broker.

## Conceptual architecture

From top to bottom:

- Topics
    - addressing is topic names
    - publish / subscribe
    - handle late joiners (at the publisher level)
    - handle network partitions (any endpoint can cache messages)
- Fragmentation
    - addressing is endpoint IDs and names
    - send / receive
    - handle fragmentation / defragmentation of messages
- Transport
    - addressing is endpoint IDs and names
    - send / receive
    - handle ACKs / resending of messages
- Discovery
    - addressing is endpoint IDs and names
    - announce / listen
    - handle add on discovery / remove on expiry
- Network
    - shared abstraction for low level network interactions
