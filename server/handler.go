package server

import "github.com/alvin0319/go-stargate-client/protocol"

type Handler interface {
	// Handle handles the wrapper packet.
	Handle(w *protocol.Wrapper) error
}

type DefaultHandler struct {
	Handler
}

func (h DefaultHandler) Handle(w *protocol.Wrapper) error {
	return nil
}
