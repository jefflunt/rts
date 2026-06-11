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
	isDragging     bool
}

// NewMinimap creates a new Minimap instance with the given map dimensions and tile size.
func NewMinimap(mapWidthTiles, mapHeightTiles, tileSize int) *Minimap {
	return &Minimap{
		MapWidthTiles:  mapWidthTiles,
		MapHeightTiles: mapHeightTiles,
		TileSize:       tileSize,
		CacheImage:     ebiten.NewImage(256, 256),
		isDragging:     false,
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
	} else {
		m.CacheImage.Clear()
	}

	if tileMap == nil {
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

var whiteImage = ebiten.NewImage(1, 1)

func init() {
	whiteImage.Fill(color.White)
}

// drawRect draws a solid colored rectangle at screen coordinates (x, y) with specified width and height.
func drawRect(screen *ebiten.Image, x, y, width, height float64, clr color.Color) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(width, height)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(whiteImage, op)
}

// Draw draws the minimap cache, the double-lined metallic border, and the viewport indicator on the screen.
func (m *Minimap) Draw(screen *ebiten.Image, cam *Camera) {
	if cam == nil || m.CacheImage == nil {
		return
	}

	screenHeight := cam.ViewportHeight
	screenMinimapX := 16.0
	screenMinimapY := float64(screenHeight) - 16.0 - 256.0

	// 1. Draw background cache image
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenMinimapX, screenMinimapY)
	screen.DrawImage(m.CacheImage, op)

	// 2. Draw 2-pixel wide double metallic border around the minimap bounds
	// Inner border: 2-pixel thick, silver color, touching minimap edge
	silverColor := color.RGBA{R: 192, G: 192, B: 192, A: 255}
	// Left line
	drawRect(screen, screenMinimapX-2, screenMinimapY-2, 2, 260, silverColor)
	// Right line
	drawRect(screen, screenMinimapX+256, screenMinimapY-2, 2, 260, silverColor)
	// Top line
	drawRect(screen, screenMinimapX-2, screenMinimapY-2, 260, 2, silverColor)
	// Bottom line
	drawRect(screen, screenMinimapX-2, screenMinimapY+256, 260, 2, silverColor)

	// Outer border: 2-pixel thick, steel grey, with a 2-pixel gap outside the inner border
	steelColor := color.RGBA{R: 90, G: 90, B: 95, A: 255}
	// Left line
	drawRect(screen, screenMinimapX-6, screenMinimapY-6, 2, 268, steelColor)
	// Right line
	drawRect(screen, screenMinimapX+256+4, screenMinimapY-6, 2, 268, steelColor)
	// Top line
	drawRect(screen, screenMinimapX-6, screenMinimapY-6, 268, 2, steelColor)
	// Bottom line
	drawRect(screen, screenMinimapX-6, screenMinimapY+256+4, 268, 2, steelColor)

	// 3. Draw white camera viewport indicator rectangle
	vx, vy, vw, vh := m.ViewportIndicatorBounds(cam)
	if vw > 0 && vh > 0 {
		whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
		// Draw thin 1-pixel white outline for viewport indicator
		// Top line
		drawRect(screen, vx, vy, vw, 1, whiteColor)
		// Bottom line
		if vh > 1 {
			drawRect(screen, vx, vy+vh-1, vw, 1, whiteColor)
		}
		// Left and Right lines
		if vh > 2 {
			drawRect(screen, vx, vy+1, 1, vh-2, whiteColor)
			if vw > 1 {
				drawRect(screen, vx+vw-1, vy+1, 1, vh-2, whiteColor)
			}
		}
	}
}

// IsDragging returns true if the user is currently clicking and dragging on the minimap.
func (m *Minimap) IsDragging() bool {
	return m.isDragging
}

// Update handles interaction (click/drag) on the minimap to pan and clamp the main camera viewport.
func (m *Minimap) Update(cam *Camera, input InputProvider) {
	if cam == nil || input == nil {
		m.isDragging = false
		return
	}

	if input.IsMouseButtonLeftPressed() {
		mx, my := input.CursorPosition()
		fx, fy := float64(mx), float64(my)

		if !m.isDragging {
			// Check if the click initiated inside the minimap area
			_, _, ok := m.ScreenToMinimap(fx, fy, cam.ViewportHeight)
			if ok {
				m.isDragging = true
			}
		}

		if m.isDragging {
			// Translate screen coordinates to absolute world coordinates.
			// ScreenToWorld clamps relative coordinates to minimap boundaries [0, 256].
			wx, wy, _ := m.ScreenToWorld(fx, fy, cam.ViewportHeight)

			// Center the camera viewport on (wx, wy)
			targetX := wx - float64(cam.ViewportWidth)/2.0
			targetY := wy - float64(cam.ViewportHeight)/2.0
			cam.SetPosition(targetX, targetY)
		}
	} else {
		m.isDragging = false
	}
}
