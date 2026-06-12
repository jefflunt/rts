# Onboarding & Setup Guide

Welcome to the team! This document details the step-by-step instructions to get the codebase running, build dependencies, and start contributing.

---

## 💻 System Dependencies

Because the project relies on **Ebitengine (Ebiten v2)** for drawing graphics and capturing mouse/keyboard inputs, your system requires native graphics and audio drivers.

### 🍎 macOS Requirements
- Xcode Command Line Tools (required for Cgo compiling).
- No extra graphics libraries are required; Metal/OpenGL are supported natively.

### 🐧 Linux Requirements
Ensure you have the development packages for `X11`, `OpenGL`, and `ALSA` installed.
- **Ubuntu/Debian**:
  ```bash
  sudo apt-get install -y libasound2-dev libgl1-mesa-dev xorg-dev libx11-dev libxrandr-dev libxi-dev libxcursor-dev libxinerama-dev libxxf86vm-dev
  ```

### 🪟 Windows Requirements
- No extra dependencies are typically needed. Ensure you have modern graphics card drivers installed.

---

## 🚀 Setup Steps

1. **Verify Go Installation**:
   Verify you are using Go version 1.20 or newer:
   ```bash
   go version
   ```

2. **Clone and Navigate**:
   ```bash
   git clone <repo-url>
   cd rts
   ```

3. **Install Dependencies**:
   Tidy up the Go module workspace to fetch required packages (e.g. Ebitengine):
   ```bash
   go mod tidy
   ```

---

## 🏃 Running the Application

The codebase has an integrated launcher supporting multiple modes:

### Local (Offline Loopback Mode)
Launches the server in a background thread and client in the main thread:
```bash
go run cmd/game/main.go --mode local
```

### Dedicated Server Mode
Launches only the authoritative TCP server:
```bash
go run cmd/game/main.go --mode server --addr "127.0.0.1:8080"
```

### Dedicated Client Mode
Launches the client and attempts to connect to the dedicated server:
```bash
go run cmd/game/main.go --mode client --addr "127.0.0.1:8080"
```

---

## 🧪 Testing and Quality Control

We practice **strict Test-Driven Development (TDD)** on this codebase. Before submitting any changes:

1. **Run All Tests**:
   Ensure everything is green:
   ```bash
   go test ./... -v
   ```

2. **Add Tests**:
   - Every file change should have a corresponding `_test.go` file.
   - Do not write implementation code without first writing a failing test.
