package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/alvin0319/go-stargate-server/protocol"
	"github.com/alvin0319/go-stargate-server/protocol/types"
)

const (
	StateAuthenticating = iota
	StateConnected
	StateDisconnected
)

const (
	PingInterval = 30 * time.Second
	PingTimeout  = 5 * time.Second
	TickInterval = 50 * time.Millisecond
)

const (
	StarGateMagic = 0x0a20
)

// Conn represents the connection of TCP server.
type Conn struct {
	net.Conn

	// Name is the name of this client. This will be empty until fully authenticated.
	Name string

	// queuedPackets contains the packets queued in.
	// This will be sent to the connection on next available tick.
	queuedPackets []*protocol.Wrapper

	// state is a state where the current connection is in.
	state int

	closed chan struct{}

	logger *slog.Logger

	handshakeData *types.HandshakeData

	password string

	lastPingTime    time.Time
	lastPongTime    time.Time
	pingPending     bool
	pingTimeoutChan chan struct{}

	h Handler

	bufReader *bufio.Reader

	listener *Listener
}

func newConn(conn net.Conn, password string, listener *Listener) *Conn {
	logger := slog.Default().With("addr", conn.RemoteAddr())

	c := &Conn{
		Name:     "",
		Conn:     conn,
		state:    StateAuthenticating,
		listener: listener,

		queuedPackets: make([]*protocol.Wrapper, 0),
		closed:        make(chan struct{}),

		logger:   logger,
		password: password,

		lastPongTime:    time.Now(),
		pingTimeoutChan: make(chan struct{}, 1),

		bufReader: bufio.NewReader(conn),
	}
	c.logger.Info("new connection established", "state", "authenticating")
	return c
}

func (c *Conn) State() int {
	return c.state
}

func (c *Conn) QueuePacket(w *protocol.Wrapper) {
	c.queuedPackets = append(c.queuedPackets, w)
}

func (c *Conn) tick() {
	c.logger.Debug("starting tick loop")
	ticker := time.NewTicker(TickInterval)
	defer ticker.Stop()

	readChan := make(chan *protocol.Wrapper)
	errChan := make(chan error)

	go c.readLoop(readChan, errChan)

	for {
		select {
		case <-c.closed:
			return
		case <-ticker.C:
			c.onTick()
		case wrapper := <-readChan:
			c.handlePacket(wrapper)
		case err := <-errChan:
			if err == io.EOF {
				c.logger.Info("connection closed")
			} else {
				c.logger.Warn("failed to read packet", "err", err)
			}
			c.closeConn()
			return
		case <-c.pingTimeoutChan:
			c.logger.Warn("ping timeout, closing connection")
			c.DisconnectAndClose("Ping timeout")
			return
		}
	}
}

func (c *Conn) readLoop(readChan chan *protocol.Wrapper, errChan chan error) {
	c.logger.Debug("starting read loop")

	peekBytes, err := c.bufReader.Peek(16)
	if err == nil {
		c.logger.Debug("first 16 bytes from client", "hex", fmt.Sprintf("% x", peekBytes))
	}

	for {
		p, err := c.ReadPacket()
		if err != nil {
			c.logger.Debug("read error", "err", err)
			errChan <- err
			return
		}
		c.logger.Debug("packet read", "id", p.P.ID())
		readChan <- p
	}
}

func (c *Conn) onTick() {
	if len(c.queuedPackets) > 0 {
		wrapper := c.queuedPackets[0]
		c.queuedPackets = c.queuedPackets[1:]

		// Build payload
		var payloadBuf bytes.Buffer

		payloadBuf.WriteByte(byte(wrapper.P.ID()))

		if wrapper.Response {
			payloadBuf.WriteByte(1)
			responseIDBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(responseIDBytes, uint32(wrapper.ResponseID))
			payloadBuf.Write(responseIDBytes)
		} else {
			payloadBuf.WriteByte(0)
		}

		if err := wrapper.P.Write(&payloadBuf); err != nil {
			c.logger.Error("failed to marshal packet", "err", err)
			c.closeConn()
			return
		}

		var finalBuf bytes.Buffer

		// Write magic (2 bytes, big-endian)
		if err := binary.Write(&finalBuf, binary.BigEndian, uint16(StarGateMagic)); err != nil {
			c.logger.Error("failed to write magic", "err", err)
			c.closeConn()
			return
		}

		// Write length (4 bytes, big-endian)
		if err := binary.Write(&finalBuf, binary.BigEndian, uint32(payloadBuf.Len())); err != nil {
			c.logger.Error("failed to write length", "err", err)
			c.closeConn()
			return
		}

		// Write payload
		finalBuf.Write(payloadBuf.Bytes())

		if _, err := c.Write(finalBuf.Bytes()); err != nil {
			c.logger.Error("failed to write packet", "err", err)
			c.closeConn()
			return
		}
	}

	if c.state == StateConnected {
		now := time.Now()

		if c.pingPending {
			if now.Sub(c.lastPingTime) >= PingTimeout {
				select {
				case c.pingTimeoutChan <- struct{}{}:
				default:
				}
			}
		} else {
			if now.Sub(c.lastPongTime) >= PingInterval {
				c.sendPing()
			}
		}
	}
}

func (c *Conn) sendPing() {
	now := time.Now()
	c.lastPingTime = now
	c.pingPending = true

	c.QueuePacket(&protocol.Wrapper{
		P: &protocol.Ping{PingTime: now.UnixMilli()},
	})
}

func (c *Conn) handlePacket(wrapper *protocol.Wrapper) {
	switch c.state {
	case StateAuthenticating:
		if wrapper.P.ID() == protocol.IDHandshake {
			handshake := wrapper.P.(*protocol.Handshake)
			c.handshakeData = &handshake.Data
			c.logger.Info("received handshake", "client", handshake.Data.ClientName, "software", handshake.Data.Software, "protocol", handshake.Data.Protocol)

			success := handshake.Data.Password == c.password
			if !success {
				c.logger.Warn("invalid password", "client", handshake.Data.ClientName)
			}

			c.QueuePacket(&protocol.Wrapper{
				P: &protocol.ServerHandshake{Success: success},
			})

			if success {
				c.state = StateConnected
				c.Name = handshake.Data.ClientName
				c.logger = c.logger.With("conn", handshake.Data.ClientName)
				c.logger.Info("authenticated")
			} else {
				go func() {
					time.Sleep(100 * time.Millisecond)
					c.DisconnectAndClose("Invalid password")
				}()
			}
		} else {
			c.logger.Warn("unexpected packet during authentication", "packetID", wrapper.P.ID(), "expected", protocol.IDHandshake)
		}
	case StateConnected:
		if c.h != nil {
			if err := c.h.Handle(wrapper); err != nil {
				c.logger.Error("failed to handle packet", "err", err)
			}
		}
		switch wrapper.P.ID() {
		case protocol.IDDisconnect:
			disconnect := wrapper.P.(*protocol.Disconnect)
			c.logger.Info("disconnected", "reason", disconnect.Reason)
			c.closeConn()
		case protocol.IDPing:
			ping := wrapper.P.(*protocol.Ping)
			c.QueuePacket(&protocol.Wrapper{
				P: &protocol.Pong{PingTime: ping.PingTime},
			})
		case protocol.IDPong:
			c.pingPending = false
			c.lastPongTime = time.Now()
			c.logger.Debug("received pong from client", "pingTime", wrapper.P.(*protocol.Pong).PingTime, "latency", time.Since(time.UnixMilli(wrapper.P.(*protocol.Pong).PingTime)))
		case protocol.IDUnknown:
			unknown := wrapper.P.(*protocol.Unknown)
			c.logger.Warn("received unknown packet", "id", unknown.PacketID)
		}
	}
}

// DisconnectAndClose sends a disconnect packet with the given reason and closes the connection.
func (c *Conn) DisconnectAndClose(reason string) {
	// Only disconnect if not already closed
	select {
	case <-c.closed:
		return
	default:
	}

	c.logger.Info("disconnecting", "reason", reason)

	// Queue disconnect packet
	c.QueuePacket(&protocol.Wrapper{
		P: &protocol.Disconnect{
			Reason: reason,
		},
		Response: false,
	})

	// Flush all queued packets (including the disconnect packet)
	for len(c.queuedPackets) > 0 {
		c.onTick()
	}

	// Now close the connection
	c.closeConn()
}

func (c *Conn) closeConn() {
	select {
	case <-c.closed:
		return
	default:
		close(c.closed)
	}

	c.state = StateDisconnected
	c.logger.Info("closing connection")

	// Remove from listener's connection tracking
	if c.listener != nil {
		c.listener.removeConnection(c)
	}

	c.Conn.Close()
}

func (c *Conn) HandshakeData() *types.HandshakeData {
	return c.handshakeData
}

// ReadPacket reads a single packet from the connection.
func (c *Conn) ReadPacket() (*protocol.Wrapper, error) {
	// Read magic (2 bytes, big-endian)
	var magic uint16
	if err := binary.Read(c.bufReader, binary.BigEndian, &magic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if magic != StarGateMagic {
		return nil, fmt.Errorf("invalid magic: expected 0x%04x, got 0x%04x", StarGateMagic, magic)
	}
	c.logger.Debug("read magic", "magic", fmt.Sprintf("0x%04x", magic))

	// Read length (4 bytes, big-endian)
	var length uint32
	if err := binary.Read(c.bufReader, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read length: %w", err)
	}
	c.logger.Debug("read payload length", "length", length)

	if length == 0 || length > 1024*1024 {
		return nil, fmt.Errorf("invalid payload length: %d", length)
	}

	// Read payload
	payloadData := make([]byte, length)
	if _, err := io.ReadFull(c.bufReader, payloadData); err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}
	c.logger.Debug("full payload", "hex", fmt.Sprintf("% x", payloadData))

	// Parse payload
	payloadBuf := bytes.NewReader(payloadData)

	packetID, err := payloadBuf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read packet ID: %w", err)
	}
	c.logger.Debug("read raw packet ID", "packetID", packetID, "hex", fmt.Sprintf("0x%02x", packetID))

	responseByte, err := payloadBuf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read response flag: %w", err)
	}
	c.logger.Debug("read response byte", "responseByte", responseByte, "isResponse", responseByte != 0)

	wrapper := &protocol.Wrapper{
		Response: responseByte != 0,
	}

	if wrapper.Response {
		responseIDBytes := make([]byte, 4)
		if _, err := io.ReadFull(payloadBuf, responseIDBytes); err != nil {
			return nil, fmt.Errorf("failed to read response ID: %w", err)
		}
		wrapper.ResponseID = uint(binary.BigEndian.Uint32(responseIDBytes))
		c.logger.Debug("read response ID", "responseID", wrapper.ResponseID)
	}

	constructor, ok := protocol.Pool[uint64(packetID)]
	if !ok {
		c.logger.Warn("unknown packet ID in pool", "packetID", packetID)
		unknownPacket := &protocol.Unknown{PacketID: uint64(packetID)}
		if err := unknownPacket.Read(payloadBuf); err != nil {
			return nil, fmt.Errorf("failed to read unknown packet: %w", err)
		}
		c.logger.Debug("unknown packet payload", "payloadLen", len(unknownPacket.Payload), "payload", fmt.Sprintf("%x", unknownPacket.Payload))
		return &protocol.Wrapper{
			P:          unknownPacket,
			Response:   wrapper.Response,
			ResponseID: wrapper.ResponseID,
		}, nil
	}

	c.logger.Debug("found packet constructor", "packetID", packetID)
	packet := constructor()
	if err := packet.Read(payloadBuf); err != nil {
		return nil, fmt.Errorf("failed to read packet: %w", err)
	}

	wrapper.P = packet
	return wrapper, nil
}

// Handler sets packet handler for this connection.
func (c *Conn) Handler(h Handler) {
	c.h = h
}
