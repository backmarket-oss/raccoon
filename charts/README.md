
# raccoon

Ephemerality in kubernetes

![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square)

![AppVersion: 0.0.3](https://img.shields.io/badge/AppVersion-0.0.3-informational?style=flat-square)

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| datadog.enabled | bool | `false` |  |
| dryRun | bool | `true` |  |
| image.pullPolicy | string | `"Always"` |  |
| image.repository | string | `"tofill"` |  |
| image.tag | string | `"latest"` |  |
| namespaceToRaccoon | string | `"default"` | the namespace on which raccoon will collect pods |
| resources.limits.cpu | string | `"100m"` |  |
| resources.limits.memory | string | `"128Mi"` |  |
| resources.requests.cpu | string | `"100m"` |  |
| resources.requests.memory | string | `"128Mi"` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
