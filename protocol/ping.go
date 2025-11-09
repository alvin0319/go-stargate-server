package protocol

import (
	"io"

	"github.com/alvin0319/go-stargate-client/util"
)

type Ping struct {
	PingTime int64
}

func (p *Ping) Read(r io.Reader) error {
	var err error
	p.PingTime, err = util.ReadInt64(r)
	return err
}

func (p *Ping) Write(w io.Writer) error {
	return util.WriteInt64(w, p.PingTime)
}

func (*Ping) ID() uint64 {
	return IDPing
}
