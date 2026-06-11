package client

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	tilemap "scrollable-tilemap/internal/map"
)

// Predefined distinct shades of green for tile types on the minimap
var (
	ColorStandard = color.RGBA{R: 34, G: 139, B: 34, A: 255}  // Forest Green
	ColorVariant1 = color.RGBA{R: 50, G: 205, B: 50, A: 255}  // Lime Green
	ColorVariant2 = color.RGBA{R: 0, G: 100, B: 0, A: 255}    // Dark Green
)

// Minimap represents the client-side minimap overlay.
// It occupies a fixed 256x256 area in the bottom-left corner of the client screen,
// offset by 16 pixels from the edges.
type Minimap struct {
	MapWidthTiles  int
	MapHeightTiles int
	TileSize       int
	CacheImage     *ebiten.Image
}

// NewMinimap creates a new Minimap instance with the given map dimensions and tile size.
func NewMinimap(mapWidthTiles, mapHeightTiles, tileSize int) *Minimap {
	return &Minimap{
		MapWidthTiles:  mapWidthTiles,
		MapHeightTiles: mapHeightTiles,
		TileSize:       tileSize,
		CacheImage:     ebiten.NewImage(256, 256),
	}
}

// MapWidthPx returns the total width of the map in pixels.
func (m *Minimap) MapWidthPx() float64 {
	return float64(m.MapWidthTiles * m.TileSize)
}

// MapHeightPx returns the total height of the map in pixels.
func (m *Minimap) MapHeightPx() float64 {
	return float64(m.MapHeightTiles * m.TileSize)
}

// WorldToMinimap translates a world coordinate in pixels into its corresponding position relative to the 256x256 minimap boundary.
// It returns relative/local minimap coordinates (rx, ry) in the range [0, 256].
func (m *Minimap) WorldToMinimap(wx, wy float64) (rx, ry float64) {
	mapWidthPx := m.MapWidthPx()
	mapHeightPx := m.MapHeightPx()

	if mapWidthPx > 0 {
		rx = (wx / mapWidthPx) * 256.0
	}
	if mapHeightPx > 0 {
		ry = (wy / mapHeightPx) * 256.0
	}

	// Clamp to the 256x256 minimap boundary
	if rx < 0 {
		rx = 0
	} else if rx > 256 {
		rx = 256
	}
	if ry < 0 {
		ry = 0
	} else if ry > 256 {
		ry = 256
	}

	return rx, ry
}

// WorldToMinimapUnclamped translates a world coordinate in pixels into its corresponding position on the minimap scale without clamping.
func (m *Minimap) WorldToMinimapUnclamped(wx, wy float64) (rx, ry float64) {
	mapWidthPx := m.MapWidthPx()
	mapHeightPx := m.MapHeightPx()

	if mapWidthPx > 0 {
		rx = (wx / mapWidthPx) * 256.0
	}
	if mapHeightPx > 0 {
		ry = (wy / mapHeightPx) * 256.0
	}

	return rx, ry
}

// MinimapToWorld converts a relative/local pixel coordinate on the 256x256 minimap back into absolute world coordinates in pixels, allowing for camera centering.
// It clamps input minimap coordinates to [0, 256] to ensure the resulting coordinates stay within world map bounds.
func (m *Minimap) MinimapToWorld(rx, ry float64) (wx, wy float64) {
	mapWidthPx := m.MapWidthPx()
	mapHeightPx := m.MapHeightPx()

	if rx < 0 {
		rx = 0
	} else if rx > 256 {
		rx = 256
	}
	if ry < 0 {
		ry = 0
	} else if ry > 256 {
		ry = 256
	}

	if mapWidthPx > 0 {
		wx = (rx / 256.0) * mapWidthPx
	}
	if mapHeightPx > 0 {
		wy = (ry / 256.0) * mapHeightPx
	}

	return wx, wy
}

// MinimapToWorldUnclamped converts a minimap pixel coordinate back into absolute world coordinates in pixels without clamping.
func (m *Minimap) MinimapToWorldUnclamped(rx, ry float64) (wx, wy float64) {
	mapWidthPx := m.MapWidthPx()
	mapHeightPx := m.MapHeightPx()

	wx = (rx / 256.0) * mapWidthPx
	wy = (ry / 256.0) * mapHeightPx

	return wx, wy
}

// ScreenToMinimap converts screen coordinates (sx, sy) to local minimap coordinates (rx, ry).
// It returns ok = true if the coordinate is within the 256x256 minimap screen boundary.
func (m *Minimap) ScreenToMinimap(sx, sy float64, screenHeight int) (rx, ry float64, ok bool) {
	screenMinimapX := 16.0
	screenMinimapY := float64(screenHeight) - 16.0 - 256.0

	rx = sx - screenMinimapX
	ry = sy - screenMinimapY

	if rx >= 0 && rx <= 256 && ry >= 0 && ry <= 256 {
		return rx, ry, true
	}
	return rx, ry, false
}

// MinimapToScreen converts local minimap coordinates (rx, ry) to screen coordinates (sx, sy).
func (m *Minimap) MinimapToScreen(rx, ry float64, screenHeight int) (sx, sy float64) {
	screenMinimapX := 16.0
	screenMinimapY := float64(screenHeight) - 16.0 - 256.0

	sx = screenMinimapX + rx
	sy = screenMinimapY + ry
	return sx, sy
}

// ScreenToWorld converts clicked screen coordinates (sx, sy) directly to absolute world coordinates in pixels.
// It returns ok = true if the click was within the minimap screen boundary.
func (m *Minimap) ScreenToWorld(sx, sy float64, screenHeight int) (wx, wy float64, ok bool) {
	rx, ry, ok := m.ScreenToMinimap(sx, sy, screenHeight)
	if !ok {
		// Even if not clicked inside, we can still translate the clamped coordinates
		wx, wy = m.MinimapToWorld(rx, ry)
		return wx, wy, false
	}
	wx, wy = m.MinimapToWorld(rx, ry)
	return wx, wy, true
}

// CalculateViewportIndicator computes the screen-space X, Y coordinates, width, and height of the white rectangle
// representing the camera's current visible viewport on top of the minimap screen area.
func (m *Minimap) CalculateViewportIndicator(camX, camY float64, viewportWidth, viewportHeight, screenHeight int) (x, y, width, height float64) {
	screenMinimapX := 16.0
	screenMinimapY := float64(screenHeight) - 16.0 - 256.0

	// Map top-left and bottom-right of viewport to minimap local space using clamping to ensure the indicator stays within bounds.
	rx1, ry1 := m.WorldToMinimap(camX, camY)
	rx2, ry2 := m.WorldToMinimap(camX+float64(viewportWidth), camY+float64(viewportHeight))

	x = screenMinimapX + rx1
	y = screenMinimapY + ry1
	width = rx2 - rx1
	height = ry2 - ry1

	return x, y, width, height
}

// ViewportIndicatorBounds computes the screen-space X, Y coordinates, width, and height of the white rectangle
// representing the camera's current visible viewport on top of the minimap screen area, using fields from the Camera struct.
func (m *Minimap) ViewportIndicatorBounds(cam *Camera) (x, y, width, height float64) {
	if cam == nil {
		return 0, 0, 0, 0
	}
	return m.CalculateViewportIndicator(cam.X, cam.Y, cam.ViewportWidth, cam.ViewportHeight, cam.ViewportHeight)
}

// GetTileColor maps a TileType to its predefined, distinct shade of green.
func (m *Minimap) GetTileColor(t tilemap.TileType) color.Color {
	switch t {
	case tilemap.TileTypeGrassStandard:
		return ColorStandard
	case tilemap.TileTypeGrassVariant1:
		return ColorVariant1
	case tilemap.TileTypeGrassVariant2:
		return ColorVariant2
	default:
		return ColorStandard
	}
}

// UpdateCache renders the tileMap's tiles into the 256x256 offscreen CacheImage.
func (m *Minimap) UpdateCache(tileMap *tilemap.Map) {
	if m.CacheImage == nil {
		m.CacheImage = ebiten.NewImage(256, 256)
	}

	if tileMap == nil {
		m.CacheImage.Clear()
		return
	}

	dotSizeX := 1
	if m.MapWidthTiles > 0 {
		dotSizeX = 256 / m.MapWidthTiles
	}
	if dotSizeX < 1 {
		dotSizeX = 1
	}

	dotSizeY := 1
	if m.MapHeightTiles > 0 {
		dotSizeY = 256 / m.MapHeightTiles
	}
	if dotSizeY < 1 {
		dotSizeY = 1
	}

	for tx := 0; tx < m.MapWidthTiles; tx++ {
		for ty := 0; ty < m.MapHeightTiles; ty++ {
			tile, err := tileMap.GetTile(tx, ty)
			if err != nil {
				continue
			}

			col := m.GetTileColor(tile.Type)
			for dx := 0; dx < dotSizeX; dx++ {
				for dy := 0; dy < dotSizeY; dy++ {
					m.CacheImage.Set(tx*dotSizeX+dx, ty*dotSizeY+dy, col)
				}
			}
		}
	}
}
