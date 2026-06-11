package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"scrollable-tilemap/internal/client"
	"scrollable-tilemap/internal/protocol"
	"scrollable-tilemap/internal/server"
)

// ServerRunner abstracts starting/stopping the TCP server.
type ServerRunner interface {
	Start() error
	Stop() error
}

// DefaultServerRunner is a wrapper for server.Server to implement ServerRunner.
type DefaultServerRunner struct {
	*server.Server
}

// AppRunner handles the execution of client, server, and local modes.
// Decoupled for isolated unit testing and mocking.
type AppRunner struct {
	DialFn      func(network, address string) (net.Conn, error)
	RunGameFn   func(game ebiten.Game) error
	NewServerFn func(addr string, proto string) ServerRunner
}

// NewDefaultAppRunner creates an AppRunner with production implementations.
func NewDefaultAppRunner() *AppRunner {
	return &AppRunner{
		DialFn: net.Dial,
		RunGameFn: func(g ebiten.Game) error {
			ebiten.SetWindowSize(1024, 768)
			ebiten.SetWindowTitle("Scrollable Tilemap")
			return ebiten.RunGame(g)
		},
		NewServerFn: func(addr string, proto string) ServerRunner {
			return &DefaultServerRunner{
				Server: server.NewServer(addr, server.ProtocolType(proto)),
			}
		},
	}
}

// Run executes the application according to the selected mode.
func (r *AppRunner) Run(mode, addr, proto string) error {
	switch mode {
	case "server":
		return r.RunServerMode(addr, proto)
	case "client":
		return r.RunClientMode(addr, proto)
	case "local":
		return r.RunLocalMode(addr, proto)
	default:
		return fmt.Errorf("invalid mode: %s", mode)
	}
}

// RunServerMode runs the server component in the foreground, blocking until an interrupt or termination signal is received.
func (r *AppRunner) RunServerMode(addr, proto string) error {
	s := r.NewServerFn(addr, proto)
	log.Printf("Starting TCP server on %s (protocol: %s)...", addr, proto)
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for termination signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Stopping TCP server...")
	if err := s.Stop(); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}
	return nil
}

// RunClientMode runs the client component, connecting to the TCP server, performing the handshake, and launching Ebitengine.
func (r *AppRunner) RunClientMode(addr, proto string) error {
	log.Printf("Connecting to TCP server on %s (protocol: %s)...", addr, proto)
	conn, err := r.DialFn("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	return r.executeClientHandshakeAndLoop(conn, proto)
}

// RunLocalMode runs both the server (in the background) and client (in the main thread) in local loopback mode.
func (r *AppRunner) RunLocalMode(addr, proto string) error {
	s := r.NewServerFn(addr, proto)
	log.Printf("Starting background TCP server on %s (protocol: %s)...", addr, proto)
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start background server: %w", err)
	}
	defer func() {
		log.Println("Stopping background TCP server...")
		_ = s.Stop()
	}()

	// Retry loop for local dialing
	var conn net.Conn
	var err error
	for i := 0; i < 10; i++ {
		conn, err = r.DialFn("tcp", addr)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to local background server: %w", err)
	}
	defer conn.Close()

	return r.executeClientHandshakeAndLoop(conn, proto)
}

func (r *AppRunner) executeClientHandshakeAndLoop(conn net.Conn, proto string) error {
	log.Println("Sending handshake request...")
	req := protocol.NewHandshakeRequestPacket("1.0.0", "client-1")
	var err error
	if proto == "gob" {
		err = protocol.EncodeGob(conn, req)
	} else {
		err = protocol.EncodeJSON(conn, req)
	}
	if err != nil {
		return fmt.Errorf("failed to encode handshake request: %w", err)
	}

	log.Println("Awaiting handshake response...")
	var resp *protocol.Packet
	if proto == "gob" {
		resp, err = protocol.DecodeGob(conn)
	} else {
		resp, err = protocol.DecodeJSON(conn)
	}
	if err != nil {
		return fmt.Errorf("failed to decode handshake response: %w", err)
	}

	if resp == nil || resp.Type != protocol.PacketTypeHandshakeResponse || resp.HandshakeResponse == nil {
		return fmt.Errorf("invalid handshake response received")
	}

	hr := resp.HandshakeResponse
	if !hr.Success {
		return fmt.Errorf("handshake failed: %s", hr.Message)
	}

	log.Printf("Handshake successful! Client ID: %s, Map size: %dx%d, Tile size: %d", hr.ClientID, hr.MapWidth, hr.MapHeight, hr.TileSize)

	initialMap := hr.InitialMap
	if initialMap == nil {
		return fmt.Errorf("server did not provide initial map")
	}

	cam := client.NewCamera(1024, 768, 300.0, 15)
	input := client.NewEbitenInputProvider()
	g := client.NewGame(cam, input, initialMap)
	g.SetMapDimensions(hr.MapWidth, hr.MapHeight, hr.TileSize)

	log.Println("Launching Ebitengine game client...")
	if err := r.RunGameFn(g); err != nil {
		return fmt.Errorf("game error: %w", err)
	}

	return nil
}

func main() {
	var mode, addr, proto string
	flag.StringVar(&mode, "mode", "local", "Execution mode: 'server', 'client', or 'local'")
	flag.StringVar(&addr, "addr", "127.0.0.1:8080", "TCP server address")
	flag.StringVar(&addr, "address", "127.0.0.1:8080", "TCP server address (alias)")
	flag.StringVar(&proto, "proto", "json", "Network protocol: 'json' or 'gob'")
	flag.StringVar(&proto, "protocol", "json", "Network protocol: 'json' or 'gob' (alias)")

	flag.Parse()

	runner := NewDefaultAppRunner()
	if err := runner.Run(mode, addr, proto); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}
