package client

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// InputProvider defines the interface for polling user inputs required by the Camera.
// This interface allows the camera's update logic to be thoroughly unit-tested
// without depending directly on Ebitengine's actual windowing and hardware systems.
type InputProvider interface {
	IsArrowUpPressed() bool
	IsArrowDownPressed() bool
	IsArrowLeftPressed() bool
	IsArrowRightPressed() bool
	CursorPosition() (int, int)
}

// EbitenInputProvider is a concrete implementation of InputProvider
// that queries the actual Ebitengine input state.
type EbitenInputProvider struct{}

// NewEbitenInputProvider creates a new EbitenInputProvider.
func NewEbitenInputProvider() *EbitenInputProvider {
	return &EbitenInputProvider{}
}

// IsArrowUpPressed returns true if the up arrow key is currently pressed.
func (e *EbitenInputProvider) IsArrowUpPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyArrowUp)
}

// IsArrowDownPressed returns true if the down arrow key is currently pressed.
func (e *EbitenInputProvider) IsArrowDownPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyArrowDown)
}

// IsArrowLeftPressed returns true if the left arrow key is currently pressed.
func (e *EbitenInputProvider) IsArrowLeftPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
}

// IsArrowRightPressed returns true if the right arrow key is currently pressed.
func (e *EbitenInputProvider) IsArrowRightPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyArrowRight)
}

// CursorPosition returns the current mouse cursor coordinates.
func (e *EbitenInputProvider) CursorPosition() (int, int) {
	return ebiten.CursorPosition()
}
