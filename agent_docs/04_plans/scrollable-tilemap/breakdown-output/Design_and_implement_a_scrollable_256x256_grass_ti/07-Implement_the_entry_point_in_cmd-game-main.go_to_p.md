# Implement the entry point in cmd/game/main.go to parse mode flags ('server', 'client', 'local') and launch the appropriate architectural component (TCP server, Ebitengine client, or both in local loopback mode).

This task is responsible for implementing the single CLI entry point for the application at 'cmd/game/main.go'. By using Go's 'flag' package, the application will support a '--mode' flag with three options: 'server', 'client', and 'local'. This design integrates client, server, and local runners in a cohesive binary, allowing flexible deployment.

Running in 'server' mode starts the TCP game server to stream the tilemap. Running in 'client' mode starts the Ebitengine-based game client. Running in 'local' mode launches the TCP server in a background goroutine and starts the client in the main thread for a seamless offline experience.
