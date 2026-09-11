# Implementation Plan: Backend Core & Resilient Multiplatform Engine

**Branch**: `001-backend-core` | **Date**: 2026-09-11 | **Spec**: [spec.md](file:///c:/Users/conta/OneDrive/AppProjects/GolangProjects/ipMonitorApp/specs/001-backend-core/spec.md)

**Input**: Feature specification from `specs/001-backend-core/spec.md` and PRD-Backend.

## Summary

Implement and standardize the core Go backend architecture for IP Monitor App using Wails v3 and pure Go SQLite (`modernc.org/sqlite`). The engine features dynamic cross-platform database path resolution (sandboxing on macOS, portability on Windows/Linux), silent OS command execution on Windows, privileged vs unprivileged ICMP socket handling, a 3-tier hybrid connectivity fallback (ICMP -> OS Ping -> TCP Probe), concurrent worker-pool scans, reactive event emission (`ips-updated`), and a dual service surface (native Wails bindings + REST `/api/...`).

## Technical Context

**Language/Version**: Go 1.25+ (`CGO_ENABLED=0`)
**Primary Dependencies**: Wails v3 (`v3.0.0-beta.17`), `modernc.org/sqlite` (Pure Go), `github.com/go-ping/ping`
**Storage**: SQLite 3 embedded (Pure Go), WAL mode enabled, `busy_timeout=5000`
**Testing**: `go test ./...` with mock network targets and in-memory/temp SQLite databases
**Target Platform**: Windows (`amd64`), macOS (`darwin/arm64`, `darwin/amd64`), Linux (`amd64`)
**Project Type**: Desktop application (Wails v3 hybrid Go + WebView)
**Performance Goals**: App launch < 1.5s; 50 concurrent target scans in < 3s
**Constraints**: Zero CGO, zero terminal windows flashing on Windows (`CREATE_NO_WINDOW`), zero macOS App Bundle write errors (`~/Library/Application Support/IPMonitor/`), zero client polling
**Scale/Scope**: Up to 200 monitored hosts per instance, background tick interval 5s–300s

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Go Centrality & Golden Rule**: All state transitions, reachability determinations, metrics calculations, and sorting are executed in Go (`internal/service/`, `internal/model/`).
- [x] **II. Zero Polling & Pure Reactive Communication**: UI relies exclusively on `ips-updated` events emitted via `app.Event.Emit`.
- [x] **III. Ultra Portability & Pure Go Zero-CGO SQLite**: `CGO_ENABLED=0` strictly enforced; `modernc.org/sqlite` used for DB access.
- [x] **IV. Multiplatform Resilience & OS Sandboxing**: `ResolveDBPath` targets `~/Library/Application Support/IPMonitor/` on Darwin and executable directory on Windows. Windows ping hides console window.
- [x] **V. Hybrid Connectivity Verification**: ICMP direct -> native ping fallback -> TCP port probe.

## Project Structure

### Documentation (this feature)

```text
specs/001-backend-core/
├── plan.md              # This file
├── research.md          # Technical decisions and rationale
├── data-model.md        # Entities, validation, schema, lifecycle
├── quickstart.md        # Validation scenarios and testing steps
├── contracts/           # API and Event bus schemas
│   ├── events.md        # Wails v3 event payloads
│   └── rest-api.md      # REST endpoints contract
└── tasks.md             # (Created in /speckit-tasks)
```

### Source Code (repository root)

```text
ipMonitorApp/
├── main.go                       # Entrypoint: Wails v3 lifecycle, bindings registration, periodic worker
├── internal/
│   ├── model/                    # Data models, SQLite connection, migrations, path resolver
│   │   ├── host.go               # Host and NetworkOverview structs
│   │   ├── settings.go           # App configuration entity
│   │   ├── db.go                 # SQLite initialization (WAL, busy timeout, tables)
│   │   └── resolver.go           # Dynamic cross-platform DB path resolution
│   ├── service/                  # Business logic and network engine
│   │   ├── monitor.go            # Periodic scanning engine & worker pool
│   │   ├── pinger.go             # Hybrid probe: ICMP, OS Ping fallback, TCP probe
│   │   └── host_service.go       # CRUD and orchestrator
│   └── api/                      # Dual interface layer
│       ├── rest.go               # HTTP REST handlers (/api/hosts, /api/settings, /api/ping)
│       └── bindings.go           # Wails v3 native binding adapters
├── frontend/                     # Dumb UI (HTML, CSS, JS/TS)
└── build/                        # Packaging assets (Info.plist, app icons, manifests)
```

**Structure Decision**: Standard Go internal package structure (`internal/model`, `internal/service`, `internal/api`). Business rules and domain logic are strictly encapsulated in `internal/`, isolated from UI presentation.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| *None*    | *N/A*      | Architecture conforms 100% to constitution. |
