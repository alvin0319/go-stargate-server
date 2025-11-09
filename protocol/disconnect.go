package protocol

import (
	"io"

	"github.com/alvin0319/go-stargate-client/util"
)

const ReasonClientShutdown = "StarGate client shutdown!"

// Disconnect is a packet that notifies client/server about disconnection.
type Disconnect struct {
	// Reason is the reason why the client is disconnected.
	Reason string
}

func (p *Disconnect) Read(r io.Reader) error {
	reason, err := util.ReadString(r)
	if err != nil {
		return err
	}
	p.Reason = reason
	return nil
}

func (p *Disconnect) Write(w io.Writer) error {
	return util.WriteString(w, p.Reason)
}

func (*Disconnect) ID() uint64 {
	return IDDisconnect
}
