package protocol

import (
	"io"

	"github.com/alvin0319/go-stargate-server/util"
)

// ServerTransfer is a packet to request a transfer of given player.
type ServerTransfer struct {
	// PlayerName is the name of player.
	PlayerName string
	// TargetServer is the name of target server.
	TargetServer string
}

func (p *ServerTransfer) Read(r io.Reader) error {
	var err error

	p.PlayerName, err = util.ReadString(r)
	if err != nil {
		return err
	}

	p.TargetServer, err = util.ReadString(r)
	if err != nil {
		return err
	}

	return nil
}

func (p *ServerTransfer) Write(w io.Writer) error {
	if err := util.WriteString(w, p.PlayerName); err != nil {
		return err
	}

	return util.WriteString(w, p.TargetServer)
}

func (*ServerTransfer) ID() uint64 {
	return IDServerTransfer
}
