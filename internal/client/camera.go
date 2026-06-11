package client

import (
	"scrollable-tilemap/internal/map"
)

// Camera represents a client-side viewport camera that manages coordinate
// conversions, panning movement (clamped to the map boundaries), and viewport culling.
type Camera struct {
	// X and Y represent the top-left corner of the camera viewport in world coordinates (pixels).
	X float64
	Y float64

	// Viewport dimensions in pixels (screen/window size).
	ViewportWidth  int
	ViewportHeight int

	// Map dimensions in pixels (e.g. 256 tiles * 32 pixels/tile = 8192 pixels).
	MapWidthPx  int
	MapHeightPx int

	// ScrollSpeed is the speed at which the camera pans (in pixels per second).
	ScrollSpeed float64

	// ScrollMargin is the distance from the window boundary (in pixels)
	// that triggers mouse-edge panning.
	ScrollMargin int
}

// NewCamera creates a new Camera with the given viewport and map configuration.
func NewCamera(viewportWidth, viewportHeight int, scrollSpeed float64, scrollMargin int) *Camera {
	mapWidthPx := tilemap.MapWidth * tilemap.TileSize
	mapHeightPx := tilemap.MapHeight * tilemap.TileSize

	c := &Camera{
		X:              0,
		Y:              0,
		ViewportWidth:  viewportWidth,
		ViewportHeight: viewportHeight,
		MapWidthPx:     mapWidthPx,
		MapHeightPx:    mapHeightPx,
		ScrollSpeed:    scrollSpeed,
		ScrollMargin:   scrollMargin,
	}
	c.ClampPosition()
	return c
}

// SetPosition sets the camera coordinates and clamps them to ensure the viewport
// stays strictly within the map boundaries.
func (c *Camera) SetPosition(x, y float64) {
	c.X = x
	c.Y = y
	c.ClampPosition()
}

// ClampPosition restricts the camera coordinates to prevent the viewport
// from panning outside the map bounds.
func (c *Camera) ClampPosition() {
	maxX := float64(c.MapWidthPx - c.ViewportWidth)
	maxY := float64(c.MapHeightPx - c.ViewportHeight)

	// If the map is smaller than the viewport, clamp to 0.
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}

	if c.X < 0 {
		c.X = 0
	} else if c.X > maxX {
		c.X = maxX
	}

	if c.Y < 0 {
		c.Y = 0
	} else if c.Y > maxY {
		c.Y = maxY
	}
}

// WorldToScreen maps world coordinates (pixels) to screen/viewport coordinates (pixels).
func (c *Camera) WorldToScreen(wx, wy float64) (sx, sy float64) {
	sx = wx - c.X
	sy = wy - c.Y
	return sx, sy
}

// ScreenToWorld maps screen/viewport coordinates (pixels) to world coordinates (pixels).
func (c *Camera) ScreenToWorld(sx, sy float64) (wx, wy float64) {
	wx = sx + c.X
	wy = sy + c.Y
	return wx, wy
}

// Update updates the camera position based on keyboard arrow keys and mouse-edge panning,
// then clamps the camera position to keep the viewport within the map boundaries.
// dt is the delta time in seconds (e.g. 1/60 for a 60fps update loop).
func (c *Camera) Update(input InputProvider, dt float64) {
	if input == nil {
		return
	}

	var dx, dy float64

	// 1. Keyboard Panning (Arrow Keys)
	if input.IsArrowLeftPressed() {
		dx -= c.ScrollSpeed * dt
	}
	if input.IsArrowRightPressed() {
		dx += c.ScrollSpeed * dt
	}
	if input.IsArrowUpPressed() {
		dy -= c.ScrollSpeed * dt
	}
	if input.IsArrowDownPressed() {
		dy += c.ScrollSpeed * dt
	}

	// 2. Mouse-Edge Panning
	mx, my := input.CursorPosition()
	// Only trigger edge panning if the cursor is within the viewport bounds.
	if mx >= 0 && mx <= c.ViewportWidth && my >= 0 && my <= c.ViewportHeight {
		if mx <= c.ScrollMargin {
			dx -= c.ScrollSpeed * dt
		} else if mx >= c.ViewportWidth-c.ScrollMargin {
			dx += c.ScrollSpeed * dt
		}

		if my <= c.ScrollMargin {
			dy -= c.ScrollSpeed * dt
		} else if my >= c.ViewportHeight-c.ScrollMargin {
			dy += c.ScrollSpeed * dt
		}
	}

	// Apply movement
	c.X += dx
	c.Y += dy

	// Keep within map boundaries
	c.ClampPosition()
}

// GetVisibleTiles calculates the minimum and maximum column and row indices of tiles
// that are currently visible within the camera's viewport.
// The returned indices are inclusive and clamped to the map boundaries.
func (c *Camera) GetVisibleTiles() (minCol, minRow, maxCol, maxRow int) {
	minCol, minRow = tilemap.WorldToTile(c.X, c.Y)
	maxCol, maxRow = tilemap.WorldToTile(c.X+float64(c.ViewportWidth), c.Y+float64(c.ViewportHeight))

	// Clamp the indices to the valid tilemap bounds.
	if minCol < 0 {
		minCol = 0
	}
	if minRow < 0 {
		minRow = 0
	}
	if maxCol >= tilemap.MapWidth {
		maxCol = tilemap.MapWidth - 1
	}
	if maxRow >= tilemap.MapHeight {
		maxRow = tilemap.MapHeight - 1
	}

	// Just in case max is somehow less than min due to weird viewport/position values
	if maxCol < minCol {
		maxCol = minCol
	}
	if maxRow < minRow {
		maxRow = minRow
	}

	return minCol, minRow, maxCol, maxRow
}
