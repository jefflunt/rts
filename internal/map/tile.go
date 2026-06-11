package tilemap

// TileType represents the type/texture of a tile.
type TileType uint8

const (
	// TileSize is the width and height of a single tile in pixels.
	TileSize = 32

	// MapWidth is the width of the map in tiles.
	MapWidth = 256

	// MapHeight is the height of the map in tiles.
	MapHeight = 256
)

const (
	// TileTypeGrassStandard is the default standard grass tile.
	TileTypeGrassStandard TileType = iota

	// TileTypeGrassVariant1 is a grass tile variant.
	TileTypeGrassVariant1

	// TileTypeGrassVariant2 is another grass tile variant.
	TileTypeGrassVariant2
)

// Tile represents a single tile on the grid map.
type Tile struct {
	Type    TileType `json:"type"`
	Variant uint8    `json:"variant"` // Used to add procedural visual variations
}
