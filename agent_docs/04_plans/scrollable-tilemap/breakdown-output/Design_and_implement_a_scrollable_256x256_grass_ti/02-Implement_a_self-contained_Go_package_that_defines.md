# Implement a self-contained Go package that defines client/server packet structures (including handshakes and map streaming data) and provides encoding/decoding utilities using both gob and JSON formats.

This task involves designing and implementing a robust Go network protocol package (e.g., 'pkg/protocol' or 'internal/protocol') that acts as the shared communication contract between the authoritative server and the Ebitengine client.

It defines the packet structures needed for the system, specifically focusing on the client-server handshake, viewport subscription coordinates, and streamed grass tilemap chunks. To facilitate efficient binary transmission while maintaining developer observability, the package will provide serialization and deserialization APIs supporting Go's native 'gob' encoding and standard 'JSON' as a fallback or debug transport.
