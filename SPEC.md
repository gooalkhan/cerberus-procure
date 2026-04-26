# Cerberus Go Project Specification

## 1. Overview
**Cerberus Go** is an isomorphic Go project designed to share business logic across multiple distribution targets. The goal is to write the core logic once in Go and deploy it as a static web application and a standalone web server.

## 2. Core Concept: Isomorphic Go
The project leverages Go's versatility to run in different environments:
- **WASM (WebAssembly)**: For client-side logic in static SPA deployments.
- **Native Go**: For server-side (net/http) execution.
- **Unified Frontend**: A single frontend codebase that interacts with the business logic through an abstraction layer, regardless of where the logic is executing.

## 3. Distribution Targets

### 3.1. Static SPA (GitHub Pages)
- **Runtime**: Browser (via WebAssembly).
- **Architecture**: The Go business logic is compiled to `.wasm`. The frontend interacts with Go via `syscall/js`.
- **Deployment**: Hosted as static files on GitHub Pages.

### 3.2. Single-Binary Web Server (net/http)
- **Runtime**: OS Native (Server).
- **Architecture**: The Go logic runs on the server. The frontend is bundled into the binary (using `go:embed`) and serves as the client. Communication happens via standard HTTP/JSON APIs.
- **Deployment**: A single standalone binary.

## 4. Architecture & Directory Structure

```text
.
├── cmd/
│   ├── wasm/           # Entry point for WASM build
│   └── server/         # Entry point for net/http server
├── internal/
│   ├── logic/          # Shared business logic (Pure Go)
│   ├── bridge/         # Adapters for WASM and HTTP
│   └── models/         # Shared data structures
├── frontend/           # Shared frontend code (React)
│   ├── src/
│   │   ├── api/        # Abstraction layer for calling Go logic
│   │   └── ...
├── build/              # Build artifacts and scripts
└── SPEC.md             # This document
```

## 5. Technology Stack
- **Language**: Go 1.21+
- **Frontend**: Vite + (React)
- **WASM**: Standard Go WASM (with `wasm_exec.js`)
- **Web Server**: `net/http` (Standard Library)
- **Persistence Strategy**: 
    - WASM: Mock DB
    - Server: SQLite

## 6. Bridge Implementation Strategy
To ensure the frontend code remains identical, an abstraction layer in the frontend will detect the environment:

1.  **WASM Mode**: Checks for `window.go` and uses WASM exports.
2.  **Server Mode**: Uses standard `fetch()` calls to the backend API.

The Go side will also use interfaces to ensure `internal/logic` remains platform-agnostic.

## 7. Build & Deployment Pipeline
- **make wasm**: Compiles Go to WASM and builds the frontend for static hosting.
- **make server**: Compiles the Go server with embedded frontend assets.
