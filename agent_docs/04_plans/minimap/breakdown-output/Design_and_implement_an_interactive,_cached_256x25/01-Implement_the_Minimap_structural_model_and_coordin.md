# Implement the Minimap structural model and coordinate translation functions within internal/client/minimap.go, including world-to-minimap, minimap-to-world, and viewport indicator size and position calculations.

This task focuses on implementing the `Minimap` data structure and coordinate translation functions in `internal/client/minimap.go`. The minimap occupies a fixed 256x256 area in the bottom-left corner of the client screen, offset by 16 pixels from the edges.

The implementation must define the struct to hold configuration parameters such as map dimensions (supporting 64x64, 128x128, and 256x256 tilemaps), and implement the mathematical translation functions:
1. World-to-Minimap: Translates a world coordinate in pixels into its corresponding position relative to the 256x256 minimap boundary.
2. Minimap-to-World: Converts a pixel coordinate clicked on the 256x256 minimap back into absolute world coordinates in pixels, allowing for camera centering.
3. Viewport Indicator bounds: Computes the screen-space X, Y coordinates, width, and height of the white rectangle representing the camera's current visible viewport on top of the minimap screen area.
