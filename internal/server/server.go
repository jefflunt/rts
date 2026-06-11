package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	tilemap "scrollable-tilemap/internal/map"
	"scrollable-tilemap/internal/protocol"
)

// ProtocolType defines the serialization format used by the server.
type ProtocolType string

const (
	ProtocolJSON ProtocolType = "json"
	ProtocolGob  ProtocolType = "gob"
)

// Server is an authoritative TCP server hosting and streaming the 256x256 grass tilemap.
type Server struct {
	address  string
	protocol ProtocolType
	tilemap  *tilemap.Map

	listener net.Listener
	clients  map[string]net.Conn
	mu       sync.Mutex
	wg       sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	running bool
}

// NewServer creates a new Server instance with a default 256x256 grass tilemap.
func NewServer(address string, proto ProtocolType) *Server {
	return NewServerWithMap(address, proto, tilemap.NewDefaultMap())
}

// NewServerWithMap creates a new Server instance with a pre-configured tilemap.
func NewServerWithMap(address string, proto ProtocolType, m *tilemap.Map) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	if proto != ProtocolGob {
		proto = ProtocolJSON
	}
	return &Server{
		address:  address,
		protocol: proto,
		tilemap:  m,
		clients:  make(map[string]net.Conn),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the server on the configured address.
// This is a non-blocking call that runs the acceptance loop in a background goroutine.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("server is already running")
	}

	l, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}
	s.listener = l
	s.running = true

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		_ = s.serve(l)
	}()

	return nil
}

// Serve runs a blocking loop to accept incoming connections on the provided listener.
// This is useful for injecting custom listeners for testing.
func (s *Server) Serve(l net.Listener) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("server is already running")
	}
	s.listener = l
	s.running = true
	s.mu.Unlock()

	return s.serve(l)
}

// serve is the core loop that accepts client connections.
func (s *Server) serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				return err
			}
		}

		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			s.handleClient(c)
		}(conn)
	}
}

// Stop shuts down the server, closing the listener and all active client connections.
func (s *Server) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.cancel()

	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}

	// Close all active client connections
	for id, conn := range s.clients {
		_ = conn.Close()
		delete(s.clients, id)
	}
	s.mu.Unlock()

	s.wg.Wait()
	return err
}

// Addr returns the listening address of the server, or nil if the server is not running.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}

// GetMap returns the authoritative map of the server.
func (s *Server) GetMap() *tilemap.Map {
	return s.tilemap
}

// Protocol returns the configured protocol of the server.
func (s *Server) Protocol() ProtocolType {
	return s.protocol
}

// IsRunning returns whether the server is currently running.
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// handleClient manages the handshake and main protocol loop for an individual connection.
func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	// 1. Manage Handshake
	var reqPacket *protocol.Packet
	var err error

	if s.protocol == ProtocolGob {
		reqPacket, err = protocol.DecodeGob(conn)
	} else {
		reqPacket, err = protocol.DecodeJSON(conn)
	}

	if err != nil {
		_ = s.sendHandshakeResponse(conn, false, "failed to decode handshake request: "+err.Error(), "")
		return
	}

	if reqPacket == nil || reqPacket.Type != protocol.PacketTypeHandshakeRequest || reqPacket.HandshakeRequest == nil {
		_ = s.sendHandshakeResponse(conn, false, "invalid handshake request", "")
		return
	}

	clientID := reqPacket.HandshakeRequest.ClientID
	if clientID == "" {
		clientID = conn.RemoteAddr().String()
	}

	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	// Add connection to tracked clients
	s.clients[clientID] = conn
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientID)
		s.mu.Unlock()
	}()

	err = s.sendHandshakeResponse(conn, true, "Handshake successful", clientID)
	if err != nil {
		return
	}

	// 2. Main Protocol Loop: Stream/Listen on-demand
	for {
		var p *protocol.Packet
		if s.protocol == ProtocolGob {
			p, err = protocol.DecodeGob(conn)
		} else {
			p, err = protocol.DecodeJSON(conn)
		}

		if err != nil {
			// Connection closed or EOF
			break
		}

		if p == nil {
			continue
		}

		switch p.Type {
		case protocol.PacketTypeViewportSubscription:
			if p.Subscription != nil {
				sub := p.Subscription
				minX, minY := sub.MinTileX, sub.MinTileY
				maxX, maxY := sub.MaxTileX, sub.MaxTileY

				if minX < 0 {
					minX = 0
				}
				if minY < 0 {
					minY = 0
				}
				if maxX >= tilemap.MapWidth {
					maxX = tilemap.MapWidth - 1
				}
				if maxY >= tilemap.MapHeight {
					maxY = tilemap.MapHeight - 1
				}

				if maxX >= minX && maxY >= minY {
					width := maxX - minX + 1
					height := maxY - minY + 1
					chunk, err := protocol.NewTilemapChunk(minX, minY, width, height, s.tilemap)
					if err == nil {
						chunkPacket := protocol.NewTilemapChunkPacket(chunk)
						if s.protocol == ProtocolGob {
							_ = protocol.EncodeGob(conn, chunkPacket)
						} else {
							_ = protocol.EncodeJSON(conn, chunkPacket)
						}
					}
				}
			}
		}
	}
}

// sendHandshakeResponse constructs and sends a HandshakeResponse packet.
func (s *Server) sendHandshakeResponse(conn net.Conn, success bool, message string, clientID string) error {
	var m *tilemap.Map
	if success {
		m = s.tilemap
	}

	resp := protocol.NewHandshakeResponsePacket(
		success,
		message,
		clientID,
		tilemap.MapWidth,
		tilemap.MapHeight,
		tilemap.TileSize,
		m,
	)

	if s.protocol == ProtocolGob {
		return protocol.EncodeGob(conn, resp)
	}
	return protocol.EncodeJSON(conn, resp)
}
