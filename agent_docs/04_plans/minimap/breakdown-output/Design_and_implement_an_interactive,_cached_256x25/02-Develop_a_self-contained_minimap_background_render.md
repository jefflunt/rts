# Develop a self-contained minimap background renderer that scales and pre-renders map tiles into a static 256x256 pixel offscreen ebiten.Image cache, supporting 64x64, 128x128, and 256x256 map dimensions with color-coded tile-type representations.

This task is responsible for implementing the background cache of the client minimap component. To optimize rendering performance, the game map's tiles are drawn onto an offscreen 256x256 pixel `ebiten.Image` buffer once when the map is first loaded or modified, rather than drawing thousands of tile dots individually every frame. This offscreen cache is then drawn onto the viewport as a single pre-rendered texture, minimizing the performance footprint.

The caching logic must dynamically scale the rendering depending on the dimension of the map. For a 256x256 map, each tile maps to a 1x1 pixel dot; for a 128x128 map, each tile scales to a 2x2 pixel dot; and for a 64x64 map, each tile scales to a 4x4 pixel dot. The color of each dot is dynamically mapped using predefined, distinct shades of green based on the tile's `TileType` (standard grass versus its different variant layouts).

This task forms a distinct and testable slice of the minimap feature. The scaling factors, image buffer creation, and mapping logic can be unit tested without requiring a running game window, making it fully actionable as a single logical unit of work.
