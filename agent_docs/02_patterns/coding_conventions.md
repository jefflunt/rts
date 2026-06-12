# Coding Conventions

To keep the codebase maintainable, clean, and robust, all code contributions must conform to the following Go coding style guidelines and structural conventions.

---

## 1. Style & Idioms

- **Standard Formatting**: All Go code must be formatted with `gofmt` and import blocks organized via `goimports`.
- **Explicit Dependency Injection**: No packages should rely on global mutable state or implicit package-level variables. Always pass dependencies (such as input providers, camera configurations, or connection wrappers) as parameters or inject them during initialization.
- **Interfaces for Decoupling**: Define interface types on the consumer side to allow mock-based testing (e.g., `InputProvider` or `ServerRunner`).
- **No Uncontrolled Panics**: Never use `panic` or `recover` for normal error states. Always return explicit `error` values.
- **Error Wrapping**: When bubble up errors, wrap them with descriptive context using `fmt.Errorf("failed to do X: %w", err)`.

---

## 2. Directory & Package Structure

- **`cmd/`**: Entrypoints of the application. Keep CLI arg parsing and bootstrap logic thin. Move execution blocks into testable runner structures.
- **`internal/`**: All business logic. Go enforces that packages inside `internal/` cannot be imported by external modules, protecting the boundaries.
  - **`internal/client/`**: Graphic updates, inputs, cameras, and user-facing views. Only imports from `internal/map` and `internal/protocol`.
  - **`internal/server/`**: Network listener, packet processing. Only imports from `internal/map` and `internal/protocol`.
  - **`internal/map/`**: Coordinate translations and structural layout. It has zero external dependencies (not even `ebiten` or `protocol`).
  - **`internal/protocol/`**: Handles the wire format, serialization protocols, and encoding/decoding.

---

## 3. Idiomatic Constructors

Always provide constructor functions for structs containing state, named `New<StructName>` or `NewDefault<StructName>`:

```go
type Camera struct {
    x, y          float64
    viewportWidth int
}

func NewCamera(w, h int) *Camera {
    return &Camera{
        viewportWidth: w,
        // ...
    }
}
```
Keep struct fields private (lowercase) when they should not be modified by external packages, exposing getters/setters as needed.
