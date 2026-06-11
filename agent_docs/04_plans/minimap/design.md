# Design Document: Interactive Minimap Overlay with Click-to-Pan

## User Story

### Headline
As a player, I want an interactive floating minimap overlay in the bottom-left corner of the screen with a viewport indicator and click-to-pan capability, so that I can easily navigate the 256x256 tilemap.

### Problem Statement
Scrolling a large 256x256 map (8192x8192 pixels) via arrow keys or mouse-edge panning can be slow and exhausting. To provide a true RTS feel and allow rapid navigation, we need a minimap. The minimap should always be exactly 256x256 pixels in screen coordinates and rendered in the bottom-left corner of the viewport. It must support multiple map sizes (64x64, 128x128, 256x256) by scaling tiles cleanly. To optimize performance, the background of the minimap should be cached on an image buffer rather than drawn pixel-by-pixel every frame. Finally, a rectangular indicator representing the current camera viewport must be drawn over the minimap, and clicking/dragging on the minimap must instantly center the camera viewport at that world position.

### Objective
- Create an interactive Minimap component inside the client.
- **Render**: Cache a 256x256 image representing the map where each tile is drawn as a colored dot corresponding to its `TileType` (standard grass vs variants).
- **Scale**: Clean scaling logic for supported sizes:
  - 256x256 map: 1 tile = 1x1 dot
  - 128x128 map: 1 tile = 2x2 dot
  - 64x64 map: 1 tile = 4x4 dot
- **Camera Indicator**: Overlay a thin, bright white rectangle showing the boundary of the camera viewport on the minimap.
- **Sleek Border**: Render a double-lined metallic border around the minimap.
- **Click-to-Pan**: Intercept mouse clicks and drags on the minimap to center the camera around that world location.

### Expected Outcome
When running the client/local mode, a 256x256 minimap with a gray border appears in the bottom-left of the 1024x768 window. The minimap displays grass tiles colored based on their type. A white box shows the currently visible area. Left-clicking or dragging inside the minimap immediately centers the main camera view on that part of the world.

---

## Architecture Overview

We will implement the minimap logic within `internal/client/minimap.go` and hook it into the main client game loop in `internal/client/game.go`.

### Minimap Coordinate Conversion
Given:
- Screen dimensions `W` and `H`.
- Minimap screen dimensions `MinimapSize = 256`.
- Minimap screen coordinates (bottom-left):
  - `mx = 16`
  - `my = H - MinimapSize - 16`
- World map pixel dimensions: `MapWidthPx` and `MapHeightPx`.

Formulas:
1. **World to Minimap (Relative)**:
   - `rx = (wx / MapWidthPx) * MinimapSize`
   - `ry = (wy / MapHeightPx) * MinimapSize`
2. **Minimap to World**:
   - `wx = (rx / MinimapSize) * MapWidthPx`
   - `wy = (ry / MinimapSize) * MapHeightPx`
3. **Viewport Indicator Size & Position**:
   - Indicator X: `vx = mx + (c.X / MapWidthPx) * MinimapSize`
   - Indicator Y: `vy = my + (c.Y / MapHeightPx) * MinimapSize`
   - Indicator Width: `vw = (ViewportWidth / MapWidthPx) * MinimapSize`
   - Indicator Height: `vh = (ViewportHeight / MapHeightPx) * MinimapSize`

### Caching Strategy
- Maintain an `*ebiten.Image` of size 256x256 as a cache.
- Generate this cache once when the map is received/loaded.
- Re-generate only if the map layout or tile types change.
- Drawing this cached image is an $O(1)$ draw call per frame rather than $O(N)$ pixel setting.

### Input Handling
In the game client's `Update()` loop:
1. Get the mouse position `(mouse_x, mouse_y)`.
2. Check if the mouse is pressed (`ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)`).
3. Check if `mouse_x` is between `mx` and `mx + 256`, and `mouse_y` is between `my` and `my + 256`.
4. If yes, convert the relative click coordinates back to world pixels, center the camera around this coordinate, and call `c.ClampPosition()`.

---

## Implementation Backlog

### ## Pending
- **Task 1: Implement Minimap Model and Coordinate Translation**
  - Create `internal/client/minimap.go` with `Minimap` struct.
  - Implement methods to calculate minimap screen boundaries, translate world coordinates to minimap relative coordinates, and translate minimap click coordinates back to world coordinates.
  - Write unit tests for all coordinate translations and map size scale factors.
- **Task 2: Implement Minimap Background Caching & Custom Drawing**
  - Implement dynamic scaling for 64, 128, and 256 map dimensions.
  - Draw tile-type color dots on the cache image (`TileTypeGrassStandard`, `TileTypeGrassVariant1`, `TileTypeGrassVariant2` colored as distinct shades of green).
  - Implement pre-rendering logic to draw into a 256x256 Ebitengine image cache.
  - Write unit tests to verify caching behavior and rendering safety.
- **Task 3: Implement Camera Indicator & Metallic Border Drawing**
  - Add logic to draw the camera's viewport rectangle on top of the minimap screen coordinates.
  - Draw a 2-pixel wide double metallic border around the minimap coordinate bounds.
  - Implement unit/mock tests.
- **Task 4: Implement Minimap Input Interactivity (Click-to-Pan & Drag-to-Pan)**
  - Add input interception to detect mouse click and drag gestures within the minimap.
  - Centering camera world coordinates at clicked location and clamping.
  - Unit/mock tests.
- **Task 5: Integrate Minimap into Game Client Loop**
  - Update `internal/client/game.go` to instantiate, pre-render, update, and draw the minimap.
  - Perform manual end-to-end verification of click-to-pan, dragging, scaling, and border overlays.

### ## Current
*(None yet. Waiting for initial approval)*

### ## Completed
*(None)*

---

## Checklist & TDD Requirements

1. **Strict Test-Driven Development**:
   - Write unit tests before implementing. Prove tests fail first, then write the minimal implementation to pass.
2. **Decoupled Math**:
   - All coordinate translation formulas must be isolated and unit-tested without requiring active Ebitengine graphic loops or window hooks.
3. **No Redundant Draws**:
   - Ensure the tile dots are ONLY drawn once into the cache when the map is first initialized, avoiding major frame-rate drops.
4. **Boundary Clamping**:
   - The click-to-pan feature must never slide the camera viewport out of the map boundaries, and must use the existing `ClampPosition()` logic in `Camera`.
