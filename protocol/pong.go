package protocol

import (
	"io"

	"github.com/alvin0319/go-stargate-client/util"
)

// Pong is a packet that responds to Ping.
type Pong struct {
	// PingTime is the time where the Ping is sent.
	PingTime int64
}

func (p *Pong) Read(r io.Reader) error {
	var err error
	p.PingTime, err = util.ReadInt64(r)
	return err
}

func (p *Pong) Write(w io.Writer) error {
	return util.WriteInt64(w, p.PingTime)
}

func (*Pong) ID() uint64 {
	return IDPong
}
