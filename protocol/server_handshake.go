package protocol

import (
	"io"

	"github.com/alvin0319/go-stargate-server/util"
)

// ServerHandshake is a packet sent by server that indicates whether handshake process succeed or not.
type ServerHandshake struct {
	// Success indicates whether the handshake process succeed or not.
	Success bool
}

func (p *ServerHandshake) Read(r io.Reader) error {
	var err error
	p.Success, err = util.ReadBool(r)
	return err
}

func (p *ServerHandshake) Write(w io.Writer) error {
	return util.WriteBool(w, p.Success)
}

func (*ServerHandshake) ID() uint64 {
	return IDServerHandshake
}
