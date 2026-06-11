package tilemap

import (
	"errors"
	"math"
)

// Map represents a 256x256 grid map.
type Map struct {
	Tiles [MapWidth][MapHeight]Tile `json:"tiles"`
}

// NewDefaultMap creates and returns a 256x256 map populated with default grass tile patterns.
func NewDefaultMap() *Map {
	m := &Map{}
	for x := 0; x < MapWidth; x++ {
		for y := 0; y < MapHeight; y++ {
			// Populate with standard grass and some variants to make it visually interesting.
			tileType := TileTypeGrassStandard
			variant := uint8(0)

			// Simple procedural pattern (e.g. checkerboard or periodic speckles)
			if (x+y)%2 == 0 {
				tileType = TileTypeGrassVariant1
				variant = uint8((x * y) % 3)
			} else if (x*3+y*7)%11 == 0 {
				tileType = TileTypeGrassVariant2
				variant = uint8((x + y) % 4)
			}

			m.Tiles[x][y] = Tile{
				Type:    tileType,
				Variant: variant,
			}
		}
	}
	return m
}

// IsValidTile checks if the given tile coordinates (tx, ty) are within the map bounds.
func IsValidTile(tx, ty int) bool {
	return tx >= 0 && tx < MapWidth && ty >= 0 && ty < MapHeight
}

// GetTile returns the tile at the given tile coordinates.
// Returns an error if the coordinates are out of bounds.
func (m *Map) GetTile(tx, ty int) (Tile, error) {
	if !IsValidTile(tx, ty) {
		return Tile{}, errors.New("tile coordinates out of bounds")
	}
	return m.Tiles[tx][ty], nil
}

// SetTile sets the tile at the given tile coordinates.
// Returns an error if the coordinates are out of bounds.
func (m *Map) SetTile(tx, ty int, tile Tile) error {
	if !IsValidTile(tx, ty) {
		return errors.New("tile coordinates out of bounds")
	}
	m.Tiles[tx][ty] = tile
	return nil
}

// WorldToTile converts continuous 2D world coordinates (pixels) to tile grid coordinates.
// Continuous coordinates are divided by TileSize and floored to determine the tile cell.
func WorldToTile(wx, wy float64) (tx, ty int) {
	tx = int(math.Floor(wx / float64(TileSize)))
	ty = int(math.Floor(wy / float64(TileSize)))
	return tx, ty
}

// TileToWorld converts tile grid coordinates to continuous 2D world coordinates (top-left of the tile).
func TileToWorld(tx, ty int) (wx, wy float64) {
	wx = float64(tx * TileSize)
	wy = float64(ty * TileSize)
	return wx, wy
}

// TileCenterToWorld converts tile grid coordinates to continuous 2D world coordinates corresponding to the center of the tile.
func TileCenterToWorld(tx, ty int) (wx, wy float64) {
	wx = float64(tx)*TileSize + float64(TileSize)/2.0
	wy = float64(ty)*TileSize + float64(TileSize)/2.0
	return wx, wy
}

// TileToIndex converts 2D tile coordinates to a 1D flat index using row-major ordering.
func TileToIndex(tx, ty int) int {
	return ty*MapWidth + tx
}

// IndexToTile converts a 1D flat index back to 2D tile coordinates using row-major ordering.
func IndexToTile(index int) (tx, ty int) {
	tx = index % MapWidth
	ty = index / MapWidth
	return tx, ty
}

// WorldToIndex converts continuous 2D world coordinates to a 1D flat tile index.
func WorldToIndex(wx, wy float64) int {
	tx, ty := WorldToTile(wx, wy)
	return TileToIndex(tx, ty)
}

// IndexToWorld converts a 1D flat tile index to continuous 2D world coordinates (top-left of the tile).
func IndexToWorld(index int) (wx, wy float64) {
	tx, ty := IndexToTile(index)
	return TileToWorld(tx, ty)
}
