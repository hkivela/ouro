### Description
This project is an attempt to build a comprehensive library to connect to IP cameras and consume video and audio streams.
The ultimate goal is to create a robust streaming proxy converting IP camera stream into something directly usable from a browser.

### Branch "testable"
The idea is to avoid running main loop in a goroutine and instead act as "listen and serve" endless loop in a method.  The caller would be free to run it in a goroutine or to run it in the main thread and issue commands from a goroutine.  Command would be passed through a method that will queue them in a synchronized way.  MAin loop will pull command from the queue and execute one at a time.  Session will issue command and wait for a response, then verify that response matches CSeq, then process it.  RTP and RTCP packets coming in while waiting for response will be pushed into respective channel.

This approach has several adavantages:
- greater flexibility for the caller
- each command could be wrapped into a testable method: issue request, wait for response, verify response
- there is no need to track requests in a map and lookup and match responses
- transport setup will happen one at a time, no need to match which transport is being set from response

There are disadvantages:
- caller won't be able to get a result of issued command
- caller may get result from the issued command but it will require complex implementation

This is worth a try though.

### Features
At this time the following functionality has been implemented:
- Connecting to camera over RTSP.
- OPTIONS, DESCRIBE, SETUP, PLAY, PAUSE, and TEARDOWN commands.
- Handling Basic and Digest authentication.
- Simplistic parsing of SDP data.
- Parsing and building Transport header.
- Handling RTSP state machine with CSeq.
- Receiving RTP and RTCP packets over TCP.
- Unwrapping RTP/RTCP packets from RTSP message.
- Parsing RTP packets for h.264 NAL units.
- Handling NAL aggrgates, fragments, DONs (Decoding Order Number) and timestamps.
- Receiving and parsing basic RTCP packets.
- Initial work on UDP listeners for RTP over UDP.
- Initial work on RTSP over HTTP.

### Building and running
Get the source:
```
go get https://github.com/aboukirev/ouro
```
Execute either `make build` or  `go build ./cmd/ouro`.
Run `./ouro "rtsp://user:password@ip-or-domain/path"`.  It will out put log of the short session to console.  At this time it just connects to camera, starts session, consumes stream for 2 seconds, pauses for 2 more seconds, then terminates connection.  

### Plans
I wanted to get real response data before I start writing tests for packets and messages.  That is in the plans.
It is hard to test RTSP as it operates as a state machine: RTSP messages (text protocol) and RTP packets (binary protocol) are coming through the same connection in an unpredictable order. 

I am building just client functionality.  I plan to finish RTP/RTCP over UDP, add sending keep-alive RTCP packets, implement RTSP over HTTP eventually.
Then the plans include transforming media streams to output HLS or MPEG-DASH.

There are many more intricacies involved in handling protocols proper.  For instance, packetisation mode from SDP can help with hadling RTP payload.  Sequence number returned in response to PLAY command can be used to find initial RTP packet to start streaming with, etc.  A better SDP parser would be useful.  Sorting out of order NAL units in MTAP NAL could be useful if there are cameras sending MTAPs.

