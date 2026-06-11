package protocol

import (
	"bytes"
	"errors"
	"testing"

	tilemap "scrollable-tilemap/internal/map"
)

func TestNewHandshakeRequestPacket(t *testing.T) {
	version := "1.0.0"
	id := "client-abc"
	p := NewHandshakeRequestPacket(version, id)

	if p == nil {
		t.Fatal("NewHandshakeRequestPacket returned nil")
	}
	if p.Type != PacketTypeHandshakeRequest {
		t.Errorf("expected type %s, got %s", PacketTypeHandshakeRequest, p.Type)
	}
	if p.HandshakeRequest == nil {
		t.Fatal("HandshakeRequest payload was nil")
	}
	if p.HandshakeRequest.ClientVersion != version {
		t.Errorf("expected client version %s, got %s", version, p.HandshakeRequest.ClientVersion)
	}
	if p.HandshakeRequest.ClientID != id {
		t.Errorf("expected client ID %s, got %s", id, p.HandshakeRequest.ClientID)
	}
}

func TestNewHandshakeResponsePacket(t *testing.T) {
	success := true
	msg := "welcome"
	clientID := "client-xyz"
	mapW := 256
	mapH := 256
	tileSize := 32
	m := tilemap.NewDefaultMap()

	p := NewHandshakeResponsePacket(success, msg, clientID, mapW, mapH, tileSize, m)

	if p == nil {
		t.Fatal("NewHandshakeResponsePacket returned nil")
	}
	if p.Type != PacketTypeHandshakeResponse {
		t.Errorf("expected type %s, got %s", PacketTypeHandshakeResponse, p.Type)
	}
	if p.HandshakeResponse == nil {
		t.Fatal("HandshakeResponse payload was nil")
	}
	if p.HandshakeResponse.Success != success {
		t.Errorf("expected success %v, got %v", success, p.HandshakeResponse.Success)
	}
	if p.HandshakeResponse.Message != msg {
		t.Errorf("expected message %s, got %s", msg, p.HandshakeResponse.Message)
	}
	if p.HandshakeResponse.ClientID != clientID {
		t.Errorf("expected client ID %s, got %s", clientID, p.HandshakeResponse.ClientID)
	}
	if p.HandshakeResponse.MapWidth != mapW {
		t.Errorf("expected map width %d, got %d", mapW, p.HandshakeResponse.MapWidth)
	}
	if p.HandshakeResponse.MapHeight != mapH {
		t.Errorf("expected map height %d, got %d", mapH, p.HandshakeResponse.MapHeight)
	}
	if p.HandshakeResponse.TileSize != tileSize {
		t.Errorf("expected tile size %d, got %d", tileSize, p.HandshakeResponse.TileSize)
	}
	if p.HandshakeResponse.InitialMap != m {
		t.Errorf("expected initial map %v, got %v", m, p.HandshakeResponse.InitialMap)
	}
}

func TestNewViewportSubscriptionPacket(t *testing.T) {
	minX, minY, maxX, maxY := 10, 20, 30, 40
	p := NewViewportSubscriptionPacket(minX, minY, maxX, maxY)

	if p == nil {
		t.Fatal("NewViewportSubscriptionPacket returned nil")
	}
	if p.Type != PacketTypeViewportSubscription {
		t.Errorf("expected type %s, got %s", PacketTypeViewportSubscription, p.Type)
	}
	if p.Subscription == nil {
		t.Fatal("Subscription payload was nil")
	}
	if p.Subscription.MinTileX != minX || p.Subscription.MinTileY != minY ||
		p.Subscription.MaxTileX != maxX || p.Subscription.MaxTileY != maxY {
		t.Errorf("unexpected subscription values: got (%d, %d, %d, %d)",
			p.Subscription.MinTileX, p.Subscription.MinTileY,
			p.Subscription.MaxTileX, p.Subscription.MaxTileY)
	}
}

func TestNewTilemapChunkPacket(t *testing.T) {
	chunk := &TilemapChunk{
		StartX: 5,
		StartY: 5,
		Width:  2,
		Height: 2,
		Tiles:  []tilemap.Tile{{Type: 1}, {Type: 2}, {Type: 3}, {Type: 4}},
	}
	p := NewTilemapChunkPacket(chunk)

	if p == nil {
		t.Fatal("NewTilemapChunkPacket returned nil")
	}
	if p.Type != PacketTypeTilemapChunk {
		t.Errorf("expected type %s, got %s", PacketTypeTilemapChunk, p.Type)
	}
	if p.Chunk != chunk {
		t.Errorf("expected chunk pointer %v, got %v", chunk, p.Chunk)
	}
}

func TestNewTilemapChunk(t *testing.T) {
	m := tilemap.NewDefaultMap()

	// Test Nil Map Error
	_, err := NewTilemapChunk(0, 0, 10, 10, nil)
	if err == nil {
		t.Error("expected error when creating chunk with nil map, got nil")
	}

	// Test Valid Chunk Creation
	startX, startY, width, height := 10, 15, 4, 3
	chunk, err := NewTilemapChunk(startX, startY, width, height, m)
	if err != nil {
		t.Fatalf("failed to create valid chunk: %v", err)
	}
	if chunk.StartX != startX || chunk.StartY != startY || chunk.Width != width || chunk.Height != height {
		t.Errorf("unexpected chunk dimensions: got (%d, %d, %d, %d), want (%d, %d, %d, %d)",
			chunk.StartX, chunk.StartY, chunk.Width, chunk.Height, startX, startY, width, height)
	}
	if len(chunk.Tiles) != width*height {
		t.Errorf("expected %d tiles, got %d", width*height, len(chunk.Tiles))
	}

	// Verify tile contents match the map
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			expectedTile, err := m.GetTile(startX+x, startY+y)
			if err != nil {
				t.Fatalf("failed to get tile from map for comparison: %v", err)
			}
			gotTile, err := chunk.GetTile(x, y)
			if err != nil {
				t.Fatalf("failed to get tile from chunk: %v", err)
			}
			if gotTile != expectedTile {
				t.Errorf("tile mismatch at chunk relative (%d, %d)", x, y)
			}
		}
	}

	// Test Out of Bounds Map Error
	_, err = NewTilemapChunk(250, 250, 10, 10, m)
	if err == nil {
		t.Error("expected out of bounds error when chunk goes beyond map dimensions, got nil")
	}
}

func TestTilemapChunkGetAndSetTile(t *testing.T) {
	chunk := &TilemapChunk{
		StartX: 0,
		StartY: 0,
		Width:  3,
		Height: 3,
		Tiles:  make([]tilemap.Tile, 9),
	}

	newTile := tilemap.Tile{Type: tilemap.TileTypeGrassVariant1, Variant: 99}

	// Set valid tile
	err := chunk.SetTile(1, 2, newTile)
	if err != nil {
		t.Fatalf("failed to set valid tile: %v", err)
	}

	// Get valid tile
	gotTile, err := chunk.GetTile(1, 2)
	if err != nil {
		t.Fatalf("failed to get valid tile: %v", err)
	}
	if gotTile != newTile {
		t.Errorf("expected tile %+v, got %+v", newTile, gotTile)
	}

	// Test bounds checking for GetTile
	boundsTests := []struct {
		x, y int
	}{
		{-1, 0},
		{0, -1},
		{3, 0},
		{0, 3},
		{4, 4},
	}

	for _, tc := range boundsTests {
		_, err := chunk.GetTile(tc.x, tc.y)
		if err == nil {
			t.Errorf("GetTile(%d, %d) expected error, got nil", tc.x, tc.y)
		}
		err = chunk.SetTile(tc.x, tc.y, newTile)
		if err == nil {
			t.Errorf("SetTile(%d, %d) expected error, got nil", tc.x, tc.y)
		}
	}

	// Test short internal Tiles slice bounds checking
	malformedChunk := &TilemapChunk{
		StartX: 0,
		StartY: 0,
		Width:  3,
		Height: 3,
		Tiles:  make([]tilemap.Tile, 2), // too short for 3x3
	}

	_, err = malformedChunk.GetTile(1, 1) // index would be 1*3 + 1 = 4
	if err == nil {
		t.Error("expected index out of bounds error from short tile slice on GetTile, got nil")
	}

	err = malformedChunk.SetTile(1, 1, newTile)
	if err == nil {
		t.Error("expected index out of bounds error from short tile slice on SetTile, got nil")
	}
}

func TestJSONSerialization(t *testing.T) {
	p := NewHandshakeRequestPacket("2.1.0", "client-123")

	// Test raw byte marshaling/unmarshaling
	data, err := MarshalJSONPacket(p)
	if err != nil {
		t.Fatalf("failed to marshal JSON packet: %v", err)
	}

	decoded, err := UnmarshalJSONPacket(data)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON packet: %v", err)
	}

	if decoded.Type != p.Type {
		t.Errorf("expected type %s, got %s", p.Type, decoded.Type)
	}
	if decoded.HandshakeRequest.ClientID != p.HandshakeRequest.ClientID {
		t.Errorf("expected client ID %s, got %s", p.HandshakeRequest.ClientID, decoded.HandshakeRequest.ClientID)
	}

	// Test streaming operations (io.Writer/io.Reader)
	var buf bytes.Buffer
	err = EncodeJSON(&buf, p)
	if err != nil {
		t.Fatalf("failed to stream encode JSON packet: %v", err)
	}

	decodedStream, err := DecodeJSON(&buf)
	if err != nil {
		t.Fatalf("failed to stream decode JSON packet: %v", err)
	}

	if decodedStream.Type != p.Type {
		t.Errorf("expected type %s, got %s", p.Type, decodedStream.Type)
	}

	// Test encoding nil packet errors
	err = EncodeJSON(&buf, nil)
	if err == nil {
		t.Error("expected error when encoding nil packet via stream, got nil")
	}

	_, err = MarshalJSONPacket(nil)
	if err == nil {
		t.Error("expected error when marshaling nil packet, got nil")
	}

	// Test decoding invalid JSON data
	_, err = UnmarshalJSONPacket([]byte(`{invalid_json`))
	if err == nil {
		t.Error("expected error when unmarshaling invalid JSON data, got nil")
	}

	_, err = DecodeJSON(bytes.NewReader([]byte(`{invalid_json`)))
	if err == nil {
		t.Error("expected error when stream decoding invalid JSON data, got nil")
	}
}

func TestGobSerialization(t *testing.T) {
	m := tilemap.NewDefaultMap()
	chunk, err := NewTilemapChunk(0, 0, 5, 5, m)
	if err != nil {
		t.Fatalf("failed to prepare chunk: %v", err)
	}

	p := NewTilemapChunkPacket(chunk)

	// Test raw byte marshaling/unmarshaling
	data, err := MarshalGobPacket(p)
	if err != nil {
		t.Fatalf("failed to marshal Gob packet: %v", err)
	}

	decoded, err := UnmarshalGobPacket(data)
	if err != nil {
		t.Fatalf("failed to unmarshal Gob packet: %v", err)
	}

	if decoded.Type != p.Type {
		t.Errorf("expected type %s, got %s", p.Type, decoded.Type)
	}
	if decoded.Chunk.Width != chunk.Width || decoded.Chunk.Height != chunk.Height {
		t.Errorf("expected dimensions %dx%d, got %dx%d", chunk.Width, chunk.Height, decoded.Chunk.Width, decoded.Chunk.Height)
	}

	// Test streaming operations (io.Writer/io.Reader)
	var buf bytes.Buffer
	err = EncodeGob(&buf, p)
	if err != nil {
		t.Fatalf("failed to stream encode Gob packet: %v", err)
	}

	decodedStream, err := DecodeGob(&buf)
	if err != nil {
		t.Fatalf("failed to stream decode Gob packet: %v", err)
	}

	if decodedStream.Type != p.Type {
		t.Errorf("expected type %s, got %s", p.Type, decodedStream.Type)
	}

	// Test encoding nil packet errors
	err = EncodeGob(&buf, nil)
	if err == nil {
		t.Error("expected error when encoding nil packet via stream, got nil")
	}

	_, err = MarshalGobPacket(nil)
	if err == nil {
		t.Error("expected error when marshaling nil packet, got nil")
	}

	// Test decoding invalid Gob data
	_, err = UnmarshalGobPacket([]byte(`corrupt_gob_bytes`))
	if err == nil {
		t.Error("expected error when unmarshaling invalid Gob data, got nil")
	}

	_, err = DecodeGob(bytes.NewReader([]byte(`corrupt_gob_bytes`)))
	if err == nil {
		t.Error("expected error when stream decoding invalid Gob data, got nil")
	}
}

type dummyStruct struct {
	Name  string
	Value int
}

func TestGenericJSONSerialization(t *testing.T) {
	val := &dummyStruct{Name: "test", Value: 42}

	var buf bytes.Buffer
	err := EncodeGenericJSON(&buf, val)
	if err != nil {
		t.Fatalf("failed to stream encode generic JSON: %v", err)
	}

	decoded, err := DecodeGenericJSON[dummyStruct](&buf)
	if err != nil {
		t.Fatalf("failed to stream decode generic JSON: %v", err)
	}

	if decoded.Name != val.Name || decoded.Value != val.Value {
		t.Errorf("decoded generic JSON struct mismatch: got %+v, want %+v", decoded, val)
	}

	// Error cases
	err = EncodeGenericJSON[dummyStruct](&buf, nil)
	if err == nil {
		t.Error("expected error when encoding nil generic value to JSON, got nil")
	}

	_, err = DecodeGenericJSON[dummyStruct](bytes.NewReader([]byte(`{invalid`)))
	if err == nil {
		t.Error("expected error when decoding invalid JSON, got nil")
	}
}

func TestGenericGobSerialization(t *testing.T) {
	val := &dummyStruct{Name: "gob-test", Value: 100}

	var buf bytes.Buffer
	err := EncodeGenericGob(&buf, val)
	if err != nil {
		t.Fatalf("failed to stream encode generic Gob: %v", err)
	}

	decoded, err := DecodeGenericGob[dummyStruct](&buf)
	if err != nil {
		t.Fatalf("failed to stream decode generic Gob: %v", err)
	}

	if decoded.Name != val.Name || decoded.Value != val.Value {
		t.Errorf("decoded generic Gob struct mismatch: got %+v, want %+v", decoded, val)
	}

	// Error cases
	err = EncodeGenericGob[dummyStruct](&buf, nil)
	if err == nil {
		t.Error("expected error when encoding nil generic value to Gob, got nil")
	}

	_, err = DecodeGenericGob[dummyStruct](bytes.NewReader([]byte(`invalid_gob`)))
	if err == nil {
		t.Error("expected error when decoding invalid Gob, got nil")
	}
}

type errWriter struct{}

func (errWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestSerializationWriterErrors(t *testing.T) {
	p := NewHandshakeRequestPacket("1.0.0", "client-id")
	val := &dummyStruct{Name: "test", Value: 42}

	// Test write errors on non-generic methods
	if err := EncodeJSON(errWriter{}, p); err == nil {
		t.Error("expected JSON encoding error on writing failure, got nil")
	}

	if err := EncodeGob(errWriter{}, p); err == nil {
		t.Error("expected Gob encoding error on writing failure, got nil")
	}

	// Test write errors on generic methods
	if err := EncodeGenericJSON[dummyStruct](errWriter{}, val); err == nil {
		t.Error("expected generic JSON encoding error on writing failure, got nil")
	}

	if err := EncodeGenericGob[dummyStruct](errWriter{}, val); err == nil {
		t.Error("expected generic Gob encoding error on writing failure, got nil")
	}
}
