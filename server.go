package main

import (
	"bufio"
	"net"
	"os"
	"os/signal"
	"h264rtp/payloads/h264"
	"github.com/evandbrown/gortp"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("logger")

var port = 5220
var addr, _ = net.ResolveIPAddr("ip", "127.0.0.1")

var rSession *rtp.Session

func receivePacket() {
	// Create and store the data receive channel.
	dataReceiver := rSession.CreateDataReceiveChan()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create file and H264 handler
	f, err := os.Create("/tmp/video.h264")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := h264.NewH264Processor(bufio.NewWriter(f))
	for {
		select {
		case rp := <-dataReceiver:
			//log.Info("Rec'd Packet")
			p.Process(rp)
		case <-interrupt:
			log.Info("Closing channel and file...")
			p.Close()
			f.Close()
			return
		}
	}
}
func main() {
	// Setup transport and RTP Session
	transport, _ := rtp.NewTransportUDP(addr, port)
	rSession = rtp.NewSession(transport, transport)
	rSession.ListenOnTransports()
	receivePacket()
}
