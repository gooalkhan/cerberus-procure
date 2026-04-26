---
trigger: always_on
glob: "**/*.{go,ts,tsx}"
description: Maintain the Isomorphic Go architecture (Logic sharing between WASM and Server).
---

# Core Mission: Isomorphic Go
This project is an **Isomorphic Go** application designed to share business logic across multiple distribution targets. Every code generation and modification must adhere to this architecture.

## 1. Distribution Targets
- **Static SPA (WASM)**: Browser runtime using WebAssembly.
- **Single-Binary Server**: OS native runtime using `net/http`.

## 2. Shared Logic Constraints
- All business logic MUST reside in `internal/logic/`.
- Logic in `internal/logic/` MUST be **pure Go** and **platform-agnostic**. It should not import `syscall/js` (WASM-specific) or `net/http` (Server-specific) directly for its core functions.
- Use interfaces to abstract environment-specific behaviors (e.g., persistence, logging).

## 3. Bridge Pattern
- Use `internal/bridge/` for environment-specific adapters.
- **WASM Bridge**: Handles `syscall/js` mapping for browser interaction.
- **Server Bridge**: Handles JSON/HTTP handlers for native execution.

## 4. Frontend Abstraction
- The frontend (`frontend/src/`) must remain target-agnostic.
- All API calls must go through the abstraction layer in `frontend/src/api/`.
- The frontend detects the environment (WASM vs. Server) and routes calls accordingly.

## 5. Persistence Strategy
- **WASM Mode**: Typically uses Mock DB or browser-based storage.
- **Server Mode**: Uses SQLite.
- Ensure repository implementations satisfy shared interfaces for both targets.
