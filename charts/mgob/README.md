# mgob

![Version: 1.8.4](https://img.shields.io/badge/Version-1.2.1-informational?style=flat-square) ![AppVersion: 1.3](https://img.shields.io/badge/AppVersion-1.3-informational?style=flat-square)

MongoDB dockerized backup agent.
Runs scheduled backups with retention, S3 & SFTP upload, notifications, instrumentation with Prometheus and more.

## Maintainers

| Name   | Email                      | Url |
| ------ | -------------------------- | --- |
| endrec | endre.czirbesz@rungway.com |     |

## Source Code

- <https://github.com/stefanprodan/mgob>

## Values

| Key                        | Type   | Default                                                                                   | Description                                                                                                |
| -------------------------- | ------ | ----------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ---- | ---- | ----- | ----- | --------------------------------------------------- |
| config                     | object | `{}`                                                                                      | Backup plans. For details, see [values.yaml](values.yaml)                                                  |
| env                        | object | `{}`                                                                                      |                                                                                                            |
| fullnameOverride           | string | `""`                                                                                      |                                                                                                            |
| image.pullPolicy           | string | `"IfNotPresent"`                                                                          | Image pull policy                                                                                          |
| image.repository           | string | `"stefanprodan/mgob"`                                                                     | Image repo                                                                                                 |
| image.tag                  | string | `""`                                                                                      | Image tag Overrides the image tag whose default is the chart appVersion.                                   |
| ingress.annotations        | object | `{}`                                                                                      |                                                                                                            |
| ingress.enabled            | bool   | `false`                                                                                   |                                                                                                            |
| ingress.hosts              | object | `{}`                                                                                      |                                                                                                            |
| ingress.tls                | object | `{}`                                                                                      |                                                                                                            |
| logLevel                   | string | `"info"`                                                                                  | log level (debug                                                                                           | info | warn | error | fatal | panic) WARNING! debug logs might include passwords! |
| nameOverride               | string | `""`                                                                                      |                                                                                                            |
| podSecurityContext         | object | `{"fsGroup":65534}`                                                                       | Pod Security Context ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/       |
| replicaCount               | int    | `1`                                                                                       | Number of replicas                                                                                         |
| resources                  | object | `{"limits":{"cpu":"100m","memory":"128Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}`   | Resource requests and limits ref: http://kubernetes.io/docs/user-guide/compute-resources/                  |
| secret                     | object | `{}`                                                                                      | Secret(s) to mount. For details, see [values.yaml](values.yaml)                                            |
| securityContext            | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"privileged":false}`   | Container Security Context ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| service.externalPort       | int    | `8090`                                                                                    | Port to access the service                                                                                 |
| service.internalPort       | int    | `8090`                                                                                    | Port to connect to in pod                                                                                  |
| service.name               | string | `"mgob"`                                                                                  | Service name                                                                                               |
| serviceAccount.annotations | object | `{}`                                                                                      | Annotations to add on service account                                                                      |
| serviceAccount.create      | bool   | `true`                                                                                    | If false, default service account will be used                                                             |
| serviceAccount.name        | string | `""`                                                                                      |                                                                                                            |
| storage.longTerm           | object | `{"accessMode":"ReadWriteOnce","name":"mgob-storage","size":"10Gi","storageClass":"gp2"}` | Persistent volume for backups, see `config.retention`                                                      |
| storage.tmp                | object | `{"accessMode":"ReadWriteOnce","name":"mgob-tmp","size":"3Gi","storageClass":"gp2"}`      | Persistent volume for temporary files                                                                      |

---

Autogenerated from chart metadata using [helm-docs v1.6.0](https://github.com/norwoodj/helm-docs/releases/v1.6.0)
