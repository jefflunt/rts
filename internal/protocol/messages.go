package protocol

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"

	tilemap "scrollable-tilemap/internal/map"
)

// PacketType represents the type of a packet.
type PacketType string

const (
	PacketTypeHandshakeRequest    PacketType = "handshake_request"
	PacketTypeHandshakeResponse   PacketType = "handshake_response"
	PacketTypeViewportSubscription PacketType = "viewport_subscription"
	PacketTypeTilemapChunk        PacketType = "tilemap_chunk"
)

// Packet is a unified envelope that can carry any message payload.
// This simplifies the networking code tremendously as the same top-level type
// can be serialized and deserialized, then route based on Type.
type Packet struct {
	Type              PacketType            `json:"type"`
	HandshakeRequest  *HandshakeRequest     `json:"handshake_request,omitempty"`
	HandshakeResponse *HandshakeResponse    `json:"handshake_response,omitempty"`
	Subscription      *ViewportSubscription `json:"subscription,omitempty"`
	Chunk             *TilemapChunk         `json:"chunk,omitempty"`
}

// HandshakeRequest is sent by the client when it first connects.
type HandshakeRequest struct {
	ClientVersion string `json:"client_version"`
	ClientID      string `json:"client_id"`
}

// HandshakeResponse is returned by the server to accept/deny a client connection,
// providing the map dimensions and initial configuration.
type HandshakeResponse struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	ClientID   string       `json:"client_id"`
	MapWidth   int          `json:"map_width"`
	MapHeight  int          `json:"map_height"`
	TileSize   int          `json:"tile_size"`
	InitialMap *tilemap.Map `json:"initial_map,omitempty"` // optional initial map
}

// ViewportSubscription is sent by the client to register its interest
// in a specific coordinate bounding box. The server uses this to stream relevant tiles.
type ViewportSubscription struct {
	MinTileX int `json:"min_tile_x"`
	MinTileY int `json:"min_tile_y"`
	MaxTileX int `json:"max_tile_x"`
	MaxTileY int `json:"max_tile_y"`
}

// TilemapChunk represents a streamed region of the tilemap.
type TilemapChunk struct {
	StartX int            `json:"start_x"`
	StartY int            `json:"start_y"`
	Width  int            `json:"width"`
	Height int            `json:"height"`
	Tiles  []tilemap.Tile `json:"tiles"` // 1D flat row-major slice of tiles
}

// Register types with Gob for interface safety and general usage
func init() {
	gob.Register(&Packet{})
	gob.Register(&HandshakeRequest{})
	gob.Register(&HandshakeResponse{})
	gob.Register(&ViewportSubscription{})
	gob.Register(&TilemapChunk{})
	gob.Register(&tilemap.Map{})
	gob.Register(tilemap.Tile{})
}

// GetTile returns the tile at the chunk-relative coordinates (x, y).
func (c *TilemapChunk) GetTile(x, y int) (tilemap.Tile, error) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return tilemap.Tile{}, errors.New("chunk-relative coordinates out of bounds")
	}
	idx := y*c.Width + x
	if idx >= len(c.Tiles) {
		return tilemap.Tile{}, errors.New("tile index out of bounds in chunk data")
	}
	return c.Tiles[idx], nil
}

// SetTile sets the tile at the chunk-relative coordinates (x, y).
func (c *TilemapChunk) SetTile(x, y int, tile tilemap.Tile) error {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return errors.New("chunk-relative coordinates out of bounds")
	}
	idx := y*c.Width + x
	if idx >= len(c.Tiles) {
		return errors.New("tile index out of bounds in chunk data")
	}
	c.Tiles[idx] = tile
	return nil
}

// NewTilemapChunk creates a new TilemapChunk from a slice or bounds of the map.
func NewTilemapChunk(startX, startY, width, height int, m *tilemap.Map) (*TilemapChunk, error) {
	if m == nil {
		return nil, errors.New("cannot create chunk from nil map")
	}
	tiles := make([]tilemap.Tile, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			tx := startX + x
			ty := startY + y
			tile, err := m.GetTile(tx, ty)
			if err != nil {
				return nil, err
			}
			tiles[y*width+x] = tile
		}
	}
	return &TilemapChunk{
		StartX: startX,
		StartY: startY,
		Width:  width,
		Height: height,
		Tiles:  tiles,
	}, nil
}

// NewHandshakeRequestPacket wraps a HandshakeRequest in a Packet.
func NewHandshakeRequestPacket(version, id string) *Packet {
	return &Packet{
		Type: PacketTypeHandshakeRequest,
		HandshakeRequest: &HandshakeRequest{
			ClientVersion: version,
			ClientID:      id,
		},
	}
}

// NewHandshakeResponsePacket wraps a HandshakeResponse in a Packet.
func NewHandshakeResponsePacket(success bool, msg, clientID string, mapW, mapH, tileSize int, m *tilemap.Map) *Packet {
	return &Packet{
		Type: PacketTypeHandshakeResponse,
		HandshakeResponse: &HandshakeResponse{
			Success:    success,
			Message:    msg,
			ClientID:   clientID,
			MapWidth:   mapW,
			MapHeight:  mapH,
			TileSize:   tileSize,
			InitialMap: m,
		},
	}
}

// NewViewportSubscriptionPacket wraps a ViewportSubscription in a Packet.
func NewViewportSubscriptionPacket(minX, minY, maxX, maxY int) *Packet {
	return &Packet{
		Type: PacketTypeViewportSubscription,
		Subscription: &ViewportSubscription{
			MinTileX: minX,
			MinTileY: minY,
			MaxTileX: maxX,
			MaxTileY: maxY,
		},
	}
}

// NewTilemapChunkPacket wraps a TilemapChunk in a Packet.
func NewTilemapChunkPacket(chunk *TilemapChunk) *Packet {
	return &Packet{
		Type:  PacketTypeTilemapChunk,
		Chunk: chunk,
	}
}

// EncodeJSON writes the packet as JSON to the given writer.
func EncodeJSON(w io.Writer, p *Packet) error {
	if p == nil {
		return errors.New("cannot encode nil packet")
	}
	return json.NewEncoder(w).Encode(p)
}

// DecodeJSON reads a packet as JSON from the given reader.
func DecodeJSON(r io.Reader) (*Packet, error) {
	var p Packet
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// MarshalJSONPacket converts the packet into a JSON byte slice.
func MarshalJSONPacket(p *Packet) ([]byte, error) {
	if p == nil {
		return nil, errors.New("cannot marshal nil packet")
	}
	return json.Marshal(p)
}

// UnmarshalJSONPacket parses a JSON byte slice into a Packet.
func UnmarshalJSONPacket(data []byte) (*Packet, error) {
	var p Packet
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// EncodeGob writes the packet using Gob encoding to the given writer.
func EncodeGob(w io.Writer, p *Packet) error {
	if p == nil {
		return errors.New("cannot encode nil packet")
	}
	return gob.NewEncoder(w).Encode(p)
}

// DecodeGob reads a packet using Gob decoding from the given reader.
func DecodeGob(r io.Reader) (*Packet, error) {
	var p Packet
	if err := gob.NewDecoder(r).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// MarshalGobPacket converts the packet into a Gob-encoded byte slice.
func MarshalGobPacket(p *Packet) ([]byte, error) {
	if p == nil {
		return nil, errors.New("cannot marshal nil packet")
	}
	var buf bytes.Buffer
	if err := EncodeGob(&buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalGobPacket parses a Gob-encoded byte slice into a Packet.
func UnmarshalGobPacket(data []byte) (*Packet, error) {
	buf := bytes.NewReader(data)
	return DecodeGob(buf)
}

// EncodeGenericJSON writes any value as JSON to the given writer.
func EncodeGenericJSON[T any](w io.Writer, val *T) error {
	if val == nil {
		return errors.New("cannot encode nil value")
	}
	return json.NewEncoder(w).Encode(val)
}

// DecodeGenericJSON reads any value as JSON from the given reader.
func DecodeGenericJSON[T any](r io.Reader) (*T, error) {
	var val T
	if err := json.NewDecoder(r).Decode(&val); err != nil {
		return nil, err
	}
	return &val, nil
}

// EncodeGenericGob writes any value using Gob encoding to the given writer.
func EncodeGenericGob[T any](w io.Writer, val *T) error {
	if val == nil {
		return errors.New("cannot encode nil value")
	}
	return gob.NewEncoder(w).Encode(val)
}

// DecodeGenericGob reads any value using Gob decoding from the given reader.
func DecodeGenericGob[T any](r io.Reader) (*T, error) {
	var val T
	if err := gob.NewDecoder(r).Decode(&val); err != nil {
		return nil, err
	}
	return &val, nil
}
