# Data Model & Schema Specification

## 1. Entities

### Host
Represents a network target monitored by the system.

| Field | Go Type | SQLite Type | Constraints / Description |
|-------|---------|-------------|---------------------------|
| `ID` | `int64` | `INTEGER` | Primary Key, Autoincrement |
| `Name` | `string` | `TEXT` | NOT NULL, descriptive label |
| `Address` | `string` | `TEXT` | NOT NULL, IPv4, IPv6, or valid FQDN/URL |
| `Port` | `int` | `INTEGER` | Optional TCP port (default 0 = auto-probe) |
| `Status` | `string` | `TEXT` | Enum: `online`, `offline`, `degraded`, `unknown` |
| `Latency` | `float64` | `REAL` | Current round-trip latency in milliseconds |
| `MinLatency` | `float64` | `REAL` | Minimum recorded latency |
| `AvgLatency` | `float64` | `REAL` | Moving average latency |
| `MaxLatency` | `float64` | `REAL` | Peak recorded latency |
| `PacketsSent` | `int` | `INTEGER` | Cumulative test count |
| `PacketsRecv` | `int` | `INTEGER` | Successful response count |
| `PacketLoss` | `float64` | `REAL` | Calculated percentage (0.0 to 100.0) |
| `LastSuccess` | `*time.Time`| `DATETIME` | Nullable timestamp of last valid response |
| `LastFailure` | `*time.Time`| `DATETIME` | Nullable timestamp of last packet drop/timeout |
| `CreatedAt` | `time.Time` | `DATETIME` | Creation timestamp (UTC) |
| `UpdatedAt` | `time.Time` | `DATETIME` | Last update timestamp (UTC) |

### AppSettings
Global operational configuration for the daemon and polling loop.

| Field | Go Type | SQLite Type | Default | Description |
|-------|---------|-------------|---------|-------------|
| `ID` | `int64` | `INTEGER` | `1` | Singleton row (ID=1) |
| `ScanInterval`| `int` | `INTEGER` | `10` | Scan interval in seconds (min: 3s) |
| `TimeoutMs` | `int` | `INTEGER` | `1500` | Individual probe timeout in milliseconds |
| `MaxWorkers` | `int` | `INTEGER` | `30` | Concurrency worker limit |
| `NotifyOffline`| `bool` | `INTEGER` | `true` | Trigger system notification on offline transition |

### NetworkOverview (Computed View Model)
Aggregate DTO broadcast to the frontend via the reactive event bus (`ips-updated`). Not stored as a physical table.

```go
type NetworkOverview struct {
    TotalHosts      int       `json:"total_hosts"`
    OnlineHosts     int       `json:"online_hosts"`
    OfflineHosts    int       `json:"offline_hosts"`
    DegradedHosts   int       `json:"degraded_hosts"`
    AverageLatency  float64   `json:"average_latency"`
    OverallUptime   float64   `json:"overall_uptime"`
    LastScanTime    time.Time `json:"last_scan_time"`
    Hosts           []Host    `json:"hosts"`
}
```

## 2. SQLite Database Schema

```sql
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    scan_interval INTEGER NOT NULL DEFAULT 10,
    timeout_ms INTEGER NOT NULL DEFAULT 1500,
    max_workers INTEGER NOT NULL DEFAULT 30,
    notify_offline INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS hosts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    port INTEGER DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'unknown',
    latency REAL DEFAULT 0,
    min_latency REAL DEFAULT 0,
    avg_latency REAL DEFAULT 0,
    max_latency REAL DEFAULT 0,
    packets_sent INTEGER DEFAULT 0,
    packets_recv INTEGER DEFAULT 0,
    packet_loss REAL DEFAULT 0,
    last_success DATETIME,
    last_failure DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_hosts_status ON hosts(status);
CREATE INDEX IF NOT EXISTS idx_hosts_address ON hosts(address);
```

## 3. State Transitions

```
 [Created] ───────────► [Unknown]
                            │
              ┌─────────────┴─────────────┐
        (Probe OK)                  (Probe Timeout)
              ▼                           ▼
          [Online]                    [Offline]
              ▲                           ▲
              │   (Loss > 0% & < 100%)    │
              └───►   [Degraded]   ◄──────┘
```
