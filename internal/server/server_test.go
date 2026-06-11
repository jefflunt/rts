package server

import (
	"net"
	"testing"
	"time"

	tilemap "scrollable-tilemap/internal/map"
	"scrollable-tilemap/internal/protocol"
)

func TestNewServer(t *testing.T) {
	addr := "127.0.0.1:0"
	s := NewServer(addr, ProtocolJSON)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.IsRunning() {
		t.Error("expected server to not be running initially")
	}
	if s.Protocol() != ProtocolJSON {
		t.Errorf("expected protocol %s, got %s", ProtocolJSON, s.Protocol())
	}
	if s.GetMap() == nil {
		t.Error("expected default tilemap to be generated")
	}

	// Test protocol fallback
	s2 := NewServer(addr, "invalid-protocol")
	if s2.Protocol() != ProtocolJSON {
		t.Errorf("expected protocol fallback to %s, got %s", ProtocolJSON, s2.Protocol())
	}

	s3 := NewServer(addr, ProtocolGob)
	if s3.Protocol() != ProtocolGob {
		t.Errorf("expected protocol %s, got %s", ProtocolGob, s3.Protocol())
	}
}

func TestServerStartStop(t *testing.T) {
	s := NewServer("127.0.0.1:0", ProtocolJSON)

	// Test double Stop when not started
	err := s.Stop()
	if err != nil {
		t.Errorf("expected nil error stopping non-running server, got %v", err)
	}

	// Start server
	err = s.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	if !s.IsRunning() {
		t.Error("expected server to be running after Start")
	}

	addr := s.Addr()
	if addr == nil {
		t.Fatal("expected server address to be non-nil after Start")
	}

	// Double start should fail
	err = s.Start()
	if err == nil {
		t.Error("expected error when starting an already running server, got nil")
	}

	// Stop server
	err = s.Stop()
	if err != nil {
		t.Errorf("failed to stop server: %v", err)
	}

	if s.IsRunning() {
		t.Error("expected server to not be running after Stop")
	}

	// Double Stop should succeed without error
	err = s.Stop()
	if err != nil {
		t.Errorf("expected nil error on second Stop, got %v", err)
	}
}

func TestServerServeWithCustomListener(t *testing.T) {
	s := NewServer("", ProtocolJSON)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	go func() {
		_ = s.Serve(l)
	}()

	// Wait briefly for server to start serving
	time.Sleep(50 * time.Millisecond)

	if !s.IsRunning() {
		t.Error("expected server to be running after Serve")
	}

	// Attempting Serve again on running server should error
	err = s.Serve(l)
	if err == nil {
		t.Error("expected error when calling Serve on already running server, got nil")
	}

	_ = s.Stop()
}

func TestServerJSONHandshakeAndSubscription(t *testing.T) {
	s := NewServer("127.0.0.1:0", ProtocolJSON)
	err := s.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer s.Stop()

	// Dial the server
	conn, err := net.Dial("tcp", s.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	// 1. Send JSON Handshake request
	req := protocol.NewHandshakeRequestPacket("1.0.0", "client-test-json")
	err = protocol.EncodeJSON(conn, req)
	if err != nil {
		t.Fatalf("failed to encode handshake request: %v", err)
	}

	// 2. Read JSON Handshake response
	resp, err := protocol.DecodeJSON(conn)
	if err != nil {
		t.Fatalf("failed to decode handshake response: %v", err)
	}

	if resp == nil || resp.Type != protocol.PacketTypeHandshakeResponse || resp.HandshakeResponse == nil {
		t.Fatalf("expected valid handshake response packet, got %+v", resp)
	}

	hr := resp.HandshakeResponse
	if !hr.Success {
		t.Errorf("handshake response indicates failure: %s", hr.Message)
	}
	if hr.ClientID != "client-test-json" {
		t.Errorf("expected client ID 'client-test-json', got '%s'", hr.ClientID)
	}
	if hr.MapWidth != tilemap.MapWidth || hr.MapHeight != tilemap.MapHeight {
		t.Errorf("unexpected map dimensions: %dx%d", hr.MapWidth, hr.MapHeight)
	}
	if hr.InitialMap == nil {
		t.Error("expected initial map to be included in successful response")
	}

	// 3. Send Viewport Subscription request
	sub := protocol.NewViewportSubscriptionPacket(10, 10, 12, 11) // 3x2 chunk
	err = protocol.EncodeJSON(conn, sub)
	if err != nil {
		t.Fatalf("failed to encode viewport subscription: %v", err)
	}

	// 4. Read Tilemap Chunk response
	chunkPacket, err := protocol.DecodeJSON(conn)
	if err != nil {
		t.Fatalf("failed to decode chunk response: %v", err)
	}

	if chunkPacket == nil || chunkPacket.Type != protocol.PacketTypeTilemapChunk || chunkPacket.Chunk == nil {
		t.Fatalf("expected valid tilemap chunk packet, got %+v", chunkPacket)
	}

	chunk := chunkPacket.Chunk
	if chunk.StartX != 10 || chunk.StartY != 10 || chunk.Width != 3 || chunk.Height != 2 {
		t.Errorf("unexpected chunk dimensions/coordinates: %+v", chunk)
	}
	if len(chunk.Tiles) != 6 {
		t.Errorf("expected 6 tiles, got %d", len(chunk.Tiles))
	}
}

func TestServerGobHandshakeAndSubscription(t *testing.T) {
	s := NewServer("127.0.0.1:0", ProtocolGob)
	err := s.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer s.Stop()

	// Dial the server
	conn, err := net.Dial("tcp", s.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	// 1. Send Gob Handshake request
	req := protocol.NewHandshakeRequestPacket("1.0.0", "client-test-gob")
	err = protocol.EncodeGob(conn, req)
	if err != nil {
		t.Fatalf("failed to encode handshake request: %v", err)
	}

	// 2. Read Gob Handshake response
	resp, err := protocol.DecodeGob(conn)
	if err != nil {
		t.Fatalf("failed to decode handshake response: %v", err)
	}

	if resp == nil || resp.Type != protocol.PacketTypeHandshakeResponse || resp.HandshakeResponse == nil {
		t.Fatalf("expected valid handshake response packet, got %+v", resp)
	}

	hr := resp.HandshakeResponse
	if !hr.Success {
		t.Errorf("handshake response indicates failure: %s", hr.Message)
	}
	if hr.ClientID != "client-test-gob" {
		t.Errorf("expected client ID 'client-test-gob', got '%s'", hr.ClientID)
	}
	if hr.InitialMap == nil {
		t.Error("expected initial map to be included in Gob response")
	}

	// 3. Send Viewport Subscription request
	sub := protocol.NewViewportSubscriptionPacket(20, 30, 24, 32) // 5x3 chunk
	err = protocol.EncodeGob(conn, sub)
	if err != nil {
		t.Fatalf("failed to encode viewport subscription: %v", err)
	}

	// 4. Read Tilemap Chunk response
	chunkPacket, err := protocol.DecodeGob(conn)
	if err != nil {
		t.Fatalf("failed to decode chunk response: %v", err)
	}

	if chunkPacket == nil || chunkPacket.Type != protocol.PacketTypeTilemapChunk || chunkPacket.Chunk == nil {
		t.Fatalf("expected valid tilemap chunk packet, got %+v", chunkPacket)
	}

	chunk := chunkPacket.Chunk
	if chunk.StartX != 20 || chunk.StartY != 30 || chunk.Width != 5 || chunk.Height != 3 {
		t.Errorf("unexpected chunk dimensions/coordinates: %+v", chunk)
	}
	if len(chunk.Tiles) != 15 {
		t.Errorf("expected 15 tiles, got %d", len(chunk.Tiles))
	}
}

func TestServerHandshakeInvalidRequest(t *testing.T) {
	s := NewServer("127.0.0.1:0", ProtocolJSON)
	err := s.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer s.Stop()

	// 1. Send totally invalid JSON
	conn, err := net.Dial("tcp", s.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))

	_, _ = conn.Write([]byte("{invalid json}\n"))

	// Server should send a HandshakeResponse indicating failure, then close the connection.
	resp, err := protocol.DecodeJSON(conn)
	if err != nil {
		t.Fatalf("failed to read handshake error response: %v", err)
	}
	if resp == nil || resp.HandshakeResponse == nil || resp.HandshakeResponse.Success {
		t.Errorf("expected failure handshake response, got: %+v", resp)
	}
	conn.Close()

	// 2. Send valid JSON but not a handshake request (e.g. viewport subscription)
	conn2, err := net.Dial("tcp", s.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn2.Close()
	_ = conn2.SetDeadline(time.Now().Add(1 * time.Second))

	sub := protocol.NewViewportSubscriptionPacket(0, 0, 5, 5)
	err = protocol.EncodeJSON(conn2, sub)
	if err != nil {
		t.Fatalf("failed to encode packet: %v", err)
	}

	resp2, err := protocol.DecodeJSON(conn2)
	if err != nil {
		t.Fatalf("failed to read handshake error response: %v", err)
	}
	if resp2 == nil || resp2.HandshakeResponse == nil || resp2.HandshakeResponse.Success {
		t.Errorf("expected failure handshake response for non-handshake first packet, got: %+v", resp2)
	}
}

func TestServerFallbackClientID(t *testing.T) {
	s := NewServer("127.0.0.1:0", ProtocolJSON)
	err := s.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer s.Stop()

	conn, err := net.Dial("tcp", s.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))

	// Send handshake with empty client ID
	req := protocol.NewHandshakeRequestPacket("1.0.0", "")
	_ = protocol.EncodeJSON(conn, req)

	resp, err := protocol.DecodeJSON(conn)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	if resp == nil || resp.HandshakeResponse == nil || !resp.HandshakeResponse.Success {
		t.Fatalf("expected successful handshake: %+v", resp)
	}
	// Fallback should be set to RemoteAddr
	if resp.HandshakeResponse.ClientID == "" {
		t.Error("expected non-empty fallback ClientID")
	}
}

func TestServerStopClosesClientsAndStopsAccepting(t *testing.T) {
	s := NewServer("127.0.0.1:0", ProtocolJSON)
	err := s.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	conn, err := net.Dial("tcp", s.Addr().String())
	if err != nil {
		s.Stop()
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))

	// Perform successful handshake
	req := protocol.NewHandshakeRequestPacket("1.0.0", "client-to-be-closed")
	_ = protocol.EncodeJSON(conn, req)
	_, _ = protocol.DecodeJSON(conn)

	// Now stop the server
	err = s.Stop()
	if err != nil {
		t.Fatalf("failed to stop server: %v", err)
	}

	// Try reading from the connection; it should return EOF (error) as server closed it
	_, err = protocol.DecodeJSON(conn)
	if err == nil {
		t.Error("expected EOF/closed error on client connection after server stops, got nil")
	}

	// Attempting to dial the server now should fail
	_, err = net.DialTimeout("tcp", s.Addr().String(), 50*time.Millisecond)
	if err == nil {
		t.Error("expected connection refusal after server has stopped, got nil")
	}
}

func TestServerViewportSubscriptionOutOfBounds(t *testing.T) {
	s := NewServer("127.0.0.1:0", ProtocolJSON)
	err := s.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer s.Stop()

	conn, err := net.Dial("tcp", s.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))

	// Handshake
	_ = protocol.EncodeJSON(conn, protocol.NewHandshakeRequestPacket("1.0.0", "c1"))
	_, _ = protocol.DecodeJSON(conn)

	// Viewport subscription with coordinates out of bounds (negative min coords)
	// Server should clamp coordinates:
	// minX < 0 becomes 0
	// minY < 0 becomes 0
	// maxX >= 256 becomes 255
	// maxY >= 256 becomes 255
	sub := protocol.NewViewportSubscriptionPacket(-5, -10, 300, 300)
	err = protocol.EncodeJSON(conn, sub)
	if err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	chunkPacket, err := protocol.DecodeJSON(conn)
	if err != nil {
		t.Fatalf("failed to read chunk packet: %v", err)
	}
	if chunkPacket == nil || chunkPacket.Chunk == nil {
		t.Fatal("expected chunk packet, got nil")
	}

	chunk := chunkPacket.Chunk
	if chunk.StartX != 0 || chunk.StartY != 0 || chunk.Width != tilemap.MapWidth || chunk.Height != tilemap.MapHeight {
		t.Errorf("unexpected clamped dimensions: start (%d,%d), size %dx%d", chunk.StartX, chunk.StartY, chunk.Width, chunk.Height)
	}
}

func TestServerConcurrentClients(t *testing.T) {
	s := NewServer("127.0.0.1:0", ProtocolJSON)
	err := s.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer s.Stop()

	const numClients = 10
	done := make(chan bool, numClients)

	for i := 0; i < numClients; i++ {
		go func(id int) {
			conn, err := net.Dial("tcp", s.Addr().String())
			if err != nil {
				done <- false
				return
			}
			defer conn.Close()

			_ = conn.SetDeadline(time.Now().Add(1 * time.Second))

			// Handshake
			req := protocol.NewHandshakeRequestPacket("1.0.0", "client-concurrent")
			if err := protocol.EncodeJSON(conn, req); err != nil {
				done <- false
				return
			}

			resp, err := protocol.DecodeJSON(conn)
			if err != nil || resp == nil || !resp.HandshakeResponse.Success {
				done <- false
				return
			}

			// Viewport Sub
			sub := protocol.NewViewportSubscriptionPacket(0, 0, 1, 1)
			if err := protocol.EncodeJSON(conn, sub); err != nil {
				done <- false
				return
			}

			chunkPacket, err := protocol.DecodeJSON(conn)
			if err != nil || chunkPacket == nil || chunkPacket.Chunk == nil {
				done <- false
				return
			}

			done <- true
		}(i)
	}

	for i := 0; i < numClients; i++ {
		select {
		case ok := <-done:
			if !ok {
				t.Error("a concurrent client connection test failed")
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for concurrent client response")
		}
	}
}
