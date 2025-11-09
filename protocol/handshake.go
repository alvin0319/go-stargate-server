package protocol

import (
	"io"

	"github.com/alvin0319/go-stargate-server/protocol/types"
)

// Handshake ...
type Handshake struct {
	// Data is the handshake data of this packet.
	Data types.HandshakeData
}

func (p *Handshake) Read(r io.Reader) error {
	return p.Data.Read(r)
}

func (p *Handshake) Write(w io.Writer) error {
	return p.Data.Write(w)
}

func (Handshake) ID() uint64 {
	return IDHandshake
}
