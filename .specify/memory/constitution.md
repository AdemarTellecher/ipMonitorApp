# IP Monitor App Constitution

## Core Principles

### I. Go Centrality & Golden Rule (Backend Authority)
The Go backend is the definitive authority for all business logic, persistent data, network evaluations, status transitions, and metric aggregations. The frontend is strictly a presentational view ("Dumb UI") and must never perform calculations, re-sorting of operational priority, or independent state syntheses. All domain logic lives in `internal/`.

### II. Zero Polling & Pure Reactive Communication
Continuous client polling (`setInterval`, repeated fetches) is strictly prohibited. State transitions, periodic background scan results, and metric broadcasts must be pushed reactively via the Wails v3 event bus (`app.Event.Emit("ips-updated", ...)`). The frontend is entirely event-driven.

### III. Ultra Portability & Pure Go Zero-CGO SQLite
The application must compile without dependencies on external C toolchains (CGO_ENABLED=0), enabling clean, seamless cross-compilation. Database persistence uses `modernc.org/sqlite` in pure Go. Binaries are self-contained and embed frontend assets.

### IV. Multiplatform Resilience & OS Sandboxing
Filesystem paths, process execution, and network privileges must respect OS-specific constraints:
- **macOS (Darwin):** SQLite database path must always resolve to `~/Library/Application Support/IPMonitor/<name>.db` (never inside `.app` bundle or current working directory). Unprivileged UDP ICMP sockets (`pinger.SetPrivileged(false)`) and App Transport Security local networking allowances are mandatory. Window close intercepts hide the window instead of killing the app, reopening on Dock tap.
- **Windows:** SQLite resolves locally for portable deployment. ICMP uses privileged raw sockets (`pinger.SetPrivileged(true)`). Any external process execution (such as fallback ping) must hide the console window (`HideWindow: true`, `CREATE_NO_WINDOW`) to prevent cmd flashes.
- **Linux:** Unprivileged UDP sockets or standard fallbacks with appropriate error resilience.

### V. Hybrid Connectivity Verification & Defense-in-Depth
Host reachability must not rely on a single naive socket probe. The network engine operates a staged fallback pipeline:
1. Direct ICMP Ping (OS-tailored privileges);
2. Native OS Ping tool execution (with hidden console flags);
3. TCP Port Probes (e.g., ports 80, 443, 22, 53, 3389) when ICMP is blocked or restricted by local firewalls.
Any failure or partial latency return must degrade gracefully without crashing or hanging worker goroutines.

## Architecture & Storage Constraints

- **SQLite Schema & Concurrency:** WAL (Write-Ahead Logging) mode and busy timeouts must be enabled on every SQLite connection. Schema migrations are managed programmatically in Go upon startup.
- **Dual Service Surface:** Business logic in `internal/service/` is exposed to the frontend via native Wails v3 service bindings as well as local HTTP REST handlers (`/api/...`) for maximum flexibility and webview compatibility.
- **Concurrent Worker Pools:** Background network scans must execute concurrently using controlled goroutine worker pools or bounded waitgroups with cancellation contexts to prevent connection exhaustion or goroutine leaks.

## Development Workflow & Quality Gates

1. **Contract Integrity:** Any change to backend data structs (`internal/model/`) must be reflected and synchronized in the bindings, REST payloads, and frontend types.
2. **Build Verification:** Changes must cleanly compile across platforms with `CGO_ENABLED=0` and pass unit tests (`go test ./...`).
3. **Multiplatform Release Standards:** All releases must strictly adhere to the packaging, codesigning, and path resolution rules defined in the `multiplatform-release-guide`.

## Governance

- The Constitution supersedes all ad-hoc architecture decisions and transient coding shortcuts.
- Any architectural change (e.g., adding CGO dependencies, introducing frontend polling, altering database path resolution) requires an amendment to this Constitution, justification in documentation, and a clear migration plan.
- All code reviews and agent tasks must verify strict compliance with these principles.

**Version**: 1.0.0 | **Ratified**: 2026-09-11 | **Last Amended**: 2026-09-11
