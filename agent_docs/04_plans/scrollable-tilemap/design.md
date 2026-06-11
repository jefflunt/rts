# Design Document: Scrollable 256x256 Tilemap with Go Client/Server Architecture

## User Story

### Headline
As a player, I want to open a scrollable game window displaying a 256x256 grass tilemap that connects to a local background server, so that I can pan around the map using arrow keys and mouse-edge scrolling.

### Problem Statement
An RTS game like StarCraft requires a large, coordinate-based map (256x256 tiles, or 8192x8192 pixels at 32x32px per tile). Rendering all 65,536 tiles on every frame is highly inefficient and causes lag. Furthermore, we want an architecture that supports multiplayer from the start by decoupling the authoritative server-side game state from the rendering client. The client must connect to a server, receive the map data, maintain its own local camera, perform viewport bounding-box culling, and support classic RTS edge panning and keyboard movement.

### Objective
Build a Go-based client/server system inside a unified binary.
- **Server**: Generates and maintains a 256x256 tilemap (all grass) and hosts a TCP server.
- **Client**: Connects to the server over TCP, fetches the map data, opens a window via **Ebitengine**, captures camera inputs (Arrow Keys + Mouse-Edge Panning), and renders only the visible tiles using camera-to-world viewport culling.
- **Single/Multi-Mode**: Provide a main package that supports `--mode server`, `--mode client`, and `--mode local` (launching both in-memory/via goroutines).

### Expected Outcome
Running the application in local mode opens a desktop window of size 1024x768. The window displays a seamless grass tilemap. Hovering the cursor near the edge of the screen or pressing the arrow keys smoothly scrolls the map. The rendering is fast and performs camera bounding-box culling.

---

## Architecture Overview

We use a single unified Go repository. The application can run in three modes controlled by CLI flags:
1. `server`: Runs the authoritative TCP server.
2. `client`: Runs the Ebitengine graphical client.
3. `local`: Runs the TCP server in a background goroutine, and then launches the client in the main thread (seamless single-player/offline experience).

### Project Layout
```
├── cmd/
│   └── game/
│       └── main.go           # CLI entry point to start client, server, or local mode
├── internal/
│   ├── protocol/
│   │   ├── messages.go       # Structs for network packet encoding/decoding (JSON or gob)
│   ├── map/
│   │   ├── tile.go           # Tile definitions and constants
│   │   └── map.go            # Map data structures and coordinate helpers
│   ├── server/
│   │   └── server.go         # TCP Server listener, connection management, map streaming
│   └── client/
│       ├── camera.go         # Camera math, viewport culling, and coordinate conversion
│       ├── input.go          # Keyboard and Mouse-Edge panning input systems
│       └── game.go           # Ebitengine Game implementation (Update, Draw, Layout)
```

### Protocol Flow
1. **Connection**: Client establishes TCP connection to Server.
2. **Handshake (Server -> Client)**: Server serializes the 256x256 map configuration and initial tile list, sending it over TCP.
3. **Keepalive/Stream (Periodic)**: (Optional for now) Server sends heartbeat or state sync packets. Client periodically can notify server of its state.

### Coordinate Systems
- **Grid Coordinates (Tile)**: `(x, y)` where `0 <= x < 256` and `0 <= y < 256`.
- **World Coordinates (Pixels)**: `(wx, wy)` where `wx = x * TileSize` and `wy = y * TileSize`.
- **Screen/Viewport Coordinates (Pixels)**: `(sx, sy)` relative to the client window (e.g. 1024x768).
  - Formula: `ScreenX = WorldX - CameraX`
  - Formula: `ScreenY = WorldY - CameraY`

---

## Implementation Backlog

### ## Pending
- **Task 1: Initialize Go module & Domain map logic**
  - Initialize Go module workspace.
  - Define core constants (TileSize = 32px, MapSize = 256).
  - Implement `internal/map` structures (Tile types, Map grid, Coordinate translation helper math).
- **Task 2: Define Client/Server Shared Protocol**
  - Create package `internal/protocol`.
  - Design serialization/deserialization structs for handshake and map data transfer using gob and JSON.
- **Task 3: Implement Authoritative TCP Server**
  - Create package `internal/server`.
  - Implement TCP listening server that serves map state.
  - Concurrent client handlers streaming serialized map configuration.
- **Task 4: Implement Client-Side Camera System**
  - Create package `internal/client`.
  - Implement Camera model containing viewport status, coordinate conversion (world/screen), and bounding-box culling calculations.
- **Task 5: Implement Client Panning Inputs**
  - Add input processing module to support keyboard arrow keys and mouse-edge boundary panning.
  - Ensure viewport coordinates smoothly pan and stay clamped within valid 256x256 map boundaries.
- **Task 6: Implement Core Ebitengine Loop & Procedural Tile Drawing**
  - Establish `ebiten.Game` implementation.
  - Render procedural checkered grass tiles to visually represent panning and coordinate transitions.
  - Integrate Camera viewport culling logic in Ebitengine's Draw loop.
- **Task 7: Build Unified CLI Entry Point**
  - Implement parsing of `--mode` flags (`server`, `client`, `local`) in `cmd/game/main.go`.
  - Wire loopback in `local` mode (launching TCP server in background goroutine and Ebitengine client on main thread).

### ## Current
*(None yet. Waiting for initial approval)*

### ## Completed
*(None)*

---

## Checklist & TDD Requirements

1. **Strict Test-Driven Development**:
   - For all logic components (camera math, viewport culling, protocol parsing, map generation), write the tests first.
   - Prove the tests fail (RED) before writing the implementation to make them pass (GREEN).
2. **Culling Verification**:
   - Viewport culling must be rigorously tested. The culling function should accept a camera coordinate, screen size, and map bounds, and return the minimum and maximum tile indices that must be rendered.
3. **Decoupled Architecture**:
   - The game client rendering loop must not directly access or write to server internals. All map data must come through the deserialized network protocol struct.
4. **No External Image Assets (Procedural Textures)**:
   - For the initial tilemap, draw the grass tile procedurally (e.g. a green square with light green specks/lines) using Ebitengine's dynamic images or vector drawing, avoiding complex asset-loading paths for now. This keeps compilation clean and instant.
