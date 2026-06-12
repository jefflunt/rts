# Scrollable RTS Tilemap Engine (with Interactive Minimap)

Welcome to the **Scrollable RTS Tilemap Engine**! This is a high-performance, real-time strategy (RTS) style client-server prototype built from the ground up in Go using the [Ebitengine (Ebiten v2)](https://ebiten.org) 2D game library. 

This repository demonstrates how to architect a modern, scalable, multiplayer-ready game with decoupled authoritative server-side state, efficient client-side camera viewport culling, and interactive tactical navigation.

---

## 🚀 Key Features

- **🌐 Authoritative TCP Server-Client Architecture**: 
  - The server holds and serves the master game map and state.
  - Supports both **JSON** and high-speed **Gob** serialization protocols over raw TCP sockets.
  - Multiple connection modes: standalone server, remote client, or an integrated local-loopback single-player mode.
- **🗺️ Giant Scrollable World**: 
  - A massive 256x256 tilemap (8,192 x 8,192 pixels at 32x32px per tile).
  - Smooth camera panning using **Keyboard Arrow Keys** or **Mouse-Edge Scrolling** (classic RTS panning style).
- **⚡ Viewport-Culling Renderer**:
  - The engine performs advanced camera bounding-box culling. Only the tiles visible within the 1024x768 viewport are rendered, maintaining low CPU/GPU usage and high framerates.
- **🗺️ Interactive Minimap Overlay**:
  - Placed in the bottom-left corner with a metallic border.
  - Uses an $O(1)$ pre-rendered cache texture for maximum performance.
  - Automatically handles different map sizes (64x64, 128x128, 256x256) by scaling tile representation.
  - Features a **live white rectangle indicating the camera viewport boundary**.
  - Supports **Click-to-Pan** and **Drag-to-Pan**: clicking or dragging anywhere on the minimap instantly centers the main camera view on that world position.
- **🧪 Production-Grade Software Engineering**:
  - Developed with strict **Test-Driven Development (TDD)**.
  - Exhaustive test coverage for coordinate systems, network handshakes, camera physics, minimap calculations, and packet encoding.

---

## 🎮 How to Play / Use

### Requirements
- [Go (1.20+)](https://go.dev/dl/) installed.
- A desktop environment supporting graphical applications (macOS, Windows, or Linux with OpenGL dependencies).

### Run in Single-Player (Local Loopback) Mode
This is the easiest way to run the game! It automatically spins up the authoritative server in a background thread and launches the graphical client seamlessly in the foreground.

```bash
go run cmd/game/main.go --mode local
```

### Run in Dedicated Client/Server Mode
If you want to run them in separate terminal windows (or even on separate machines on a local network):

1. **Start the Authoritative Server**:
   ```bash
   go run cmd/game/main.go --mode server --addr "0.0.0.0:8080"
   ```

2. **Connect the Client**:
   ```bash
   go run cmd/game/main.go --mode client --addr "127.0.0.1:8080"
   ```

### Command-Line Arguments
Customize how the game runs with these flags:
- `--mode`: Options are `local` (default), `server`, or `client`.
- `--addr` (or `--address`): The TCP address to bind/connect to (default `127.0.0.1:8080`).
- `--proto` (or `--protocol`): Serialization format to use. Options are `json` (default) or `gob`.

---

## ⌨️ Controls

- **Arrow Keys (Up, Down, Left, Right)**: Pan the camera view smoothly across the grass map.
- **Mouse-Edge Scrolling**: Move your mouse cursor within 15 pixels of any edge of the screen to pan the camera in that direction.
- **Minimap Left-Click / Drag**: Press and hold the left mouse button anywhere inside the minimap overlay (bottom-left corner) to instantly snap or drag the camera viewport to that position in the world.

---

## 🏗️ Architecture & Codebase Layout

```
├── cmd/
│   └── game/
│       ├── main.go               # Unified entry point parsing CLI flags
│       └── main_test.go          # CLI integration and mode testing
├── internal/
│   ├── client/
│   │   ├── game.go               # Main Ebitengine game loop (Update, Draw, Layout)
│   │   ├── camera.go             # Camera math, conversions, bounds, and viewport culling
│   │   ├── input.go              # Abstracted keyboard & mouse panning controllers
│   │   └── minimap.go            # Minimap coordinate translations, rendering cache, and dragging
│   ├── map/
│   │   ├── tile.go               # Tile definitions (grass types, colors)
│   │   └── map.go                # Tile map grid data structures
│   ├── protocol/
│   │   └── messages.go           # Handshake protocol, packet parsing, and JSON/Gob encoders
│   └── server/
│       └── server.go             # Authoritative TCP server and network connection handler
└── agent_docs/                   # Comprehensive AI and human alignment documentation
```

---

## 🛠️ Development & Testing

We use standard Go tools for build and verification. Since we adhere to high-coverage standards and TDD, all modules have robust unit and integration tests.

### Run All Tests
```bash
go test ./... -v
```

---

## 📖 AI Agent Documentation (`agent_docs`)
This project leverages the [**`agent_docs` framework**](https://github.com/jefflunt/agent_docs), which provides a structured approach for AI-human continuous alignment and progressive context disclosure.

Inside the `/agent_docs` folder, you will find:
- `01_orientation/`: Structural overview, domain logic, and onboarding guides.
- `02_patterns/`: Coding conventions, testing frameworks, and architectural standards.
- `03_deep_dives/`: Deep analyses of subsystem boundaries, caching strategies, and protocols.
- `04_plans/`: Active feature planning backlogs and completed item tracking.
- `templates/`: Boilerplate markdown files for expanding the documentation.

See [agent_docs/README.md](agent_docs/README.md) for more details.

---

## 📄 License
This project is open-source and available under the [MIT License](LICENSE).
