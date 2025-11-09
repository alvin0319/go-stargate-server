package protocol

// Wrapper is a wrapper struct for packet which contains Packet, ResponseID (if it contains response)
type Wrapper struct {
	// P is the Packet itself.
	P Packet
	// Response indicates whether this packet is response.
	Response bool
	// ResponseID is the ID of response if Response is true.
	ResponseID uint
}
