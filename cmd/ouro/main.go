package main

import (
	// "encoding/hex"
	"flag"
	"log"
	"os"
	"time"

	"github.com/aboukirev/ouro/net/h264"
	"github.com/aboukirev/ouro/net/rtcp"
	"github.com/aboukirev/ouro/net/rtp"
	"github.com/aboukirev/ouro/net/rtsp"
)

var (
	current int // FIXME: current is racy
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Panic("Wrong number of arguments.  Exactly one is expected.")
	}
	url := flag.Arg(0)

	current = rtsp.StageInit
	log.SetOutput(os.Stdout)
	sess := rtsp.NewSession()
	go rtspHandler(sess.Stage)
	go rtpHandler(sess.Data)
	go rtcpHandler(sess.Control)
	err := sess.Open(url, rtsp.ProtoTCP)
	if err != nil {
		log.Fatalln(err)
	}
	for current != rtsp.StageReady {
		time.Sleep(time.Second)
	}
	if err := sess.Play(); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 5)
	if err := sess.Pause(); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 5)
	if err := sess.Teardown(); err != nil {
		log.Fatal(err)
	}
	for current != rtsp.StageDone {
		time.Sleep(time.Second)
	}
}

func rtspHandler(ch chan int) {
	for {
		select {
		case stage := <-ch:
			current = stage
			if stage == rtsp.StageReady {
				log.Println("Ready")
			} else if stage == rtsp.StagePlay {
				log.Println("Playing")
			} else if stage == rtsp.StagePause {
				log.Println("Pausing")
			} else if stage == rtsp.StageDone {
				log.Println("Done")
				return
			}
		}
	}
}

func rtpHandler(ch chan rtsp.RawPacket) {
	nalsink := h264.NewNALSink()
	for {
		select {
		case msg := <-ch:
			var p *rtp.Packet
			var err error
			if p, err = rtp.Unpack(msg.Payload); err != nil {
				log.Println(err)
				return
			}

			log.Printf("RTP [%d] PT=%d, CC=%d, M=%t, SN=%d\r\n", msg.Channel, p.PT(), p.CC(), p.M(), p.SN)
			buf := p.PL
			if buf != nil {
				err := nalsink.Push(buf, p.TS)
				if err != nil {
					log.Println(err)
					// log.Println(hex.Dump(buf))
					return
				}
				for _, nal := range nalsink.Units {
					log.Printf("NAL Zero=%t, RefIdc=%d, Type=%d, Size=%d\r\n", nal.ZeroBit(), nal.RefIdc(), nal.Type(), len(nal.Data))
					// TODO: Detect IDR (Type == 5)
					// TODO: Feed video packets to HLS/MP4/DASH emitter.
				}
			}
		}
	}
}

func rtcpHandler(ch chan rtsp.RawPacket) {
	for {
		select {
		case msg := <-ch:
			var p *rtcp.Packet
			var err error
			if p, err = rtcp.Unpack(msg.Payload); err != nil {
				log.Println(err)
				// log.Panicln(hex.Dump(msg.Payload))
				return
			}

			log.Printf("RTCP [%d] PT=%d, LN=%d, C=%d\r\n", msg.Channel, p.PT, p.LN, p.C())
		}
	}
}