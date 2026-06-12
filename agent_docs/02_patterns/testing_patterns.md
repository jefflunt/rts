# Testing Patterns

This codebase was built with **Strict Test-Driven Development (TDD)**. Tests are first-class citizens and must cover all coordinate boundaries, network interactions, and UI-state mutations.

---

## 1. Core Testing Rules

1. **No Implementation Without Failing Tests**: If you need to add a new function or field, write the test first. Prove it fails (RED), implement the feature (GREEN), and refactor.
2. **Deterministic & Fast**: Tests must not rely on real network timing, Sleep calls, or external files. Everything must be mocked or executed using standard in-memory constructs.
3. **Mocking External Interactions**:
   - For network layers: Use standard mock connections or standard in-memory pipes (`net.Pipe()`).
   - For user input: Use interface-based input providers (like `InputProvider` containing methods like `IsKeyPressed`) and swap them with mock structures in tests.

---

## 2. Testing UI and Game Loops without opening a Windows GUI

Normally, launching `ebiten.RunGame` blocks the current thread and tries to initialize OpenGL/Metal.
To test UI loops without opening actual windows, we decouple the runner in `cmd/game/main.go` and inject mock run functions:

```go
type AppRunner struct {
    DialFn      func(network, address string) (net.Conn, error)
    RunGameFn   func(game ebiten.Game) error
    NewServerFn func(addr string, proto string) ServerRunner
}
```

In standard operation, `RunGameFn` executes `ebiten.RunGame(g)`. In unit/integration tests, `RunGameFn` simply calls `g.Update()` and/or `g.Draw()` directly inside an assertions block, verifying that state transitions or render boundaries behave correctly without booting a GUI window.

---

## 3. Example of Go Unit Tests

```go
func TestCameraCulling(t *testing.T) {
    cam := NewCamera(1024, 768, 300.0, 15)
    
    // Position camera near the top-left boundary
    cam.SetPosition(0, 0)
    
    // Assert visible grid indices
    startX, endX, startY, endY := cam.GetVisibleGridIndices(256, 256, 32)
    
    if startX != 0 || startY != 0 {
        t.Errorf("expected top-left visible bounds to start at (0,0), got (%d,%d)", startX, startY)
    }
}
```
Ensure all tests run successfully using:
```bash
go test ./...
```
