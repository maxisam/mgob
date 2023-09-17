# On-Demand Operations

This document describes the on-demand operations available through `mgob`'s Web API endpoints.

## Available API Endpoints

| Endpoint                 | Description                        |
| ------------------------ | ---------------------------------- |
| `mgob-host:8090/storage` | File server                        |
| `mgob-host:8090/status`  | Backup job statuses                |
| `mgob-host:8090/metrics` | Prometheus metrics endpoint        |
| `mgob-host:8090/version` | `mgob` version and runtime details |
| `mgob-host:8090/debug`   | pprof debugging endpoint           |
| `mgob-host:8090/restore` | Restore API                        |

## Performing On-Demand Operations

### On-Demand Backup

**Endpoint:** HTTP POST `mgob-host:8090/backup/:planID`

**Example:**

```bash
curl -X POST http://mgob-host:8090/backup/mongo-debug
```

**Response:**

```json
{
  "plan": "mongo-debug",
  "file": "mongo-debug-1494256295.gz",
  "duration": "3.635186255s",
  "size": "455 kB",
  "timestamp": "2017-05-08T15:11:35.940141701Z"
}
```

### Retrieving Scheduler Status

**Endpoint:**

- HTTP GET `mgob-host:8090/status`
- HTTP GET `mgob-host:8090/status/:planID`

**Example:**

```bash
curl -X GET http://mgob-host:8090/status/mongo-debug
```

**Response:**

```json
{
  "plan": "mongo-debug",
  "next_run": "2017-05-13T14:32:00+03:00",
  "last_run": "2017-05-13T11:31:00.000622589Z",
  "last_run_status": "200",
  "last_run_log": "Backup finished in 2.339055539s archive mongo-debug-1494675060.gz size 527 kB"
}
```

### On-Demand Restoration

To restore a backup from within the mgob container, use the on-demand /restore/:planID/:file API.

**Endpoint:** HTTP POST `mgob-host:8090/restore/:planID/:file`

**Example:**

```bash
curl -X POST http://mgob-host:8090/restore/mongo-test/mongo-test-1494056760.gz
```

**Response:**

```json
{
  "plan": "mongo-test",
  "name": "mongo-test-1494056760.gz",
  "duration": "2.4180213s",
  "timestamp": "2017-05-06T14:52:40.000000001Z"
}
```

**Special Thanks**

[<img src="../.etc/deranged.svg" width="45" height="20" />](https://github.com/derangeddk) for sponsoring this feature.
