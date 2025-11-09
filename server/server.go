package server

import (
	"net"
	"sync"
)

type Listener struct {
	incoming chan *Conn
	close    chan struct{}
	password string

	mu          sync.RWMutex
	connections map[*Conn]struct{}

	listener net.Listener
}

// Accept accepts *Conn from the Listener.
func (l *Listener) Accept() *Conn {
	c := <-l.incoming
	return c
}

// Close closes the server and gracefully disconnects all clients.
func (l *Listener) Close() {
	// Stop accepting new connections
	l.listener.Close()

	// Close all existing connections with disconnect packet
	l.mu.RLock()
	conns := make([]*Conn, 0, len(l.connections))
	for conn := range l.connections {
		conns = append(conns, conn)
	}
	l.mu.RUnlock()

	// Disconnect all clients with proper message
	for _, conn := range conns {
		conn.DisconnectAndClose("StarGate server shutdown")
	}

	// Signal shutdown
	close(l.close)
}

// addConnection adds a connection to the tracking map.
func (l *Listener) addConnection(c *Conn) {
	l.mu.Lock()
	l.connections[c] = struct{}{}
	l.mu.Unlock()
}

// removeConnection removes a connection from the tracking map.
func (l *Listener) removeConnection(c *Conn) {
	l.mu.Lock()
	delete(l.connections, c)
	l.mu.Unlock()
}

// Listen binds the TCP server on specified addr.
func Listen(addr, password string) (*Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	listener := &Listener{
		incoming:    make(chan *Conn),
		close:       make(chan struct{}),
		password:    password,
		connections: make(map[*Conn]struct{}),
		listener:    l,
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-listener.close:
					return
				default:
					continue
				}
			}
			c := newConn(conn, listener.password, listener)
			listener.addConnection(c)
			go c.tick()
			listener.incoming <- c
		}
	}()
	return listener, nil
}

// Connections returns a copied slice of connections.
func (l *Listener) Connections() []*Conn {
	l.mu.RLock()
	defer l.mu.RUnlock()
	conns := make([]*Conn, 0, len(l.connections))
	for conn := range l.connections {
		conns = append(conns, conn)
	}
	return conns
}
