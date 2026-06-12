# Deep Dive: Network Handshake & Protocols

## Overview
The protocol subsystem handles synchronous, low-latency client-server handshaking over standard TCP sockets. It supports dual serialization mechanisms: JSON (highly readable/debuggable) and Gob (native Go binary stream, highly optimized).

---

## Purpose
Establishing a secure and aligned handshake sequence is critical before initializing the client-side game loop. The client must verify its version, exchange credentials, and receive the authoritative game map structure synchronously before drawing any frames.

---

## Components

- **`Packet` Struct (`internal/protocol/messages.go`)**:
  - The universal envelope for all message types. Contains metadata like `Type` and payload fields (`HandshakeRequest`, `HandshakeResponse`).
- **Encoders and Decoders (`EncodeJSON`, `DecodeJSON`, `EncodeGob`, `DecodeGob`)**:
  - Wrappers around standard Go `json` and `encoding/gob` libraries.
  - Encode and decode `Packet` envelopes to/from standard `net.Conn` interfaces.
- **Authoritative Server Listener (`internal/server/server.go`)**:
  - Starts a TCP socket listener on a specified address.
  - Spawns concurrent connection handler goroutines (`handleConnection`) which synchronously read handshakes, perform validation, generate map grids, and stream initial maps.
- **AppRunner Network Setup (`cmd/game/main.go`)**:
  - Leverages mockable network dial/serve functions to support local in-memory loopback and dedicated TCP sockets interchangeably.

---

## Data Flow

```
1. Client dials TCP ───> Establishes raw socket connection to Server
2. Client sends:       [HandshakeRequestPacket] (Version, ClientID)
3. Server receives ───> Decodes and validates version. If mismatch: sends HandshakeResponse(Success=false)
4. Server generates ──> Map grid (256x256 grassland tiles)
5. Server sends:       [HandshakeResponsePacket] (Success=true, MapDimensions, InitialMap)
6. Client boots ──────> Receives response, instantiates game client, starts Ebitengine UI loop
```

---

## Concerns & Constraints

- **Blocking Network I/O**: Network handshakes are synchronous and blocking. The client cannot and should not start the game loop until the handshake succeeds.
- **Serialization Formats**: 
  - **Gob** is Go-specific and fast. When adding non-Go clients in the future, **JSON** ensures cross-platform interoperability.
  - Using dual codecs helps developers balance performance vs. debuggability during development.
