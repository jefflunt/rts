# Implement a client-side camera system in Go using Ebitengine that handles world-to-screen and screen-to-world coordinate mapping, input-driven movement (supporting arrow keys and mouse-edge panning), and viewport bounding-box culling calculations to render only visible tiles from the 256x256 grass map.

This task focuses on implementing a robust, self-contained camera system for the Ebitengine-based client. The camera is responsible for three primary functions:

1. Coordinate Conversion: Mapping coordinate spaces between the world coordinates (the 256x256 grass tilemap space) and screen coordinates (the viewport space displayed to the user). This includes converting mouse-click screen coordinates back to map/world coordinates and projecting world entities onto the screen viewport.

2. Movement Logic: Implementing camera panning controls that allow the user to scroll through the 256x256 map. The movement must support two modes of input: keyboard arrow keys and mouse-edge panning (moving the camera when the cursor is near the edge of the window).

3. Viewport Bounding-Box Culling: Calculating the bounding box of the current viewport relative to the tilemap. Based on the screen size and the current camera coordinates, the system must calculate the minimum and maximum row/column indices of tiles that are currently visible. The renderer will use these calculated bounds to cull off-screen tiles, ensuring high rendering performance by only drawing visible grass tiles.
