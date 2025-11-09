package types

import (
	"io"

	"github.com/alvin0319/go-stargate-server/util"
)

const (
	SoftwarePocketMine = iota
	SoftwarePM5
)

type HandshakeData struct {
	// ClientName is the name of this client.
	// The server will use this client name as the identifier of connection.
	ClientName string
	// Password is the user-input password of connection.
	// The server will deny connection if password does not match.
	Password string
	// Software is the name of software client is using.
	// It is either one of constants obove. or can be other.
	Software int32
	// Protocol is the version of protocol client uses.
	Protocol int32
}

func (d *HandshakeData) Read(r io.Reader) error {
	var err error

	d.Software, err = util.ReadInt32(r)
	if err != nil {
		return err
	}

	d.ClientName, err = util.ReadString(r)
	if err != nil {
		return err
	}

	d.Password, err = util.ReadString(r)
	if err != nil {
		return err
	}

	d.Protocol, err = util.ReadInt32(r)
	if err != nil {
		return err
	}

	return nil
}

func (d *HandshakeData) Write(w io.Writer) error {
	if err := util.WriteInt32(w, d.Software); err != nil {
		return err
	}

	if err := util.WriteString(w, d.ClientName); err != nil {
		return err
	}

	if err := util.WriteString(w, d.Password); err != nil {
		return err
	}

	if err := util.WriteInt32(w, d.Protocol); err != nil {
		return err
	}

	return nil
}
