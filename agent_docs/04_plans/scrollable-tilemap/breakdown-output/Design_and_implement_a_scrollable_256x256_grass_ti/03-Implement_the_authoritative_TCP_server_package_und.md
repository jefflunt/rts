# Implement the authoritative TCP server package under internal/server that generates and hosts a 256x256 grass tilemap, listens for incoming client connections, and streams the serialized map configuration and initial tile data to connecting clients using the established network protocols.

This task focuses on implementing the authoritative TCP server package within the Go application codebase. The server will generate and maintain the authoritative 256x256 grass tilemap, ensuring a single source of truth for all clients. It will listen on a designated TCP port and await incoming client connections.

Upon connection, the server manages the network handshake. It will serialize the game's map metadata and tile data using the protocol defined under `internal/protocol/` and stream this data over the TCP socket to the client. Each client handler will run concurrently in its own goroutine to allow non-blocking map streaming and scalable connection handling.
