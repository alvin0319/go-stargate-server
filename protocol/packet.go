package protocol

import "io"

// Pool is a map of packet IDs to a function that returns a new packet of that type.
var Pool = map[uint64]func() Packet{
	IDHandshake:       func() Packet { return &Handshake{} },
	IDServerHandshake: func() Packet { return &ServerHandshake{} },
	IDDisconnect:      func() Packet { return &Disconnect{} },
	IDPing:            func() Packet { return &Ping{} },
	IDPong:            func() Packet { return &Pong{} },
	IDServerTransfer:  func() Packet { return &ServerTransfer{} },
}

// Packet is a interface that can read and write packet data.
type Packet interface {
	// Read reads the packet payload from the reader.
	Read(r io.Reader) error

	// Write writes the packet payload to the writer.
	Write(w io.Writer) error

	// ID returns the ID of this packet.
	ID() uint64
}
