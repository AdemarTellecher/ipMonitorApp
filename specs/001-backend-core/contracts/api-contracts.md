# Interface Contracts: Events & REST API

## 1. Wails v3 Reactive Event Bus (`contracts/events.md`)

The backend publishes data directly to the frontend Webview context via `app.Event.Emit`. The frontend registers listeners via `wails.Events.On`.

### Event: `ips-updated`

- **Direction**: Go Backend -> Webview Frontend
- **Trigger**: Emitted immediately upon completion of any background scan cycle, or after manual host CRUD/retest operations.
- **Payload Schema**:
  ```json
  {
    "total_hosts": 12,
    "online_hosts": 11,
    "offline_hosts": 1,
    "degraded_hosts": 0,
    "average_latency": 14.82,
    "overall_uptime": 91.66,
    "last_scan_time": "2026-09-11T20:15:00Z",
    "hosts": [
      {
        "id": 1,
        "name": "Cloudflare DNS",
        "address": "1.1.1.1",
        "port": 53,
        "status": "online",
        "latency": 9.41,
        "min_latency": 8.12,
        "avg_latency": 9.85,
        "max_latency": 15.20,
        "packets_sent": 120,
        "packets_recv": 120,
        "packet_loss": 0.0,
        "last_success": "2026-09-11T20:15:00Z",
        "last_failure": null
      }
    ]
  }
  ```

---

## 2. HTTP REST Endpoints (`contracts/rest-api.md`)

All REST endpoints operate on the local HTTP server embedded inside Wails v3 (serving the Webview).

### `GET /api/hosts`
- **Description**: Returns all configured hosts with current stats.
- **Response**: `200 OK` (JSON array of `Host`)

### `POST /api/hosts`
- **Description**: Adds a new target host. Validates address format and triggers an initial asynchronous probe.
- **Request Body**:
  ```json
  {
    "name": "Gateway Router",
    "address": "192.168.1.1",
    "port": 0
  }
  ```
- **Response**: `201 Created` with created `Host` object, or `400 Bad Request` with error description.

### `DELETE /api/hosts/{id}`
- **Description**: Removes target host from monitoring database.
- **Response**: `204 No Content` or `404 Not Found`.

### `POST /api/hosts/{id}/test`
- **Description**: Triggers an immediate out-of-band connectivity probe for a specific host.
- **Response**: `200 OK` with updated `Host` metrics.

### `GET /api/settings`
- **Description**: Retrieves current monitoring parameters.
- **Response**: `200 OK` with `AppSettings`.

### `PUT /api/settings`
- **Description**: Updates interval, timeouts, or concurrency workers.
- **Request Body**: Partial or full `AppSettings` payload.
- **Response**: `200 OK`.
