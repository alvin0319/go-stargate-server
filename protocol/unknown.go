package protocol

import "io"

// Unknown is the packet that is not known to the client.
// It contains the packet ID and payload.
type Unknown struct {
	// PacketID is the ID of this packet.
	PacketID uint64
	// Payload contains the raw payload of this packet.
	Payload []byte
}

func (p *Unknown) Read(r io.Reader) error {
	// Read all remaining data as payload
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	p.Payload = data
	return nil
}

func (p *Unknown) Write(w io.Writer) error {
	_, err := w.Write(p.Payload)
	return err
}

func (*Unknown) ID() uint64 {
	return IDUnknown
}
