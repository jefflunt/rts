# Domain Logic & Coordinate Systems

This document describes the core domain concepts, data structures, and mathematical formulas used across the codebase.

---

## 1. Core Domain Entities

### The Tile Grid
The map is a uniform 2D grid of size $256 \times 256$ tiles.
- **Tile**: A discrete 32x32 pixel square.
- **Tile Types**:
  - `TileTypeGrassStandard` (Value `0`): Standard grass.
  - `TileTypeGrassVariant1` (Value `1`): Grass variation 1 (for visual checkerboarding).
  - `TileTypeGrassVariant2` (Value `2`): Grass variation 2.

### Coordinate Systems

| Coordinate System | Units | Range (for 256x256 map) | Description |
|---|---|---|---|
| **Grid (Tile) Coordinates** | Tiles | $x, y \in [0, 255]$ | Row/column indices in the map grid array. |
| **World Coordinates** | Pixels | $wx, wy \in [0, 8191]$ | Absolute pixel positions within the full world. |
| **Screen Coordinates** | Pixels | $sx, sy \in \text{Viewport}$ | Relative pixel positions inside the desktop window ($1024 \times 768$). |

---

## 2. Mathematical Conversions

### Grid Coordinates $\leftrightarrow$ World Coordinates
Given tile size $T = 32$ pixels:

$$\text{WorldX} = \text{GridX} \times T$$

$$\text{WorldY} = \text{GridY} \times T$$

$$\text{GridX} = \lfloor \text{WorldX} / T \rfloor$$

$$\text{GridY} = \lfloor \text{WorldY} / T \rfloor$$

### World Coordinates $\leftrightarrow$ Screen Coordinates
Given camera coordinate $(C_x, C_y)$ representing the top-left corner of the camera viewport in the world:

$$\text{ScreenX} = \text{WorldX} - C_x$$

$$\text{ScreenY} = \text{WorldY} - C_y$$

$$\text{WorldX} = \text{ScreenX} + C_x$$

$$\text{WorldY} = \text{ScreenY} + C_y$$

### World Coordinates $\leftrightarrow$ Minimap Coordinates
The minimap is drawn in a dedicated HUD rect on the screen:
- Minimap size is $M = 256$ pixels.
- Bottom-left margin: $X_{\text{offset}} = 16$ px, $Y_{\text{offset}} = \text{ScreenHeight} - M - 16$ px.
- World pixel dimensions: $W_{\text{world}} = \text{MapWidth} \times T$, $H_{\text{world}} = \text{MapHeight} \times T$.

To translate an absolute world position $(W_x, W_y)$ to minimap relative coordinate $(R_x, R_y) \in [0, 256]$:

$$R_x = \left( \frac{W_x}{W_{\text{world}}} \right) \times M$$

$$R_y = \left( \frac{W_y}{H_{\text{world}}} \right) \times M$$

To translate a relative minimap click $(R_x, R_y)$ back to world coordinates $(W_x, W_y)$ to center the camera:

$$W_x = \left( \frac{R_x}{M} \right) \times W_{\text{world}}$$

$$W_y = \left( \frac{R_y}{M} \right) \times H_{\text{world}}$$

---

## 3. Performance Optimizations

### Camera Viewport Culling
Rendering all 65,536 tiles per frame would be extremely inefficient. The camera determines the visible bounding box in world space:
- Camera bounds: $[C_x, C_x + V_{\text{width}}]$ and $[C_y, C_y + V_{\text{height}}]$.
- The index range of visible tiles is computed as:

$$\text{startGridX} = \max\left(0, \lfloor C_x / T \rfloor\right)$$

$$\text{endGridX} = \min\left(\text{MapWidth} - 1, \lfloor (C_x + V_{\text{width}}) / T \rfloor\right)$$

$$\text{startGridY} = \max\left(0, \lfloor C_y / T \rfloor\right)$$

$$\text{endGridY} = \min\left(\text{MapHeight} - 1, \lfloor (C_y + V_{\text{height}}) / T \rfloor\right)$$

The drawing routine only loops through $\text{GridX} \in [\text{startGridX}, \text{endGridX}]$ and $\text{GridY} \in [\text{startGridY}, \text{endGridY}]$, which reduces the draw calls from 65,536 to just $\approx 800$ per frame!

### Minimap Background Caching
Instead of drawing the minimap pixel-by-pixel or tile-by-tile on every single frame, we instantiate a single $256 \times 256$ `*ebiten.Image` cache.
- The cache is fully pre-rendered once when the client connects and receives the map data.
- During the frame drawing phase, the cached image is drawn onto the screen in a single $O(1)$ draw call.
- The camera's dynamic viewport indicator and the border are layered on top of this cached background on every frame.
