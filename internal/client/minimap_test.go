package client

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	tilemap "scrollable-tilemap/internal/map"
)

func TestNewMinimap(t *testing.T) {
	m := NewMinimap(128, 128, 16)
	if m.MapWidthTiles != 128 {
		t.Errorf("expected MapWidthTiles to be 128, got %d", m.MapWidthTiles)
	}
	if m.MapHeightTiles != 128 {
		t.Errorf("expected MapHeightTiles to be 128, got %d", m.MapHeightTiles)
	}
	if m.TileSize != 16 {
		t.Errorf("expected TileSize to be 16, got %d", m.TileSize)
	}

	if w := m.MapWidthPx(); w != 2048 {
		t.Errorf("expected MapWidthPx to be 2048, got %f", w)
	}
	if h := m.MapHeightPx(); h != 2048 {
		t.Errorf("expected MapHeightPx to be 2048, got %f", h)
	}
}

func TestWorldToMinimap(t *testing.T) {
	m := NewMinimap(256, 256, 32) // 8192 x 8192 px

	tests := []struct {
		name       string
		wx, wy     float64
		expectedRx float64
		expectedRy float64
	}{
		{"Middle of map", 4096, 4096, 128, 128},
		{"Top-left corner", 0, 0, 0, 0},
		{"Bottom-right corner", 8192, 8192, 256, 256},
		{"Clamped negative", -100, -200, 0, 0},
		{"Clamped out of bounds", 10000, 12000, 256, 256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rx, ry := m.WorldToMinimap(tt.wx, tt.wy)
			if rx != tt.expectedRx || ry != tt.expectedRy {
				t.Errorf("WorldToMinimap(%f, %f) = (%f, %f); want (%f, %f)", tt.wx, tt.wy, rx, ry, tt.expectedRx, tt.expectedRy)
			}
		})
	}
}

func TestWorldToMinimapUnclamped(t *testing.T) {
	m := NewMinimap(256, 256, 32) // 8192 x 8192 px

	tests := []struct {
		name       string
		wx, wy     float64
		expectedRx float64
		expectedRy float64
	}{
		{"Middle of map", 4096, 4096, 128, 128},
		{"Top-left corner", 0, 0, 0, 0},
		{"Bottom-right corner", 8192, 8192, 256, 256},
		{"Negative values unclamped", -4096, -8192, -128, -256},
		{"Out of bounds unclamped", 16384, 16384, 512, 512},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rx, ry := m.WorldToMinimapUnclamped(tt.wx, tt.wy)
			if rx != tt.expectedRx || ry != tt.expectedRy {
				t.Errorf("WorldToMinimapUnclamped(%f, %f) = (%f, %f); want (%f, %f)", tt.wx, tt.wy, rx, ry, tt.expectedRx, tt.expectedRy)
			}
		})
	}
}

func TestMinimapToWorld(t *testing.T) {
	m := NewMinimap(256, 256, 32) // 8192 x 8192 px

	tests := []struct {
		name       string
		rx, ry     float64
		expectedWx float64
		expectedWy float64
	}{
		{"Middle of minimap", 128, 128, 4096, 4096},
		{"Top-left corner", 0, 0, 0, 0},
		{"Bottom-right corner", 256, 256, 8192, 8192},
		{"Clamped negative", -50, -100, 0, 0},
		{"Clamped out of bounds", 300, 500, 8192, 8192},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wx, wy := m.MinimapToWorld(tt.rx, tt.ry)
			if wx != tt.expectedWx || wy != tt.expectedWy {
				t.Errorf("MinimapToWorld(%f, %f) = (%f, %f); want (%f, %f)", tt.rx, tt.ry, wx, wy, tt.expectedWx, tt.expectedWy)
			}
		})
	}
}

func TestMinimapToWorldUnclamped(t *testing.T) {
	m := NewMinimap(256, 256, 32) // 8192 x 8192 px

	tests := []struct {
		name       string
		rx, ry     float64
		expectedWx float64
		expectedWy float64
	}{
		{"Middle of minimap", 128, 128, 4096, 4096},
		{"Top-left corner", 0, 0, 0, 0},
		{"Bottom-right corner", 256, 256, 8192, 8192},
		{"Negative values unclamped", -128, -256, -4096, -8192},
		{"Out of bounds unclamped", 512, 512, 16384, 16384},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wx, wy := m.MinimapToWorldUnclamped(tt.rx, tt.ry)
			if wx != tt.expectedWx || wy != tt.expectedWy {
				t.Errorf("MinimapToWorldUnclamped(%f, %f) = (%f, %f); want (%f, %f)", tt.rx, tt.ry, wx, wy, tt.expectedWx, tt.expectedWy)
			}
		})
	}
}

func TestZeroMapDimensions(t *testing.T) {
	m := NewMinimap(0, 0, 0)
	rx, ry := m.WorldToMinimap(100, 100)
	if rx != 0 || ry != 0 {
		t.Errorf("expected 0 for zero map, got (%f, %f)", rx, ry)
	}

	rxUnclamped, ryUnclamped := m.WorldToMinimapUnclamped(100, 100)
	if rxUnclamped != 0 || ryUnclamped != 0 {
		t.Errorf("expected 0 for zero map unclamped, got (%f, %f)", rxUnclamped, ryUnclamped)
	}

	wx, wy := m.MinimapToWorld(128, 128)
	if wx != 0 || wy != 0 {
		t.Errorf("expected 0 for zero map minimap-to-world, got (%f, %f)", wx, wy)
	}

	wxUnclamped, wyUnclamped := m.MinimapToWorldUnclamped(128, 128)
	if wxUnclamped != 0 || wyUnclamped != 0 {
		t.Errorf("expected 0 for zero map minimap-to-world unclamped, got (%f, %f)", wxUnclamped, wyUnclamped)
	}
}

func TestScreenToMinimap(t *testing.T) {
	m := NewMinimap(256, 256, 32)
	screenHeight := 600

	// screenMinimapX = 16
	// screenMinimapY = 600 - 16 - 256 = 328

	tests := []struct {
		name         string
		sx, sy       float64
		expectedRx   float64
		expectedRy   float64
		expectedOk   bool
	}{
		{"Minimap top-left screen position", 16, 328, 0, 0, true},
		{"Minimap bottom-right screen position", 272, 584, 256, 256, true},
		{"Minimap center screen position", 144, 456, 128, 128, true},
		{"Left out of bounds", 10, 456, -6, 128, false},
		{"Right out of bounds", 280, 456, 264, 128, false},
		{"Top out of bounds", 144, 320, 128, -8, false},
		{"Bottom out of bounds", 144, 590, 128, 262, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rx, ry, ok := m.ScreenToMinimap(tt.sx, tt.sy, screenHeight)
			if ok != tt.expectedOk {
				t.Errorf("ScreenToMinimap(%f, %f) ok = %v; want %v", tt.sx, tt.sy, ok, tt.expectedOk)
			}
			if rx != tt.expectedRx || ry != tt.expectedRy {
				t.Errorf("ScreenToMinimap(%f, %f) = (%f, %f); want (%f, %f)", tt.sx, tt.sy, rx, ry, tt.expectedRx, tt.expectedRy)
			}
		})
	}
}

func TestMinimapToScreen(t *testing.T) {
	m := NewMinimap(256, 256, 32)
	screenHeight := 600

	// screenMinimapX = 16
	// screenMinimapY = 328

	tests := []struct {
		name       string
		rx, ry     float64
		expectedSx float64
		expectedSy float64
	}{
		{"Minimap local top-left", 0, 0, 16, 328},
		{"Minimap local bottom-right", 256, 256, 272, 584},
		{"Minimap local center", 128, 128, 144, 456},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sx, sy := m.MinimapToScreen(tt.rx, tt.ry, screenHeight)
			if sx != tt.expectedSx || sy != tt.expectedSy {
				t.Errorf("MinimapToScreen(%f, %f) = (%f, %f); want (%f, %f)", tt.rx, tt.ry, sx, sy, tt.expectedSx, tt.expectedSy)
			}
		})
	}
}

func TestScreenToWorld(t *testing.T) {
	m := NewMinimap(256, 256, 32) // 8192 x 8192 px
	screenHeight := 600

	// screenMinimapX = 16
	// screenMinimapY = 328

	tests := []struct {
		name         string
		sx, sy       float64
		expectedWx   float64
		expectedWy   float64
		expectedOk   bool
	}{
		{"Minimap top-left click", 16, 328, 0, 0, true},
		{"Minimap bottom-right click", 272, 584, 8192, 8192, true},
		{"Minimap center click", 144, 456, 4096, 4096, true},
		{"Outside left click (clamped)", 10, 456, 0, 4096, false},
		{"Outside top click (clamped)", 144, 300, 4096, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wx, wy, ok := m.ScreenToWorld(tt.sx, tt.sy, screenHeight)
			if ok != tt.expectedOk {
				t.Errorf("ScreenToWorld(%f, %f) ok = %v; want %v", tt.sx, tt.sy, ok, tt.expectedOk)
			}
			if wx != tt.expectedWx || wy != tt.expectedWy {
				t.Errorf("ScreenToWorld(%f, %f) = (%f, %f); want (%f, %f)", tt.sx, tt.sy, wx, wy, tt.expectedWx, tt.expectedWy)
			}
		})
	}
}

func TestCalculateViewportIndicator(t *testing.T) {
	m := NewMinimap(256, 256, 32) // 8192 x 8192 px
	screenHeight := 600

	// screenMinimapX = 16
	// screenMinimapY = 328
	// camera top-left at (0, 0), viewport width 800, viewport height 600
	// rx1, ry1 = WorldToMinimap(0, 0) = (0, 0)
	// rx2, ry2 = WorldToMinimap(800, 600) = (800/8192*256, 600/8192*256) = (25, 18.75)
	// x = 16 + 0 = 16
	// y = 328 + 0 = 328
	// w = 25 - 0 = 25
	// h = 18.75 - 0 = 18.75

	x, y, w, h := m.CalculateViewportIndicator(0, 0, 800, 600, screenHeight)
	if x != 16.0 {
		t.Errorf("expected x to be 16.0, got %f", x)
	}
	if y != 328.0 {
		t.Errorf("expected y to be 328.0, got %f", y)
	}
	if math.Abs(w-25.0) > 1e-9 {
		t.Errorf("expected width to be 25.0, got %f", w)
	}
	if math.Abs(h-18.75) > 1e-9 {
		t.Errorf("expected height to be 18.75, got %f", h)
	}
}

func TestViewportIndicatorBounds(t *testing.T) {
	m := NewMinimap(256, 256, 32)

	// Nil Camera
	x, y, w, h := m.ViewportIndicatorBounds(nil)
	if x != 0 || y != 0 || w != 0 || h != 0 {
		t.Errorf("expected zeroes for nil camera, got (%f, %f, %f, %f)", x, y, w, h)
	}

	// Valid Camera
	cam := NewCamera(800, 600, 300.0, 10)
	cam.SetPosition(4096, 4096) // middle of world

	// ViewportIndicatorBounds calls:
	// m.CalculateViewportIndicator(cam.X, cam.Y, cam.ViewportWidth, cam.ViewportHeight, cam.ViewportHeight)
	// cam.ViewportHeight = 600 as the screen height
	// cam.X, cam.Y = 4096, 4096
	// rx1, ry1 = WorldToMinimap(4096, 4096) = (128, 128)
	// rx2, ry2 = WorldToMinimap(4096+800, 4096+600) = (4896/8192*256, 4696/8192*256) = (153, 146.75)
	// screenMinimapY = 600 - 16 - 256 = 328
	// x = 16 + 128 = 144
	// y = 328 + 128 = 456
	// w = 153 - 128 = 25
	// h = 146.75 - 128 = 18.75

	x2, y2, w2, h2 := m.ViewportIndicatorBounds(cam)
	if x2 != 144.0 {
		t.Errorf("expected x to be 144.0, got %f", x2)
	}
	if y2 != 456.0 {
		t.Errorf("expected y to be 456.0, got %f", y2)
	}
	if math.Abs(w2-25.0) > 1e-9 {
		t.Errorf("expected width to be 25.0, got %f", w2)
	}
	if math.Abs(h2-18.75) > 1e-9 {
		t.Errorf("expected height to be 18.75, got %f", h2)
	}
}

func TestNewMinimapCacheImage(t *testing.T) {
	m := NewMinimap(128, 128, 16)
	if m.CacheImage == nil {
		t.Fatal("expected CacheImage to be initialized, got nil")
	}
	bounds := m.CacheImage.Bounds()
	if bounds.Dx() != 256 || bounds.Dy() != 256 {
		t.Errorf("expected CacheImage bounds to be 256x256, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGetTileColor(t *testing.T) {
	m := NewMinimap(256, 256, 16)

	tests := []struct {
		tileType tilemap.TileType
		expected color.RGBA
	}{
		{tilemap.TileTypeGrassStandard, ColorStandard},
		{tilemap.TileTypeGrassVariant1, ColorVariant1},
		{tilemap.TileTypeGrassVariant2, ColorVariant2},
		{tilemap.TileType(99), ColorStandard}, // Default fallback
	}

	for _, tt := range tests {
		res := m.GetTileColor(tt.tileType)
		resRGBA := color.RGBAModel.Convert(res).(color.RGBA)
		if resRGBA != tt.expected {
			t.Errorf("GetTileColor(%d) = %v, expected %v", tt.tileType, resRGBA, tt.expected)
		}
	}
}

func TestUpdateCacheValidSizes(t *testing.T) {
	// Ebiten doesn't support reading pixels in headless tests without a running game window,
	// so we verify that UpdateCache runs successfully without panics for all supported dimensions,
	// properly initializes/clears the cache, and maintains correct image dimensions.
	
	dimensions := []struct {
		width, height int
	}{
		{64, 64},
		{128, 128},
		{256, 256},
		{0, 0}, // Zero sizes boundary check
	}

	for _, dim := range dimensions {
		t.Run("dim", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("UpdateCache panicked for dimensions %dx%d: %v", dim.width, dim.height, r)
				}
			}()

			m := NewMinimap(dim.width, dim.height, 32)
			tileMap := tilemap.NewDefaultMap() // Populate standard default map

			m.UpdateCache(tileMap)

			if m.CacheImage == nil {
				t.Error("expected CacheImage to be non-nil after UpdateCache")
			}
			bounds := m.CacheImage.Bounds()
			if bounds.Dx() != 256 || bounds.Dy() != 256 {
				t.Errorf("expected CacheImage bounds to be 256x256, got %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestUpdateCacheNilMap(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UpdateCache panicked for nil map: %v", r)
		}
	}()

	m := NewMinimap(256, 256, 16)
	m.UpdateCache(nil)

	if m.CacheImage == nil {
		t.Error("expected CacheImage to be non-nil after UpdateCache(nil)")
	}
	bounds := m.CacheImage.Bounds()
	if bounds.Dx() != 256 || bounds.Dy() != 256 {
		t.Errorf("expected CacheImage bounds to be 256x256, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestMinimapDraw(t *testing.T) {
	m := NewMinimap(256, 256, 32)
	cam := NewCamera(800, 600, 300.0, 10)
	screen := ebiten.NewImage(800, 600)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Minimap.Draw panicked: %v", r)
		}
	}()

	m.Draw(screen, cam)
}

func TestMinimapUpdateNilParams(t *testing.T) {
	m := NewMinimap(256, 256, 32)
	cam := NewCamera(800, 600, 300.0, 10)
	mock := &mockInputProvider{click: true, mx: 144, my: 456}

	// Nil Camera
	m.isDragging = true
	m.Update(nil, mock)
	if m.IsDragging() {
		t.Error("expected IsDragging to be false when cam is nil")
	}

	// Nil Input
	m.isDragging = true
	m.Update(cam, nil)
	if m.IsDragging() {
		t.Error("expected IsDragging to be false when input is nil")
	}
}

func TestMinimapUpdateClickAndDragInside(t *testing.T) {
	m := NewMinimap(256, 256, 32) // 8192 x 8192 px
	cam := NewCamera(800, 600, 300.0, 10)
	cam.SetPosition(1000, 1000)

	// screenMinimapX = 16
	// screenMinimapY = 600 - 16 - 256 = 328
	// Middle of minimap: mx = 16 + 128 = 144, my = 328 + 128 = 456
	// This maps to relative rx = 128, ry = 128
	// Which maps to world: wx = 4096, wy = 4096
	// Centering camera (800x600) on (4096, 4096):
	// targetX = 4096 - 400 = 3696
	// targetY = 4096 - 300 = 3796

	mock := &mockInputProvider{
		click: true,
		mx:    144,
		my:    456,
	}

	if m.IsDragging() {
		t.Error("expected IsDragging to be false initially")
	}

	m.Update(cam, mock)

	if !m.IsDragging() {
		t.Error("expected IsDragging to be true after clicking inside the minimap")
	}

	if cam.X != 3696 || cam.Y != 3796 {
		t.Errorf("expected camera position to be (3696, 3796), got (%f, %f)", cam.X, cam.Y)
	}
}

func TestMinimapUpdateClickOutside(t *testing.T) {
	m := NewMinimap(256, 256, 32)
	cam := NewCamera(800, 600, 300.0, 10)
	cam.SetPosition(1000, 1000)

	// Outside position: mx = 10, my = 10
	mock := &mockInputProvider{
		click: true,
		mx:    10,
		my:    10,
	}

	m.Update(cam, mock)

	if m.IsDragging() {
		t.Error("expected IsDragging to remain false after clicking outside the minimap")
	}

	if cam.X != 1000 || cam.Y != 1000 {
		t.Errorf("expected camera position to remain unchanged at (1000, 1000), got (%f, %f)", cam.X, cam.Y)
	}
}

func TestMinimapUpdateContinueDragOutside(t *testing.T) {
	m := NewMinimap(256, 256, 32) // 8192 x 8192 px
	cam := NewCamera(800, 600, 300.0, 10)
	cam.SetPosition(1000, 1000)

	// Step 1: Click inside to initiate dragging
	mock := &mockInputProvider{
		click: true,
		mx:    144,
		my:    456,
	}
	m.Update(cam, mock)

	if !m.IsDragging() {
		t.Fatal("expected drag to initiate")
	}

	// Step 2: Drag outside to the left (mx = 5, my = 456)
	// This should map to rx = 5 - 16 = -11, ry = 128
	// rx clamps to 0, ry = 128
	// Which maps to world: wx = 0, wy = 4096
	// Centering camera (800x600) on (0, 4096):
	// targetX = 0 - 400 = -400 -> clamps to 0
	// targetY = 4096 - 300 = 3796
	mock.mx = 5
	m.Update(cam, mock)

	if !m.IsDragging() {
		t.Error("expected IsDragging to remain true when moving outside while holding click")
	}

	if cam.X != 0 || cam.Y != 3796 {
		t.Errorf("expected camera position to be clamped to (0, 3796), got (%f, %f)", cam.X, cam.Y)
	}
}

func TestMinimapUpdateReleaseButton(t *testing.T) {
	m := NewMinimap(256, 256, 32)
	cam := NewCamera(800, 600, 300.0, 10)
	cam.SetPosition(1000, 1000)

	// Step 1: Drag is active
	m.isDragging = true

	// Step 2: Release button (click = false), mouse moves to (144, 456)
	mock := &mockInputProvider{
		click: false,
		mx:    144,
		my:    456,
	}

	m.Update(cam, mock)

	if m.IsDragging() {
		t.Error("expected IsDragging to become false after mouse button is released")
	}

	if cam.X != 1000 || cam.Y != 1000 {
		t.Errorf("expected camera position to remain unchanged at (1000, 1000), got (%f, %f)", cam.X, cam.Y)
	}
}

func TestMinimapScalingAndCoordinates(t *testing.T) {
	sizes := []struct {
		widthTiles, heightTiles int
		tileSize                int
		expectedPx              float64
	}{
		{64, 64, 32, 2048},
		{128, 128, 32, 4096},
		{256, 256, 32, 8192},
	}

	for _, sz := range sizes {
		m := NewMinimap(sz.widthTiles, sz.heightTiles, sz.tileSize)
		if m.MapWidthPx() != sz.expectedPx {
			t.Errorf("expected map width in px to be %f for %dx%d map, got %f", sz.expectedPx, sz.widthTiles, sz.heightTiles, m.MapWidthPx())
		}
		if m.MapHeightPx() != sz.expectedPx {
			t.Errorf("expected map height in px to be %f for %dx%d map, got %f", sz.expectedPx, sz.widthTiles, sz.heightTiles, m.MapHeightPx())
		}

		// Test center of map
		rx, ry := m.WorldToMinimap(sz.expectedPx/2, sz.expectedPx/2)
		if rx != 128 || ry != 128 {
			t.Errorf("expected center of %dx%d map to map to (128, 128) on minimap, got (%f, %f)", sz.widthTiles, sz.heightTiles, rx, ry)
		}

		// Test map to world from minimap center
		wx, wy := m.MinimapToWorld(128, 128)
		if wx != sz.expectedPx/2 || wy != sz.expectedPx/2 {
			t.Errorf("expected minimap center (128, 128) to map to center of %dx%d world, got (%f, %f)", sz.widthTiles, sz.heightTiles, wx, wy)
		}
	}
}

func TestMinimapClickAndDragDifferentMapSizes(t *testing.T) {
	sizes := []struct {
		widthTiles, heightTiles int
		tileSize                int
		expectedCamX            float64
		expectedCamY            float64
	}{
		{64, 64, 32, 624, 724},      // 2048x2048, center wx=1024, wy=1024, cam = wx - vpW/2 = 1024-400 = 624, wy - vpH/2 = 1024-300 = 724
		{128, 128, 32, 1648, 1748},  // 4096x4096, center wx=2048, wy=2048, cam = 2048-400 = 1648, 2048-300 = 1748
		{256, 256, 32, 3696, 3796},  // 8192x8192, center wx=4096, wy=4096, cam = 4096-400 = 3696, 4096-300 = 3796
	}

	for _, sz := range sizes {
		m := NewMinimap(sz.widthTiles, sz.heightTiles, sz.tileSize)
		cam := NewCamera(800, 600, 300.0, 10)
		cam.SetMapDimensions(sz.widthTiles, sz.heightTiles, sz.tileSize)
		cam.SetPosition(0, 0)

		// Click exactly in the middle of the minimap
		// screenHeight = 600
		// screenMinimapY = 600 - 16 - 256 = 328
		// Middle: mx = 16 + 128 = 144, my = 328 + 128 = 456
		mock := &mockInputProvider{
			click: true,
			mx:    144,
			my:    456,
		}

		m.Update(cam, mock)

		if !m.IsDragging() {
			t.Errorf("expected dragging to be true for %dx%d map", sz.widthTiles, sz.heightTiles)
		}

		if cam.X != sz.expectedCamX || cam.Y != sz.expectedCamY {
			t.Errorf("for %dx%d map, expected camera position to be (%f, %f), got (%f, %f)", sz.widthTiles, sz.heightTiles, sz.expectedCamX, sz.expectedCamY, cam.X, cam.Y)
		}
	}
}
