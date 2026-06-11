package client

import (
	"testing"
)

type mockInputProvider struct {
	up    bool
	down  bool
	left  bool
	right bool
	mx    int
	my    int
}

func (m *mockInputProvider) IsArrowUpPressed() bool {
	return m.up
}

func (m *mockInputProvider) IsArrowDownPressed() bool {
	return m.down
}

func (m *mockInputProvider) IsArrowLeftPressed() bool {
	return m.left
}

func (m *mockInputProvider) IsArrowRightPressed() bool {
	return m.right
}

func (m *mockInputProvider) CursorPosition() (int, int) {
	return m.mx, m.my
}

func TestNewCamera(t *testing.T) {
	// Viewport 800x600, scroll speed 300 px/sec, margin 10 px
	cam := NewCamera(800, 600, 300.0, 10)

	if cam.ViewportWidth != 800 || cam.ViewportHeight != 600 {
		t.Errorf("expected viewport 800x600, got %dx%d", cam.ViewportWidth, cam.ViewportHeight)
	}
	if cam.ScrollSpeed != 300.0 {
		t.Errorf("expected ScrollSpeed 300.0, got %f", cam.ScrollSpeed)
	}
	if cam.ScrollMargin != 10 {
		t.Errorf("expected ScrollMargin 10, got %d", cam.ScrollMargin)
	}
	if cam.X != 0 || cam.Y != 0 {
		t.Errorf("expected initial position (0, 0), got (%f, %f)", cam.X, cam.Y)
	}
}

func TestCameraClampPosition(t *testing.T) {
	cam := NewCamera(800, 600, 300.0, 10)

	// Test negative clamping
	cam.SetPosition(-100, -200)
	if cam.X != 0 || cam.Y != 0 {
		t.Errorf("expected negative coordinates to clamp to (0,0), got (%f, %f)", cam.X, cam.Y)
	}

	// Test clamping to maximum bound
	// Max X should be MapWidthPx - ViewportWidth = 256*32 - 800 = 8192 - 800 = 7392
	// Max Y should be MapHeightPx - ViewportHeight = 256*32 - 600 = 8192 - 600 = 7592
	cam.SetPosition(10000, 10000)
	if cam.X != 7392 || cam.Y != 7592 {
		t.Errorf("expected coordinates to clamp to (7392, 7592), got (%f, %f)", cam.X, cam.Y)
	}

	// Test intermediate valid coordinates
	cam.SetPosition(500, 600)
	if cam.X != 500 || cam.Y != 600 {
		t.Errorf("expected coordinates to be (500, 600), got (%f, %f)", cam.X, cam.Y)
	}
}

func TestCoordinateConversion(t *testing.T) {
	cam := NewCamera(800, 600, 300.0, 10)
	cam.SetPosition(200, 300)

	// WorldToScreen
	// World point (250, 400) relative to camera (200, 300) should be screen point (50, 100)
	sx, sy := cam.WorldToScreen(250, 400)
	if sx != 50 || sy != 100 {
		t.Errorf("expected screen coordinates (50, 100), got (%f, %f)", sx, sy)
	}

	// ScreenToWorld
	// Screen point (50, 100) relative to camera (200, 300) should be world point (250, 400)
	wx, wy := cam.ScreenToWorld(50, 100)
	if wx != 250 || wy != 400 {
		t.Errorf("expected world coordinates (250, 400), got (%f, %f)", wx, wy)
	}
}

func TestCameraUpdateKeyboard(t *testing.T) {
	cam := NewCamera(800, 600, 300.0, 10)
	cam.SetPosition(100, 100)

	// 1. No input should result in no movement
	mock := &mockInputProvider{mx: 400, my: 300} // middle of screen
	cam.Update(mock, 1.0)
	if cam.X != 100 || cam.Y != 100 {
		t.Errorf("expected no movement with no input, got (%f, %f)", cam.X, cam.Y)
	}

	// 2. Test left movement
	cam.SetPosition(100, 100)
	mock.left = true
	cam.Update(mock, 0.1) // 300 * 0.1 = 30px left
	if cam.X != 70 || cam.Y != 100 {
		t.Errorf("expected X to decrease to 70, got (%f, %f)", cam.X, cam.Y)
	}
	mock.left = false

	// 3. Test right movement
	cam.SetPosition(100, 100)
	mock.right = true
	cam.Update(mock, 0.2) // 300 * 0.2 = 60px right
	if cam.X != 160 || cam.Y != 100 {
		t.Errorf("expected X to increase to 160, got (%f, %f)", cam.X, cam.Y)
	}
	mock.right = false

	// 4. Test up movement
	cam.SetPosition(100, 100)
	mock.up = true
	cam.Update(mock, 0.1) // 300 * 0.1 = 30px up
	if cam.X != 100 || cam.Y != 70 {
		t.Errorf("expected Y to decrease to 70, got (%f, %f)", cam.X, cam.Y)
	}
	mock.up = false

	// 5. Test down movement
	cam.SetPosition(100, 100)
	mock.down = true
	cam.Update(mock, 0.2) // 300 * 0.2 = 60px down
	if cam.X != 100 || cam.Y != 160 {
		t.Errorf("expected Y to increase to 160, got (%f, %f)", cam.X, cam.Y)
	}
	mock.down = false

	// 6. Combined movement (up + right)
	cam.SetPosition(100, 100)
	mock.up = true
	mock.right = true
	cam.Update(mock, 0.1) // 30px up, 30px right
	if cam.X != 130 || cam.Y != 70 {
		t.Errorf("expected (130, 70), got (%f, %f)", cam.X, cam.Y)
	}
	mock.up = false
	mock.right = false
}

func TestCameraUpdateMouseEdgePanning(t *testing.T) {
	// Viewport 800x600, scroll speed 300 px/sec, margin 10 px
	cam := NewCamera(800, 600, 300.0, 10)

	// 1. Mouse at left edge (X <= 10)
	cam.SetPosition(100, 100)
	mock := &mockInputProvider{mx: 5, my: 300}
	cam.Update(mock, 0.1) // 300 * 0.1 = 30px left
	if cam.X != 70 || cam.Y != 100 {
		t.Errorf("expected X to pan left to 70, got (%f, %f)", cam.X, cam.Y)
	}

	// 2. Mouse at right edge (X >= 790)
	cam.SetPosition(100, 100)
	mock = &mockInputProvider{mx: 795, my: 300}
	cam.Update(mock, 0.1) // 300 * 0.1 = 30px right
	if cam.X != 130 || cam.Y != 100 {
		t.Errorf("expected X to pan right to 130, got (%f, %f)", cam.X, cam.Y)
	}

	// 3. Mouse at top edge (Y <= 10)
	cam.SetPosition(100, 100)
	mock = &mockInputProvider{mx: 400, my: 5}
	cam.Update(mock, 0.1) // 300 * 0.1 = 30px up
	if cam.X != 100 || cam.Y != 70 {
		t.Errorf("expected Y to pan up to 70, got (%f, %f)", cam.X, cam.Y)
	}

	// 4. Mouse at bottom edge (Y >= 590)
	cam.SetPosition(100, 100)
	mock = &mockInputProvider{mx: 400, my: 595}
	cam.Update(mock, 0.1) // 300 * 0.1 = 30px down
	if cam.X != 100 || cam.Y != 130 {
		t.Errorf("expected Y to pan down to 130, got (%f, %f)", cam.X, cam.Y)
	}

	// 5. Mouse outside viewport (e.g. negative or out of bounds) - should NOT trigger panning
	cam.SetPosition(100, 100)
	mock = &mockInputProvider{mx: -5, my: 300}
	cam.Update(mock, 0.1)
	if cam.X != 100 || cam.Y != 100 {
		t.Errorf("expected no panning with negative mouse coordinates, got (%f, %f)", cam.X, cam.Y)
	}

	cam.SetPosition(100, 100)
	mock = &mockInputProvider{mx: 805, my: 300}
	cam.Update(mock, 0.1)
	if cam.X != 100 || cam.Y != 100 {
		t.Errorf("expected no panning with out-of-bounds mouse coordinates, got (%f, %f)", cam.X, cam.Y)
	}
}

func TestGetVisibleTiles(t *testing.T) {
	// Viewport 800x600, tile size is 32.
	cam := NewCamera(800, 600, 300.0, 10)

	// 1. Camera at (0, 0)
	cam.SetPosition(0, 0)
	minCol, minRow, maxCol, maxRow := cam.GetVisibleTiles()
	// TileSize = 32. 
	// minCol = floor(0 / 32) = 0
	// maxCol = floor(800 / 32) = 25 (clamped to MapWidth - 1 is not exceeded since MapWidth is 256)
	// minRow = floor(0 / 32) = 0
	// maxRow = floor(600 / 32) = 18
	if minCol != 0 || minRow != 0 || maxCol != 25 || maxRow != 18 {
		t.Errorf("expected (0, 0, 25, 18), got (%d, %d, %d, %d)", minCol, minRow, maxCol, maxRow)
	}

	// 2. Camera at near-boundary to test clamping (Max camera positions: 7392, 7592)
	cam.SetPosition(7392, 7592)
	minCol, minRow, maxCol, maxRow = cam.GetVisibleTiles()
	// minCol = floor(7392 / 32) = 231
	// maxCol = floor((7392 + 800) / 32) = floor(8192 / 32) = 256, clamped to MapWidth - 1 (255)
	// minRow = floor(7592 / 32) = 237
	// maxRow = floor((7592 + 600) / 32) = floor(8192 / 32) = 256, clamped to MapHeight - 1 (255)
	if minCol != 231 || minRow != 237 || maxCol != 255 || maxRow != 255 {
		t.Errorf("expected (231, 237, 255, 255), got (%d, %d, %d, %d)", minCol, minRow, maxCol, maxRow)
	}
}
