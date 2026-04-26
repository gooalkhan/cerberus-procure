# Cerberus Go

Cerberus Go is an **Isomorphic Go** project designed to share business logic across multiple distribution targets. Write your core logic once in Go and deploy it as either a static web application (WASM) or a standalone, single-binary web server.

## 🚀 Core Concept: Isomorphic Go

This project leverages Go's versatility to run in different environments:
- **WASM (WebAssembly)**: For client-side logic in static SPA deployments.
- **Native Go**: For high-performance server-side execution.
- **Unified Frontend**: A single React codebase that interacts with the business logic through an abstraction layer, regardless of the execution environment.

## 🛠 Prerequisites

- **Go**: 1.21+
- **Node.js**: 18+ (for frontend development)
- **Make**: For build automation

## 🏗 Build & Deployment

### 1. Initial Setup
Install frontend dependencies:
```bash
make frontend-install
```

### 2. Static SPA Mode (WASM)
Compiles Go business logic to WebAssembly and builds the frontend for static hosting (e.g., GitHub Pages).
```bash
make wasm
```

### 3. Single-Binary Server Mode
Builds a standalone binary that includes both the Go server and the embedded frontend assets.
```bash
make server
./server_bin
```

## 📂 Project Structure

- `cmd/`: Entry points for build targets (`wasm`, `server`).
- `internal/logic/`: **The Core.** Shared, platform-agnostic business logic (Pure Go).
- `internal/bridge/`: Adapters for WASM (`syscall/js`) and Native (`net/http`) environments.
- `internal/models/`: Shared data structures.
- `frontend/`: React frontend codebase.
    - `src/api/`: Abstraction layer that detects environment and routes logic calls.

## 🌉 Architecture

Cerberus Go uses a **Bridge Pattern** to maintain environment independence:
1. **Logic** is kept pure and unaware of its surroundings.
2. **Adapters** in the bridge layer handle environment-specific input/output.
3. The **Frontend API** detects if it's running in a WASM-enabled browser or communicating with a remote server, ensuring a seamless developer experience.

---
For more technical details, refer to [SPEC.md](./SPEC.md).
