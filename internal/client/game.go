package client

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	tilemap "scrollable-tilemap/internal/map"
)

// Game implements the ebiten.Game interface for the client.
type Game struct {
	camera     *Camera
	input      InputProvider
	tileMap    *tilemap.Map
	tileImages map[tilemap.TileType]map[uint8][2]*ebiten.Image // [Type][Variant][isEven (0=false, 1=true)]
	minimap    *Minimap
}

// NewGame creates and returns a new Game instance.
func NewGame(camera *Camera, input InputProvider, m *tilemap.Map) *Game {
	g := &Game{
		camera:     camera,
		input:      input,
		tileMap:    m,
		tileImages: make(map[tilemap.TileType]map[uint8][2]*ebiten.Image),
	}
	if m != nil {
		mapWidthTiles := tilemap.MapWidth
		mapHeightTiles := tilemap.MapHeight
		tileSize := tilemap.TileSize
		if camera != nil {
			mapWidthTiles = camera.MapWidthPx / tilemap.TileSize
			mapHeightTiles = camera.MapHeightPx / tilemap.TileSize
		}
		g.minimap = NewMinimap(mapWidthTiles, mapHeightTiles, tileSize)
		g.minimap.UpdateCache(m)
	}
	g.pregenerateTiles()
	return g
}

// Update updates the game state. Ebitengine calls Update 60 times per second.
func (g *Game) Update() error {
	if g.camera != nil {
		g.camera.Update(g.input, 1.0/60.0)
	}
	if g.minimap != nil {
		g.minimap.Update(g.camera, g.input)
	}
	return nil
}

// Draw draws the game screen. Ebitengine calls Draw every frame.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.camera == nil || g.tileMap == nil {
		// If map or camera is nil, draw a dark green background.
		screen.Fill(color.RGBA{R: 20, G: 80, B: 20, A: 255})
		return
	}

	// 1. Get visible tiles (viewport culling bounding box)
	minCol, minRow, maxCol, maxRow := g.camera.GetVisibleTiles()

	// 2. Iterate and render only visible tiles
	for col := minCol; col <= maxCol; col++ {
		for row := minRow; row <= maxRow; row++ {
			tile, err := g.tileMap.GetTile(col, row)
			if err != nil {
				continue
			}

			// Get top-left of the tile in world coordinates
			wx, wy := tilemap.TileToWorld(col, row)

			// Convert world coordinates to screen/viewport coordinates
			sx, sy := g.camera.WorldToScreen(wx, wy)

			// Alternate colors for a clean checkered pattern
			isEven := (col+row)%2 == 0

			tileImg := g.getTileImage(tile.Type, tile.Variant, isEven)
			if tileImg != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(sx, sy)
				screen.DrawImage(tileImg, op)
			}
		}
	}

	// 3. Draw minimap overlay if initialized
	if g.minimap != nil {
		g.minimap.Draw(screen, g.camera)
	}
}

// Layout takes the outside size (e.g., window size) and returns the game's logical screen size.
// It updates the camera's viewport dimensions to handle window resizing dynamically.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if g.camera != nil {
		g.camera.ViewportWidth = outsideWidth
		g.camera.ViewportHeight = outsideHeight
		g.camera.ClampPosition()
		return outsideWidth, outsideHeight
	}
	return outsideWidth, outsideHeight
}

// SetMap updates the client's local tilemap.
func (g *Game) SetMap(m *tilemap.Map) {
	g.tileMap = m
	if m != nil {
		if g.minimap == nil {
			mapWidthTiles := tilemap.MapWidth
			mapHeightTiles := tilemap.MapHeight
			tileSize := tilemap.TileSize
			if g.camera != nil {
				mapWidthTiles = g.camera.MapWidthPx / tilemap.TileSize
				mapHeightTiles = g.camera.MapHeightPx / tilemap.TileSize
			}
			g.minimap = NewMinimap(mapWidthTiles, mapHeightTiles, tileSize)
		}
		g.minimap.UpdateCache(m)
	}
}

// SetMapDimensions updates the camera's and minimap's map dimensions and regenerates cache.
func (g *Game) SetMapDimensions(widthTiles, heightTiles, tileSize int) {
	if g.camera != nil {
		g.camera.SetMapDimensions(widthTiles, heightTiles, tileSize)
	}
	if g.minimap == nil {
		g.minimap = NewMinimap(widthTiles, heightTiles, tileSize)
	} else {
		g.minimap.MapWidthTiles = widthTiles
		g.minimap.MapHeightTiles = heightTiles
		g.minimap.TileSize = tileSize
	}
	if g.tileMap != nil {
		g.minimap.UpdateCache(g.tileMap)
	}
}

// GetMap returns the client's current tilemap.
func (g *Game) GetMap() *tilemap.Map {
	return g.tileMap
}

// GetMinimap returns the game's Minimap instance.
func (g *Game) GetMinimap() *Minimap {
	return g.minimap
}

// GetCamera returns the camera instance used by the game loop.
func (g *Game) GetCamera() *Camera {
	return g.camera
}

// GetInput returns the input provider instance used by the game loop.
func (g *Game) GetInput() InputProvider {
	return g.input
}

// getTileImage retrieves a cached tile image or generates it if not already cached.
func (g *Game) getTileImage(tileType tilemap.TileType, variant uint8, isEven bool) *ebiten.Image {
	if g.tileImages == nil {
		g.tileImages = make(map[tilemap.TileType]map[uint8][2]*ebiten.Image)
	}

	variants, exists := g.tileImages[tileType]
	if !exists {
		variants = make(map[uint8][2]*ebiten.Image)
		g.tileImages[tileType] = variants
	}

	evenIdx := 0
	if isEven {
		evenIdx = 1
	}

	imgs, exists := variants[variant]
	if !exists {
		img := g.createProceduralTile(tileType, variant, isEven)
		imgs = [2]*ebiten.Image{}
		imgs[evenIdx] = img
		variants[variant] = imgs
		return img
	}

	img := imgs[evenIdx]
	if img == nil {
		img = g.createProceduralTile(tileType, variant, isEven)
		imgs[evenIdx] = img
		variants[variant] = imgs
	}
	return img
}

// createProceduralTile creates a deterministic procedural 32x32 grass tile.
func (g *Game) createProceduralTile(tileType tilemap.TileType, variant uint8, isEven bool) *ebiten.Image {
	img := ebiten.NewImage(tilemap.TileSize, tilemap.TileSize)

	// 1. Determine base checkered colors
	var baseColor color.RGBA
	if isEven {
		// Lighter grass green
		baseColor = color.RGBA{R: 34, G: 160, B: 34, A: 255}
	} else {
		// Darker grass green
		baseColor = color.RGBA{R: 28, G: 130, B: 28, A: 255}
	}
	img.Fill(baseColor)

	// 2. Add deterministic procedural grass speckles/lines based on Type and Variant.
	// Use an LCG pseudo-random sequence seeded with the type and variant.
	seed := int(tileType)*100 + int(variant)*10
	rng := seed
	nextRand := func() int {
		rng = (rng*1103515245 + 12345) & 0x7fffffff
		return rng
	}

	// Brighter grass green or darker contrast speckles
	var speckleColor color.RGBA
	if isEven {
		speckleColor = color.RGBA{R: 50, G: 205, B: 50, A: 255} // Lime green speckles on light green
	} else {
		speckleColor = color.RGBA{R: 34, G: 139, B: 34, A: 255} // Forest green speckles on dark green
	}

	switch tileType {
	case tilemap.TileTypeGrassStandard:
		// Vertical/slanted blades of grass
		numBlades := 3 + (variant % 3)
		for i := uint8(0); i < numBlades; i++ {
			bx := nextRand() % (tilemap.TileSize - 4)
			by := nextRand() % (tilemap.TileSize - 6)
			// Draw 1x3 vertical blade of grass
			img.Set(bx, by, speckleColor)
			img.Set(bx+1, by-1, speckleColor)
			img.Set(bx+1, by-2, speckleColor)
		}
	case tilemap.TileTypeGrassVariant1:
		// Horizontal speckles
		numSpeckles := 5 + (variant % 4)
		for i := uint8(0); i < numSpeckles; i++ {
			sx := nextRand() % (tilemap.TileSize - 2)
			sy := nextRand() % (tilemap.TileSize - 2)
			img.Set(sx, sy, speckleColor)
			img.Set(sx+1, sy, speckleColor)
		}
	case tilemap.TileTypeGrassVariant2:
		// Small grass cluster/crosses
		numClusters := 2 + (variant % 2)
		for i := uint8(0); i < numClusters; i++ {
			cx := nextRand() % (tilemap.TileSize - 4)
			cy := nextRand() % (tilemap.TileSize - 4)
			img.Set(cx+1, cy, speckleColor)
			img.Set(cx, cy+1, speckleColor)
			img.Set(cx+1, cy+1, speckleColor)
			img.Set(cx+2, cy+1, speckleColor)
			img.Set(cx+1, cy+2, speckleColor)
		}
	}

	// 3. Very subtle border to highlight individual grid/coordinate boundaries
	borderColor := color.RGBA{R: 20, G: 100, B: 20, A: 255}
	for i := 0; i < tilemap.TileSize; i++ {
		img.Set(i, 0, borderColor)
		img.Set(i, tilemap.TileSize-1, borderColor)
		img.Set(0, i, borderColor)
		img.Set(tilemap.TileSize-1, i, borderColor)
	}

	return img
}

// pregenerateTiles pre-populates cache with common grass tiles to avoid runtime stutter.
func (g *Game) pregenerateTiles() {
	if g.tileImages == nil {
		g.tileImages = make(map[tilemap.TileType]map[uint8][2]*ebiten.Image)
	}

	types := []tilemap.TileType{
		tilemap.TileTypeGrassStandard,
		tilemap.TileTypeGrassVariant1,
		tilemap.TileTypeGrassVariant2,
	}

	for _, t := range types {
		for v := uint8(0); v < 8; v++ {
			_ = g.getTileImage(t, v, true)
			_ = g.getTileImage(t, v, false)
		}
	}
}
