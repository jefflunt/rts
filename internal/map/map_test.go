package tilemap

import (
	"testing"
)

func TestNewDefaultMap(t *testing.T) {
	m := NewDefaultMap()
	if m == nil {
		t.Fatal("NewDefaultMap returned nil")
	}

	// Check dimensions
	for x := 0; x < MapWidth; x++ {
		for y := 0; y < MapHeight; y++ {
			tile := m.Tiles[x][y]
			if tile.Type != TileTypeGrassStandard && tile.Type != TileTypeGrassVariant1 && tile.Type != TileTypeGrassVariant2 {
				t.Errorf("Tile at (%d, %d) has unexpected type: %v", x, y, tile.Type)
			}
		}
	}
}

func TestIsValidTile(t *testing.T) {
	tests := []struct {
		tx, ty int
		want   bool
	}{
		{0, 0, true},
		{MapWidth - 1, MapHeight - 1, true},
		{128, 128, true},
		{-1, 0, false},
		{0, -1, false},
		{MapWidth, 0, false},
		{0, MapHeight, false},
		{-100, -100, false},
		{1000, 1000, false},
	}

	for _, tc := range tests {
		got := IsValidTile(tc.tx, tc.ty)
		if got != tc.want {
			t.Errorf("IsValidTile(%d, %d) = %v; want %v", tc.tx, tc.ty, got, tc.want)
		}
	}
}

func TestGetAndSetTile(t *testing.T) {
	m := NewDefaultMap()

	// Valid tile
	tileToSet := Tile{Type: TileTypeGrassVariant1, Variant: 42}
	err := m.SetTile(10, 20, tileToSet)
	if err != nil {
		t.Errorf("SetTile(10, 20) failed: %v", err)
	}

	gotTile, err := m.GetTile(10, 20)
	if err != nil {
		t.Errorf("GetTile(10, 20) failed: %v", err)
	}
	if gotTile.Type != tileToSet.Type || gotTile.Variant != tileToSet.Variant {
		t.Errorf("GetTile(10, 20) returned %+v; want %+v", gotTile, tileToSet)
	}

	// Invalid tile Set
	err = m.SetTile(-1, 0, tileToSet)
	if err == nil {
		t.Error("SetTile(-1, 0) expected error, got nil")
	}

	// Invalid tile Get
	_, err = m.GetTile(0, -1)
	if err == nil {
		t.Error("GetTile(0, -1) expected error, got nil")
	}
}

func TestCoordinateTranslation(t *testing.T) {
	tests := []struct {
		wx, wy float64
		tx, ty int
	}{
		{0.0, 0.0, 0, 0},
		{15.0, 15.0, 0, 0},
		{31.9, 31.9, 0, 0},
		{32.0, 32.0, 1, 1},
		{64.0, 96.0, 2, 3},
		{-1.0, -1.0, -1, -1},
	}

	for _, tc := range tests {
		tx, ty := WorldToTile(tc.wx, tc.wy)
		if tx != tc.tx || ty != tc.ty {
			t.Errorf("WorldToTile(%f, %f) = (%d, %d); want (%d, %d)", tc.wx, tc.wy, tx, ty, tc.tx, tc.ty)
		}

		// Since TileToWorld is the top-left of the tile, we only test exact conversions
		if tc.wx == float64(tc.tx*TileSize) && tc.wy == float64(tc.ty*TileSize) {
			wx, wy := TileToWorld(tc.tx, tc.ty)
			if wx != tc.wx || wy != tc.wy {
				t.Errorf("TileToWorld(%d, %d) = (%f, %f); want (%f, %f)", tc.tx, tc.ty, wx, wy, tc.wx, tc.wy)
			}
		}
	}
}

func TestTileCenterToWorld(t *testing.T) {
	tests := []struct {
		tx, ty int
		wx, wy float64
	}{
		{0, 0, 16.0, 16.0},
		{1, 2, 48.0, 80.0},
	}

	for _, tc := range tests {
		wx, wy := TileCenterToWorld(tc.tx, tc.ty)
		if wx != tc.wx || wy != tc.wy {
			t.Errorf("TileCenterToWorld(%d, %d) = (%f, %f); want (%f, %f)", tc.tx, tc.ty, wx, wy, tc.wx, tc.wy)
		}
	}
}

func TestFlatIndexConversion(t *testing.T) {
	tests := []struct {
		tx, ty int
		index  int
	}{
		{0, 0, 0},
		{1, 0, 1},
		{0, 1, 256},
		{255, 255, 65535},
	}

	for _, tc := range tests {
		idx := TileToIndex(tc.tx, tc.ty)
		if idx != tc.index {
			t.Errorf("TileToIndex(%d, %d) = %d; want %d", tc.tx, tc.ty, idx, tc.index)
		}

		tx, ty := IndexToTile(tc.index)
		if tx != tc.tx || ty != tc.ty {
			t.Errorf("IndexToTile(%d) = (%d, %d); want (%d, %d)", tc.index, tx, ty, tc.tx, tc.ty)
		}
	}
}

func TestWorldAndIndexCombined(t *testing.T) {
	wx, wy := 35.0, 70.0 // tx = 1, ty = 2
	// index = 2 * 256 + 1 = 513
	expectedIndex := 513

	idx := WorldToIndex(wx, wy)
	if idx != expectedIndex {
		t.Errorf("WorldToIndex(%f, %f) = %d; want %d", wx, wy, idx, expectedIndex)
	}

	rwx, rwy := IndexToWorld(expectedIndex)
	expectedWx, expectedWy := float64(1*TileSize), float64(2*TileSize)
	if rwx != expectedWx || rwy != expectedWy {
		t.Errorf("IndexToWorld(%d) = (%f, %f); want (%f, %f)", expectedIndex, rwx, rwy, expectedWx, expectedWy)
	}
}
