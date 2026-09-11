# Technical Research & Architectural Decisions

## 1. SQLite Pure Go Driver & WAL Mode

- **Decision**: Use `modernc.org/sqlite` with `CGO_ENABLED=0` rather than `mattn/go-sqlite3`.
- **Rationale**: `mattn/go-sqlite3` relies heavily on CGO and an external C compiler (GCC/MinGW). `modernc.org/sqlite` is translated directly from C to pure Go, enabling zero-CGO compilation and simple cross-compilation from any host to Windows, macOS, and Linux without native toolchain hurdles.
- **Connection Optimization**:
  - DSN parameters: `_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)`.
  - WAL mode permits concurrent reads while a write transaction is active, avoiding lock contention during background polling updates.
- **Alternatives Considered**:
  - `mattn/go-sqlite3`: Rejected due to strict requirement for CGO and compilation friction in cross-platform pipelines.
  - JSON file storage: Rejected due to concurrency hazards, lack of indexing, and poor durability during system crashes.

## 2. Dynamic Cross-Platform Database Path Resolution

- **Decision**: Resolve database path based on `runtime.GOOS`:
  - **macOS (`darwin`)**: `filepath.Join(os.UserHomeDir(), "Library", "Application Support", "IPMonitor", "ipmonitor.db")`
  - **Windows / Linux**: `filepath.Join(filepath.Dir(os.Executable()), "ipmonitor.db")` with fallback to user config directory if executable directory is read-only.
- **Rationale**:
  - macOS App Bundles (`/Applications/IP Monitor.app/Contents/MacOS`) are root-owned and read-only. Attempting to create SQLite databases in the executable directory triggers instant crashes with `readonly database` / `permission denied`.
  - Windows portable apps favor zero-install co-location of data files beside the `.exe`.
- **Alternatives Considered**:
  - Unconditionally using `os.UserConfigDir()`: Rejected on Windows because users expect portable `.exe` folders to retain their database.
  - Checking `.app/Contents/MacOS` in the path string: Rejected because macOS Translocation / Gatekeeper alters the path string dynamically, breaking substring checks.

## 3. Multi-tier Hybrid Connectivity Probing Engine

- **Decision**: A 3-step sequential probe with bounded timeouts:
  1. **Direct ICMP Socket (`go-ping/ping`)**:
     - Windows: `pinger.SetPrivileged(true)` (WinSock raw socket API).
     - Darwin / Linux: `pinger.SetPrivileged(false)` (UDP ICMP unprivileged socket without root).
  2. **Native OS Ping Executable Fallback**:
     - macOS: `/sbin/ping -c 1 -W 1500 <target>` (has system `setuid-root`).
     - Windows: `cmd.exe /C ping -n 1 -w 1500 <target>` wrapped in `syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}` (`CREATE_NO_WINDOW`).
  3. **TCP Port Fallback**:
     - `net.DialTimeout("tcp", net.JoinHostPort(host, port), 800*time.Millisecond)` for common ports (80, 443, 22, 53, 3389).
- **Rationale**: Corporate intranets and home routers often disable ICMP Echo reply or require elevated privileges. TCP probes ensure that servers hosting active HTTP/SSH services are correctly reported as reachable.
- **Alternatives Considered**:
  - Only ICMP: Results in false-negative "offline" indicators for hardened firewalls.
  - Only TCP: Ineffective for network switches, gateways, and printers that don't expose web ports.

## 4. Concurrency & Goroutine Pool Management

- **Decision**: Bounded worker pool using buffered channels or `sync.WaitGroup` with a semaphore channel (capacity = `maxWorkers`, default 30).
- **Rationale**: Scanning 100+ hosts simultaneously without throttling can exhaust OS socket descriptors, trigger local security alarms, and induce socket buffer packet drops.
- **Cancellation**: Every scan cycle accepts a `context.WithTimeout(ctx, cycleTimeout)` to ensure stale DNS lookups or hung TCP handshakes terminate cleanly before the next cycle begins.
