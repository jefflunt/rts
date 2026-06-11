package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"scrollable-tilemap/internal/map"
	"scrollable-tilemap/internal/protocol"
)

type mockServerRunner struct {
	started  bool
	stopped  bool
	startErr error
	stopErr  error
}

func (m *mockServerRunner) Start() error {
	if m.startErr != nil {
		return m.startErr
	}
	m.started = true
	return nil
}

func (m *mockServerRunner) Stop() error {
	if m.stopErr != nil {
		return m.stopErr
	}
	m.stopped = true
	return nil
}

func TestNewDefaultAppRunner(t *testing.T) {
	runner := NewDefaultAppRunner()
	if runner == nil {
		t.Fatal("expected runner to be non-nil")
	}
	if runner.DialFn == nil {
		t.Error("expected DialFn to be initialized")
	}
	if runner.RunGameFn == nil {
		t.Error("expected RunGameFn to be initialized")
	}
	if runner.NewServerFn == nil {
		t.Error("expected NewServerFn to be initialized")
	}
}

func TestAppRunner_Run_InvalidMode(t *testing.T) {
	runner := NewDefaultAppRunner()
	err := runner.Run("invalid_mode", "127.0.0.1:8080", "json")
	if err == nil {
		t.Fatal("expected error with invalid mode, got nil")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("expected 'invalid mode' error message, got: %v", err)
	}
}

func TestRunServerMode_Success(t *testing.T) {
	mockServer := &mockServerRunner{}
	runner := &AppRunner{
		NewServerFn: func(addr string, proto string) ServerRunner {
			return mockServer
		},
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- runner.RunServerMode("127.0.0.1:0", "json")
	}()

	// Wait for the goroutine to block on signal.Notify
	time.Sleep(50 * time.Millisecond)

	// Send SIGTERM to ourselves to trigger signal reception
	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("expected no error from RunServerMode, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("RunServerMode timed out waiting for signal")
	}

	if !mockServer.started {
		t.Error("expected server to be started")
	}
	if !mockServer.stopped {
		t.Error("expected server to be stopped")
	}
}

func TestRunServerMode_StartError(t *testing.T) {
	mockServer := &mockServerRunner{startErr: fmt.Errorf("start failed")}
	runner := &AppRunner{
		NewServerFn: func(addr string, proto string) ServerRunner {
			return mockServer
		},
	}

	err := runner.RunServerMode("127.0.0.1:0", "json")
	if err == nil {
		t.Error("expected error on start failure, got nil")
	}
	if !strings.Contains(err.Error(), "failed to start server") {
		t.Errorf("expected start failure error message, got: %v", err)
	}
}

func TestRunServerMode_StopError(t *testing.T) {
	mockServer := &mockServerRunner{stopErr: fmt.Errorf("stop failed")}
	runner := &AppRunner{
		NewServerFn: func(addr string, proto string) ServerRunner {
			return mockServer
		},
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- runner.RunServerMode("127.0.0.1:0", "json")
	}()

	time.Sleep(50 * time.Millisecond)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	select {
	case err := <-errChan:
		if err == nil {
			t.Error("expected error on stop failure, got nil")
		}
		if !strings.Contains(err.Error(), "failed to stop server") {
			t.Errorf("expected stop failure error message, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("RunServerMode timed out waiting for signal")
	}
}

func TestRunClientMode_Success_JSON(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	runner := &AppRunner{
		DialFn: func(network, address string) (net.Conn, error) {
			return c1, nil
		},
		RunGameFn: func(game ebiten.Game) error {
			return nil
		},
	}

	go func() {
		req, err := protocol.DecodeJSON(c2)
		if err != nil {
			return
		}
		if req.Type != protocol.PacketTypeHandshakeRequest {
			return
		}

		m := tilemap.NewDefaultMap()
		resp := protocol.NewHandshakeResponsePacket(true, "welcome", "client-1", tilemap.MapWidth, tilemap.MapHeight, tilemap.TileSize, m)
		_ = protocol.EncodeJSON(c2, resp)
	}()

	err := runner.RunClientMode("127.0.0.1:8080", "json")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunClientMode_Success_Gob(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	runner := &AppRunner{
		DialFn: func(network, address string) (net.Conn, error) {
			return c1, nil
		},
		RunGameFn: func(game ebiten.Game) error {
			return nil
		},
	}

	go func() {
		req, err := protocol.DecodeGob(c2)
		if err != nil {
			return
		}
		if req.Type != protocol.PacketTypeHandshakeRequest {
			return
		}

		m := tilemap.NewDefaultMap()
		resp := protocol.NewHandshakeResponsePacket(true, "welcome", "client-1", tilemap.MapWidth, tilemap.MapHeight, tilemap.TileSize, m)
		_ = protocol.EncodeGob(c2, resp)
	}()

	err := runner.RunClientMode("127.0.0.1:8080", "gob")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunClientMode_DialError(t *testing.T) {
	runner := &AppRunner{
		DialFn: func(network, address string) (net.Conn, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	err := runner.RunClientMode("127.0.0.1:8080", "json")
	if err == nil {
		t.Fatal("expected error on dial failure, got nil")
	}
	if !strings.Contains(err.Error(), "failed to connect to server") {
		t.Errorf("expected connect failure message, got: %v", err)
	}
}

func TestRunClientMode_HandshakeFailed(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	runner := &AppRunner{
		DialFn: func(network, address string) (net.Conn, error) {
			return c1, nil
		},
	}

	go func() {
		req, _ := protocol.DecodeJSON(c2)
		if req != nil {
			resp := protocol.NewHandshakeResponsePacket(false, "version mismatch", "", 0, 0, 0, nil)
			_ = protocol.EncodeJSON(c2, resp)
		}
	}()

	err := runner.RunClientMode("127.0.0.1:8080", "json")
	if err == nil {
		t.Fatal("expected error due to failed handshake, got nil")
	}
	if !strings.Contains(err.Error(), "handshake failed") {
		t.Errorf("expected handshake failed error message, got: %v", err)
	}
}

func TestRunClientMode_MissingMap(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	runner := &AppRunner{
		DialFn: func(network, address string) (net.Conn, error) {
			return c1, nil
		},
	}

	go func() {
		req, _ := protocol.DecodeJSON(c2)
		if req != nil {
			resp := protocol.NewHandshakeResponsePacket(true, "welcome", "client-1", tilemap.MapWidth, tilemap.MapHeight, tilemap.TileSize, nil)
			_ = protocol.EncodeJSON(c2, resp)
		}
	}()

	err := runner.RunClientMode("127.0.0.1:8080", "json")
	if err == nil {
		t.Fatal("expected error due to missing map, got nil")
	}
	if !strings.Contains(err.Error(), "server did not provide initial map") {
		t.Errorf("expected missing map error message, got: %v", err)
	}
}

func TestRunClientMode_InvalidHandshake(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	runner := &AppRunner{
		DialFn: func(network, address string) (net.Conn, error) {
			return c1, nil
		},
	}

	go func() {
		req, _ := protocol.DecodeJSON(c2)
		if req != nil {
			resp := protocol.NewViewportSubscriptionPacket(0, 0, 10, 10)
			_ = protocol.EncodeJSON(c2, resp)
		}
	}()

	err := runner.RunClientMode("127.0.0.1:8080", "json")
	if err == nil {
		t.Fatal("expected error due to invalid handshake packet type, got nil")
	}
	if !strings.Contains(err.Error(), "invalid handshake response received") {
		t.Errorf("expected invalid handshake error message, got: %v", err)
	}
}

func TestRunClientMode_GameError(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	runner := &AppRunner{
		DialFn: func(network, address string) (net.Conn, error) {
			return c1, nil
		},
		RunGameFn: func(game ebiten.Game) error {
			return fmt.Errorf("graphics device failure")
		},
	}

	go func() {
		req, _ := protocol.DecodeJSON(c2)
		if req != nil {
			m := tilemap.NewDefaultMap()
			resp := protocol.NewHandshakeResponsePacket(true, "welcome", "client-1", tilemap.MapWidth, tilemap.MapHeight, tilemap.TileSize, m)
			_ = protocol.EncodeJSON(c2, resp)
		}
	}()

	err := runner.RunClientMode("127.0.0.1:8080", "json")
	if err == nil {
		t.Fatal("expected error due to game execution failure, got nil")
	}
	if !strings.Contains(err.Error(), "game error") {
		t.Errorf("expected game error message, got: %v", err)
	}
}

func TestRunLocalMode_Success(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	mockServer := &mockServerRunner{}
	runner := &AppRunner{
		NewServerFn: func(addr string, proto string) ServerRunner {
			return mockServer
		},
		DialFn: func(network, address string) (net.Conn, error) {
			return c1, nil
		},
		RunGameFn: func(game ebiten.Game) error {
			return nil
		},
	}

	go func() {
		req, err := protocol.DecodeJSON(c2)
		if err != nil {
			return
		}
		if req.Type != protocol.PacketTypeHandshakeRequest {
			return
		}

		m := tilemap.NewDefaultMap()
		resp := protocol.NewHandshakeResponsePacket(true, "welcome", "client-1", tilemap.MapWidth, tilemap.MapHeight, tilemap.TileSize, m)
		_ = protocol.EncodeJSON(c2, resp)
	}()

	err := runner.RunLocalMode("127.0.0.1:8080", "json")
	if err != nil {
		t.Fatalf("expected no error from RunLocalMode, got: %v", err)
	}

	if !mockServer.started {
		t.Error("expected server to be started")
	}
	if !mockServer.stopped {
		t.Error("expected server to be stopped")
	}
}

func TestRunLocalMode_ServerStartError(t *testing.T) {
	mockServer := &mockServerRunner{startErr: fmt.Errorf("port in use")}
	runner := &AppRunner{
		NewServerFn: func(addr string, proto string) ServerRunner {
			return mockServer
		},
	}

	err := runner.RunLocalMode("127.0.0.1:8080", "json")
	if err == nil {
		t.Fatal("expected error from failed background server start, got nil")
	}
	if !strings.Contains(err.Error(), "failed to start background server") {
		t.Errorf("expected start background server failure message, got: %v", err)
	}
}

func TestRunLocalMode_DialError(t *testing.T) {
	mockServer := &mockServerRunner{}
	dialAttempts := 0
	runner := &AppRunner{
		NewServerFn: func(addr string, proto string) ServerRunner {
			return mockServer
		},
		DialFn: func(network, address string) (net.Conn, error) {
			dialAttempts++
			return nil, fmt.Errorf("connection refused")
		},
	}

	// Override log output to avoid noise in test output
	oldLogWriter := os.Stderr
	defer func() { os.Stderr = oldLogWriter }()
	os.Stderr, _ = os.Open(os.DevNull)

	err := runner.RunLocalMode("127.0.0.1:8080", "json")
	if err == nil {
		t.Fatal("expected dial failure in local mode, got nil")
	}
	if dialAttempts != 10 {
		t.Errorf("expected exactly 10 dial attempts (retry loop), got: %d", dialAttempts)
	}
	if !mockServer.started {
		t.Error("expected server to be started")
	}
	if !mockServer.stopped {
		t.Error("expected server to be stopped during cleanup")
	}
}
