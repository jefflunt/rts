# Integrate the pre-rendered and interactive Minimap component into the main client game loop in internal/client/game.go by modifying the Game struct to house the Minimap, instantiating it with the map and camera, handling user interaction during the update phase, and drawing the background cache, metallic border, and viewport indicator in the viewport's bottom-left corner.

This task covers the integration of the interactive Minimap component into the client's core game loop in 'internal/client/game.go'. It involves updating the 'Game' structure to house the Minimap instance, hooking up its lifecycle methods, and ensuring the minimap renders correctly in the bottom-left viewport position.

During initialization in 'NewGame', a new Minimap instance is instantiated and pre-renders its colored tile dot cache based on the current map structure. If 'SetMap' is called, the Minimap's map reference is updated and the cache is regenerated to reflect map changes.

In the 'Update' method, the Minimap's update logic is executed to process input clicks and drags within the bottom-left 256x256 overlay bounds. This coordinates with the camera to instantly pan and clamp to the corresponding map positions. Finally, inside 'Draw', the Minimap is drawn on top of the map layer to render the background cache, the double-lined metallic border, and the white camera viewport boundary indicator.
