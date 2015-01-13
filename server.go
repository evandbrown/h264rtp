package main

import (
	"fmt"
	"bufio"
	"github.com/evandbrown/gortp"
	"github.com/evandbrown/rtp-receiver/payloads/h264"
	"net"
	"os"
	"os/signal"
)

var port = 5220
var addr, _ = net.ResolveIPAddr("ip", "127.0.0.1")

var rSession *rtp.Session

var stop bool
var stopRecv chan bool
var stopCtrl chan bool

var eventNamesNew = []string{"NewStreamData", "NewStreamCtrl"}
var eventNamesRtcp = []string{"SR", "RR", "SDES", "BYE"}



/**func processNAL(nalu h264.NALU, f *os.File) {
	switch {
	case nalu.NUT() <= 23:
		writePacket(nalu.Payload(), f)
	case nalu.NUT() == 28: //FU-A
		fmt.Println("----------------------------------------")
		fua := h264.FUAUnit{&nalu}
		fmt.Println(fua)
		if len(fuaBuff) == 0 { // Append the first fragment
			fmt.Println("First packet")
			fuaBuff = make([]h264.FUAUnit, 0)
			fuaBuff = append(fuaBuff, fua)
		} else {
			if nalu.Seq()-fuaBuff[len(fuaBuff)-1].Seq() != 1 {
				// RTP Sequence Numbers must be sequential
				fuaBuff = nil
			} else {
				fuaBuff = append(fuaBuff, fua)
			}
		}
		fmt.Printf("H1: %b\n  ", fua.Payload()[1])
		fmt.Printf("Start? %v\n  ", fua.Start())
		fmt.Printf("End? %v\n", fua.End())
		fmt.Printf("Type? %v\n", fua.PayNUT())
		if fua.End() {
			fmt.Println("Wrote FU-A")
			writePacket(fua.Payload(), f)
			fuaBuff = nil
		}
	default:
		fmt.Println("Found some weird packet")
	}
}
**/

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

	h := h264.NewHandler(bufio.NewWriter(f))

	for {
		fmt.Println("Waiting for a packet...")
		select {
		case rp := <-dataReceiver:
			fmt.Println(".")
			h.HandlePkt(rp)
			rp.FreePacket()
		case <-stopRecv:
			fmt.Println(".")
			return
		case <-interrupt:
			fmt.Println(".")
			fmt.Println("Closing file")
			f.Close()
			return
		}
	}
}

func initialize() {
	stopRecv = make(chan bool, 1)
	stopCtrl = make(chan bool, 1)
}

func main() {
	initialize()

	// Setup transport and RTP Session
	transport, _ := rtp.NewTransportUDP(addr, port)
	rSession = rtp.NewSession(transport, transport)
	rSession.ListenOnTransports()
	receivePacket()
}
