# mgob

[![Release](https://github.com/maxisam/mgob/actions/workflows/release.yml/badge.svg)](https://github.com/maxisam/mgob/actions/workflows/release.yml)
[![Build Status](https://github.com/maxisam/mgob/actions/workflows/build.yml/badge.svg)](https://github.com/maxisam/mgob/actions/workflows/build.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/maxisam/mgob)](https://hub.docker.com/r/maxisam/mgob/)
[![GitHub release](https://img.shields.io/github/release/maxisam/mgob.svg)](https://GitHub.com/maxisam/mgob/releases/)
[![GitHub Sponsors](https://img.shields.io/github/sponsors/maxisam?label=Sponsor%20this%20project%20%E2%9D%A4%EF%B8%8F&)](https://github.com/sponsors/maxisam)
![GitHub](https://img.shields.io/github/license/maxisam/mgob)

**MGOB** is a MongoDB backup automation tool designed using Go. This fork introduces a variety of enhancements over the original repository by [stefanprodan](https://github.com/stefanprodan/mgob), which is set to be archived. Check out the [archival discussion here](https://github.com/stefanprodan/mgob/issues/161).

> Note: New features are being added to this fork exclusively.

## Enhancements in This Fork

- Backup validation
- Retry mechanism for backups
- MS Team notification support
- Updated email notification mechanism addressing the [certificate issue](https://github.com/stefanprodan/mgob/issues/160)
- Go updated to 1.21
- Dependencies updated
- Introduced `warnOnly` option for all notification channels
- Integrated Github Actions for CI/CD
- New Helm Chart with enhanced metrics, liveness probe, and other features
- Multiple Docker image releases catering to different backup solutions
- Option to skip local backup when retention is set to 0 ([#42](https://github.com/maxisam/mgob/pull/42), Credit: @aneagoe)
- On-demand restore API
- Load config from environment variables to override config file. syntax: `PLAN-ID_KEY_PROPERTY` (e.g. `mongo-test_SMTP_SERVER=smtp.company.com`)

### Helm Chart

```bash
helm pull oci://registry-1.docker.io/maxisam/mgob --version 1.8.3
helm upgrade --install mgob maxisam/mgob --namespace mgob --create-namespace
```

### Breaking Changes

- v2: in config, sftp.private_key -> sftp.privateKey

## Original Features

- schedule backups
- local backups retention
- upload to S3 Object Storage (Minio, AWS, Google Cloud, Azure)
- upload to gcloud storage
- upload to SFTP
- upload to any [Rclone](https://rclone.org/) supported storage
- notifications (Email, Slack)
- instrumentation with Prometheus
- http file server for local backups and logs
- distributed as an Alpine Docker image

## Installation

MGOB is available on Docker Hub at [maxisam/mgob](https://hub.docker.com/repository/docker/maxisam/mgob).

Supported tags:

- `maxisam/mgob:latest` latest stable [release](https://github.com/maxisam/mgob/releases)

Compatibility matrix:

| MGOB                     | MongoDB |
| ------------------------ | ------- |
| `stefanprodan/mgob:0.9`  | 3.4     |
| `stefanprodan/mgob:0.10` | 3.6     |
| `stefanprodan/mgob:1.0`  | 4.0     |
| `stefanprodan/mgob:1.1`  | 4.2     |
| `maxisam/mgob:1.10`      | 5.0     |
| `maxisam/mgob:1.12`      | 7.0     |

Docker:

```bash
docker run -dp 8090:8090 --name mgob \
    -v "/mgob/config:/config" \
    -v "/mgob/storage:/storage" \
    -v "/mgob/tmp:/tmp" \
    -v "/mgob/data:/data" \
    stefanprodan/mgob \
    -LogLevel=info
```

## Configuration

Define a backup plan (yaml format) for each database you want to backup inside the `config` dir.
The yaml file name is being used as the backup plan ID, no white spaces or special characters are allowed.

[READ MORE](.document/BACKUP_PLAN.md)

## On-Demand Operations

MGOB exposes a set of HTTP endpoints for on-demand operations like backup, restore, status, metrics, and version.

READ MORE: [On-Demand Operations](.document/ON_DEMAND_OPERATION.md)

## Logs

READ MORE: [Logs](.document/LOGS.md)

## Metrics

READ MORE: [Metrics](.document/METRICS.md)

## Restore

READ MORE: [Restore](.document/RESTORE.md)

## Special Thanks

- [stefanprodan](https://github.com/stefanprodan) for the original repository
- [<img src=".etc/deranged.svg" width="45" height="20" />](https://github.com/derangeddk)
  First awesome sponsor!!

## Sponsors [![GitHub Sponsors](https://img.shields.io/github/sponsors/maxisam?label=Sponsor%20this%20project%20%E2%9D%A4%EF%B8%8F&)](https://github.com/sponsors/maxisam)

<!-- sponsors --><!-- sponsors -->
