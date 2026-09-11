# Quickstart & Verification Guide: Backend Core

This guide details steps to validate the Go backend core, multiplatform resilience, and reactive data pipeline.

## 1. Prerequisites

- Go 1.25+ installed (`go version`)
- Git (`git --version`)
- Windows, macOS, or Linux workstation

## 2. Validation Scenarios

### Scenario A: Zero-CGO Build & Cross-Compilation Check

Verify that the codebase compiles with CGO disabled and without native C compilers:

```powershell
# In PowerShell (Windows)
$env:CGO_ENABLED="0"
go build -v -o ipmonitor-test.exe main.go
```

**Expected Outcome**: Compilation succeeds without linker errors or warnings regarding missing GCC/MinGW.

---

### Scenario B: SQLite Path Resolution & Migration Sanity

Execute unit tests for `model.ResolveDBPath`:

```powershell
go test -v ./internal/model/... -run TestResolveDBPath
```

**Expected Outcome**:
- On Windows: returns executable directory.
- On macOS simulation: returns `~/Library/Application Support/IPMonitor/ipmonitor.db`.
- Database schema tables (`hosts`, `settings`) auto-migrate without errors.

---

### Scenario C: Hybrid Connectivity Engine & Process Isolation

Test probing against active and inactive targets:

```powershell
go test -v ./internal/service/... -run TestPingerHybrid
```

**Expected Outcome**:
- Valid host (e.g. `1.1.1.1` or `8.8.8.8`) returns `status: online` and valid latency in ms.
- Fallback ping does not launch any visible `cmd.exe` or terminal window on Windows.
- Unreachable IP (e.g. `192.0.2.1` TEST-NET-1) times out cleanly after `TimeoutMs` and flags `status: offline`.

---

### Scenario D: Reactive Event Emission Simulation

Run integration test for the periodic monitor worker:

```powershell
go test -v ./internal/service/... -run TestMonitorWorkerEventEmission
```

**Expected Outcome**:
- Background scanner executes batch of hosts within the configured worker pool limits.
- On completion of scan round, `ips-updated` event is dispatched with complete `NetworkOverview` JSON payload.
