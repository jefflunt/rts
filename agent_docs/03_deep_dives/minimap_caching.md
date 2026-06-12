# Deep Dive: Minimap Caching & Interactive Navigation

## Overview
The Minimap subsystem renders a compact $256 \times 256$ pixel graphical representation of the massive $256 \times 256$ tilemap inside the game's HUD. It provides real-time position tracking of the main camera viewport and supports clicking and dragging to instantly reposition the view.

---

## Purpose
In real-time strategy (RTS) games, maps are too large to navigate solely via screen edge scrolling or keyboard arrows. The minimap serves as an essential tactical tool for instant world navigation and strategic spatial awareness.

---

## Components

- **`Minimap` Struct (`internal/client/minimap.go`)**:
  - Holds layout settings (border size, padding, width, height, screen coordinates).
  - Maintains `*ebiten.Image` cache pointers for background textures.
  - Converts world coordinates to minimap coordinates, and click coordinates back to world coordinates.
- **Pre-Renderer (`PreRenderBackground`)**:
  - Executes once upon connection when map data is received.
  - Iterates through the full map grid, resolves the `TileType` color, and sets pixels on the $256 \times 256$ image buffer.
- **Input Interceptor (`Update()`)**:
  - Captures left-mouse clicks or click-drags occurring inside the minimap coordinate boundaries on screen.
  - Instantly computes target world coordinates and adjusts the Camera position.
- **Border and Indicator Renderer (`Draw()`)**:
  - Draws the double-lined metallic border around the minimap bounds.
  - Calculates the camera's bounding box relative to the minimap scale and overlays a dynamic, high-contrast white rectangle.

---

## Data Flow

```
1. Handshake Complete ────> Store Map Data ────> Pre-render 256x256 Dot Texture (Cache)
                                                       │
                                                       ▼
2. Every Frame: Update() ──> Mouse Pressed? ────> Translate (Screen -> Minimap -> World) ────> Set Camera X/Y
                                                       │
                                                       ▼
3. Every Frame: Draw() ───> Draw Pre-rendered Cache ──> Calculate Viewport Indicator Box ──> Draw White Rect HUD
```

---

## Concerns & Constraints

- **Draw Call Overhead**: Setting individual pixels or drawing 65,536 dots per frame inside the game loop causes severe performance degradation. The $O(1)$ pre-rendered image cache completely eliminates this, reducing drawing complexity to a single Ebitengine image draw call.
- **Clamping**: Mouse coordinates can easily drag outside the minimap bounds if the drag gesture is fast. The panning logic must clamp the calculated target world coordinates strictly between `0` and `MapWidth * TileSize`, ensuring the camera never glides into black space.
