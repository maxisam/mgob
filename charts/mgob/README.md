# mgob

![Version: 2.1.0](https://img.shields.io/badge/Version-2.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 2.0.18](https://img.shields.io/badge/AppVersion-2.0.18-informational?style=flat-square)

A Helm chart for Mgob,  MongoDB dockerized backup agent.
Runs scheduled backups with retention, S3 & SFTP upload, notifications, instrumentation with Prometheus and more.

## Source Code

* <https://github.com/maxisam/mgob>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | common | 1.x.x |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| kubeVersion | string | `""` |  |
| nameOverride | string | `""` |  |
| fullnameOverride | string | `""` |  |
| namespaceOverride | string | `""` |  |
| commonLabels | object | `{}` |  |
| commonAnnotations.app | string | `"mgob"` |  |
| clusterDomain | string | `"cluster.local"` |  |
| logLevel | string | `"info"` |  |
| image.registry | string | `"docker.io"` |  |
| image.repository | string | `"maxisam/mgob"` |  |
| image.tag | string | `"2.0.18-all"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.pullSecrets | list | `[]` |  |
| envSecrets.annotations | object | `{}` |  |
| envSecrets.secrets | object | `{}` |  |
| extraEnvSecrets | list | `[]` |  |
| mountSecrets | list | `[]` |  |
| env | list | `[]` |  |
| config | object | `{}` | Backup plans. For details, see [values.yaml](values.yaml) |
| replicaCount | int | `1` |  |
| sidecars | list | `[]` |  |
| lifecycleHooks | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| updateStrategy.type | string | `"RollingUpdate"` |  |
| podAffinityPreset | string | `""` |  |
| podAntiAffinityPreset | string | `"soft"` |  |
| nodeAffinityPreset.type | string | `""` |  |
| nodeAffinityPreset.key | string | `""` |  |
| nodeAffinityPreset.values | list | `[]` |  |
| affinity | object | `{}` |  |
| nodeSelector | object | `{}` |  |
| tolerations | list | `[]` |  |
| resources.limits.cpu | string | `"500m"` |  |
| resources.limits.memory | string | `"256Mi"` |  |
| resources.requests.cpu | string | `"100m"` |  |
| resources.requests.memory | string | `"128Mi"` |  |
| podSecurityContext.fsGroup | int | `65534` |  |
| containerSecurityContext.enabled | bool | `false` |  |
| containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| containerSecurityContext.privileged | bool | `false` |  |
| containerSecurityContext.runAsUser | int | `0` |  |
| containerSecurityContext.runAsNonRoot | bool | `false` |  |
| livenessProbe.enabled | bool | `true` |  |
| livenessProbe.initialDelaySeconds | int | `5` |  |
| livenessProbe.timeoutSeconds | int | `1` |  |
| livenessProbe.periodSeconds | int | `20` |  |
| livenessProbe.failureThreshold | int | `6` |  |
| livenessProbe.successThreshold | int | `1` |  |
| readinessProbe.enabled | bool | `true` |  |
| readinessProbe.initialDelaySeconds | int | `5` |  |
| readinessProbe.timeoutSeconds | int | `1` |  |
| readinessProbe.periodSeconds | int | `20` |  |
| readinessProbe.failureThreshold | int | `6` |  |
| readinessProbe.successThreshold | int | `1` |  |
| startupProbe.enabled | bool | `true` |  |
| startupProbe.initialDelaySeconds | int | `10` |  |
| startupProbe.timeoutSeconds | int | `1` |  |
| startupProbe.periodSeconds | int | `20` |  |
| startupProbe.failureThreshold | int | `6` |  |
| startupProbe.successThreshold | int | `1` |  |
| customLivenessProbe | object | `{}` |  |
| customStartupProbe | object | `{}` |  |
| customReadinessProbe | object | `{}` |  |
| service.type | string | `"ClusterIP"` |  |
| service.ports.http | int | `8090` |  |
| service.nodePorts.http | string | `""` |  |
| service.clusterIP | string | `""` |  |
| service.extraPorts | list | `[]` |  |
| service.loadBalancerIP | string | `""` |  |
| service.loadBalancerSourceRanges | list | `[]` |  |
| service.externalTrafficPolicy | string | `"Cluster"` |  |
| service.annotations | object | `{}` |  |
| service.sessionAffinity | string | `"None"` |  |
| service.sessionAffinityConfig | object | `{}` |  |
| ingress.enabled | bool | `false` |  |
| ingress.pathType | string | `"Prefix"` |  |
| ingress.apiVersion | string | `""` |  |
| ingress.hostname | string | `"mgob.local"` |  |
| ingress.path | string | `"/"` |  |
| ingress.annotations | object | `{}` |  |
| ingress.tls | bool | `false` |  |
| ingress.tlsSecretName | string | `""` |  |
| ingress.extraPaths | list | `[]` |  |
| ingress.selfSigned | bool | `false` |  |
| ingress.ingressClassName | string | `"nginx"` |  |
| ingress.extraHosts | list | `[]` |  |
| ingress.extraTls | list | `[]` |  |
| ingress.secrets | list | `[]` |  |
| ingress.extraRules | list | `[]` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automountServiceAccountToken | bool | `true` |  |
| metrics.serviceMonitor.enabled | bool | `true` |  |
| metrics.serviceMonitor.port | string | `"http"` |  |
| metrics.serviceMonitor.namespace | string | `""` |  |
| metrics.serviceMonitor.interval | string | `"30s"` |  |
| metrics.serviceMonitor.scrapeTimeout | string | `"10s"` |  |
| storage.longTerm | object | `{"name":"mgob-storage","spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"100Gi"}},"storageClassName":"standard"}}` | Persistent volume for backups, see `config.retention` |
| storage.tmp | object | `{"name":"mgob-tmp","spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"10Gi"}},"storageClassName":"standard"}}` | Persistent volume for temporary files |
| storage.restoreTmp.name | string | `"mgob-restore-tmp"` |  |
| storage.restoreTmp.spec.accessModes[0] | string | `"ReadWriteOnce"` |  |
| storage.restoreTmp.spec.resources.requests.storage | string | `"100Gi"` |  |
| storage.restoreTmp.spec.storageClassName | string | `"standard"` |  |
| mongodb.enabled | bool | `true` |  |
| mongodb.port | int | `27017` |  |
| mongodb.image.registry | string | `"docker.io"` |  |
| mongodb.image.repository | string | `"mongo"` |  |
| mongodb.image.tag | string | `"4.4.6"` |  |
| mongodb.image.pullPolicy | string | `"IfNotPresent"` |  |
| mongodb.securityContext.enabled | bool | `true` |  |
| mongodb.securityContext.allowPrivilegeEscalation | bool | `false` |  |
| mongodb.securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| mongodb.securityContext.privileged | bool | `false` |  |
| mongodb.securityContext.runAsUser | int | `1001` |  |
| mongodb.securityContext.runAsNonRoot | bool | `false` |  |
| mongodb.resources.limits.cpu | string | `"500m"` |  |
| mongodb.resources.limits.memory | string | `"512Mi"` |  |
| mongodb.resources.requests.cpu | string | `"200m"` |  |
| mongodb.resources.requests.memory | string | `"300Mi"` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
