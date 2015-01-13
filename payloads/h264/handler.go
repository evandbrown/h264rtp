package h264

import (
	"bufio"
	"github.com/evandbrown/gortp"
	)

type Handler interface {
	Close() error
	HandlePkt(*rtp.DataPacket) error
}

type H264Handler struct {
	writer *bufio.Writer
	singleUnit chan SingleUnit
	stop chan bool
}

func (u *H264Handler) Close() error {
	u.writer.Flush()
	return nil
}

func (u *H264Handler) HandlePkt(p *rtp.DataPacket) error {
	n := FromRTP(p)
	switch {
	case n.NUT() <= 23:

		u.singleUnit <- SingleUnit{n}
	}
	return nil
}

func (u *H264Handler) naluWriter(singleUnit chan SingleUnit, stop chan bool) {
	select {
	case nalu := <-singleUnit:
		u.writer.Write([]byte{0x00, 0x00, 0x01})
		u.writer.Write(nalu.Payload())
		u.writer.Flush()
	case <-stop:
		u.writer.Flush()
	}
}

func NewHandler(w *bufio.Writer) Handler {
		singleUnit := make(chan SingleUnit)
		stop := make(chan bool)
		handler := &H264Handler{writer: w, singleUnit: singleUnit, stop: stop}
		go handler.naluWriter(handler.singleUnit, handler.stop)
		return handler
}
