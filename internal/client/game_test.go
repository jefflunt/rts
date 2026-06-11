package client

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	tilemap "scrollable-tilemap/internal/map"
)

func TestNewGame(t *testing.T) {
	cam := NewCamera(800, 600, 300.0, 10)
	input := &mockInputProvider{mx: 400, my: 300}
	m := tilemap.NewDefaultMap()

	g := NewGame(cam, input, m)

	if g.GetCamera() != cam {
		t.Errorf("expected camera to be set correctly")
	}
	if g.GetInput() != input {
		t.Errorf("expected input to be set correctly")
	}
	if g.GetMap() != m {
		t.Errorf("expected map to be set correctly")
	}

	// Verify pregeneration populated some tile images
	if len(g.tileImages) == 0 {
		t.Errorf("expected tileImages cache to be populated during pregeneration")
	}

	// Let's verify standard grass tile exists in cache
	variants, ok := g.tileImages[tilemap.TileTypeGrassStandard]
	if !ok {
		t.Errorf("expected TileTypeGrassStandard to be cached")
	}
	if len(variants) == 0 {
		t.Errorf("expected variants to be populated")
	}
}

func TestGameUpdate(t *testing.T) {
	cam := NewCamera(800, 600, 300.0, 10)
	cam.SetPosition(100, 100)
	// Put mouse in the center so edge panning is not triggered
	input := &mockInputProvider{right: true, mx: 400, my: 300}
	m := tilemap.NewDefaultMap()

	g := NewGame(cam, input, m)

	// Since Game.Update calls g.camera.Update(g.input, 1.0/60.0),
	// camera should move to the right by 300 * (1/60) = 5 pixels.
	err := g.Update()
	if err != nil {
		t.Fatalf("unexpected error from Update: %v", err)
	}

	if cam.X != 105 {
		t.Errorf("expected camera to move to X=105, got X=%f", cam.X)
	}

	// Test Update with nil camera
	gNilCam := NewGame(nil, input, m)
	err = gNilCam.Update()
	if err != nil {
		t.Errorf("unexpected error from Update with nil camera: %v", err)
	}
}

func TestGameLayout(t *testing.T) {
	cam := NewCamera(800, 600, 300.0, 10)
	input := &mockInputProvider{mx: 400, my: 300}
	m := tilemap.NewDefaultMap()

	g := NewGame(cam, input, m)

	w, h := g.Layout(1024, 768)
	if w != 1024 || h != 768 {
		t.Errorf("expected logical screen size 1024x768, got %dx%d", w, h)
	}

	if cam.ViewportWidth != 1024 || cam.ViewportHeight != 768 {
		t.Errorf("expected camera viewport size to update to 1024x768, got %dx%d", cam.ViewportWidth, cam.ViewportHeight)
	}

	// Test Layout with nil camera
	gNilCam := NewGame(nil, input, m)
	w, h = gNilCam.Layout(1024, 768)
	if w != 1024 || h != 768 {
		t.Errorf("expected logical screen size 1024x768 with nil camera, got %dx%d", w, h)
	}
}

func TestGameSetMap(t *testing.T) {
	cam := NewCamera(800, 600, 300.0, 10)
	input := &mockInputProvider{mx: 400, my: 300}
	m1 := tilemap.NewDefaultMap()
	m2 := tilemap.NewDefaultMap()

	// Modify m2 slightly to distinguish it
	_ = m2.SetTile(0, 0, tilemap.Tile{Type: tilemap.TileTypeGrassVariant2, Variant: 5})

	g := NewGame(cam, input, m1)
	if g.GetMap() != m1 {
		t.Errorf("expected map to be m1 initially")
	}

	g.SetMap(m2)
	if g.GetMap() != m2 {
		t.Errorf("expected map to be updated to m2")
	}

	tile, err := g.GetMap().GetTile(0, 0)
	if err != nil || tile.Type != tilemap.TileTypeGrassVariant2 || tile.Variant != 5 {
		t.Errorf("expected updated map's tile values, got type=%v, variant=%d", tile.Type, tile.Variant)
	}
}

func TestGameDrawNil(t *testing.T) {
	// Create a game with nil camera or map
	g := &Game{
		camera:     nil,
		input:      nil,
		tileMap:    nil,
		tileImages: nil,
	}

	screen := ebiten.NewImage(100, 100)

	// In nil case, Draw should run fine and not panic.
	// Since ReadPixels/At is not supported in headless mode prior to RunGame,
	// we just verify that calling Draw does not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Draw panicked: %v", r)
		}
	}()

	g.Draw(screen)
}

func TestGameDrawValid(t *testing.T) {
	cam := NewCamera(100, 100, 300.0, 10)
	input := &mockInputProvider{mx: 50, my: 50}
	m := tilemap.NewDefaultMap()

	// Ensure (0,0) and (1,1) have distinct values
	_ = m.SetTile(0, 0, tilemap.Tile{Type: tilemap.TileTypeGrassStandard, Variant: 1})
	_ = m.SetTile(1, 1, tilemap.Tile{Type: tilemap.TileTypeGrassVariant1, Variant: 2})

	g := NewGame(cam, input, m)

	screen := ebiten.NewImage(100, 100)

	// Verify Draw executes without any panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Draw panicked: %v", r)
		}
	}()

	g.Draw(screen)

	// Move camera and draw again
	cam.SetPosition(32, 32)
	g.Draw(screen)
}

func TestPregenerateTilesWithVaryingCache(t *testing.T) {
	cam := NewCamera(100, 100, 300.0, 10)
	input := &mockInputProvider{mx: 50, my: 50}
	m := tilemap.NewDefaultMap()

	g := &Game{
		camera:     cam,
		input:      input,
		tileMap:    m,
		tileImages: nil, // force nil cache to test lazy initialization
	}

	// This should lazy initialize tileImages
	img := g.getTileImage(tilemap.TileTypeGrassStandard, 1, true)
	if img == nil {
		t.Fatalf("expected generated tile image to be non-nil")
	}

	if len(g.tileImages) == 0 {
		t.Errorf("expected tileImages map to be initialized")
	}

	// Retrieve again, should hit cache
	img2 := g.getTileImage(tilemap.TileTypeGrassStandard, 1, true)
	if img2 != img {
		t.Errorf("expected same image pointer from cache")
	}

	// Retrieve opposite even/odd variant, should generate new image
	img3 := g.getTileImage(tilemap.TileTypeGrassStandard, 1, false)
	if img3 == nil {
		t.Fatalf("expected generated tile image to be non-nil")
	}
	if img3 == img {
		t.Errorf("expected different image pointer for isEven=false")
	}
}

func TestGameMinimapIntegration(t *testing.T) {
	cam := NewCamera(800, 600, 300.0, 10)
	input := &mockInputProvider{mx: 400, my: 300}
	m := tilemap.NewDefaultMap()

	g := NewGame(cam, input, m)

	// Verify minimap was created and initialized
	minimap := g.GetMinimap()
	if minimap == nil {
		t.Fatalf("expected minimap to be non-nil after NewGame with a map")
	}

	if minimap.MapWidthTiles != tilemap.MapWidth || minimap.MapHeightTiles != tilemap.MapHeight {
		t.Errorf("expected minimap dimensions to match tilemap dimensions")
	}

	// Test SetMap with a new map
	m2 := tilemap.NewDefaultMap()
	g.SetMap(m2)
	if g.GetMinimap() == nil {
		t.Errorf("expected minimap to remain non-nil after SetMap")
	}

	// Verify minimap is initialized if NewGame was called with nil map
	gNilMap := NewGame(cam, input, nil)
	if gNilMap.GetMinimap() != nil {
		t.Errorf("expected minimap to be nil initially when game map is nil")
	}
	gNilMap.SetMap(m)
	if gNilMap.GetMinimap() == nil {
		t.Fatalf("expected minimap to be created when SetMap is called with a non-nil map")
	}

	// Verify Game.Draw runs with minimap without panic
	screen := ebiten.NewImage(800, 600)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Game.Draw panicked with minimap: %v", r)
		}
	}()
	g.Draw(screen)
}

func TestGameMinimapUpdateInteraction(t *testing.T) {
	cam := NewCamera(800, 600, 300.0, 10)
	cam.SetPosition(1000, 1000)

	// In 800x600 screen, the minimap overlay Y coordinate starts at:
	// 600 - 16 - 256 = 328
	// X coordinate starts at: 16
	// The center of the minimap is at:
	// X = 16 + 128 = 144
	// Y = 328 + 128 = 456
	input := &mockInputProvider{
		click: true,
		mx:    144,
		my:    456,
	}
	m := tilemap.NewDefaultMap()

	g := NewGame(cam, input, m)

	// Call Update, which should trigger the minimap interaction,
	// dragging the camera to center on the world position matching the minimap center.
	// Map size: 256 * 32 = 8192.
	// Minimap center: (128, 128).
	// World position: (4096, 4096).
	// Target camera position (centering on 4096, 4096 with 800x600 viewport):
	// targetX = 4096 - 400 = 3696
	// targetY = 4096 - 300 = 3796
	err := g.Update()
	if err != nil {
		t.Fatalf("unexpected error from Update: %v", err)
	}

	if cam.X != 3696 || cam.Y != 3796 {
		t.Errorf("expected camera position to pan to (3696, 3796) via minimap interaction, got (%f, %f)", cam.X, cam.Y)
	}

	if !g.GetMinimap().IsDragging() {
		t.Error("expected minimap to be in dragging state")
	}

	// Release mouse and verify it is no longer dragging on next Update
	input.click = false
	err = g.Update()
	if err != nil {
		t.Fatalf("unexpected error from Update: %v", err)
	}

	if g.GetMinimap().IsDragging() {
		t.Error("expected minimap to stop dragging after click release")
	}
}
