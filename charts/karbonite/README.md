# Karbonite Helm chart

This is the Helm chart for [Karbonite](https://github.com/steromano87/karbonite) deployment.

## Install

```bash
# Enable OCI support
export HELM_EXPERIMENTAL_OCI=1

# Install chart
helm install my-karbonite oci://ghcr.io/steromano87/karbonite
```

## Values reference

| Key                          | Description                                                                                                                                                                      | Default value         |
|------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------|
| `replicaCount`               | Number of Karbonite replicas to deploy                                                                                                                                           | `1`                   |
| `imagePullSecrets`           | Optional pull secrets for image pull                                                                                                                                             | `{}`                  |
| `image.repository`           | Override for the image repository                                                                                                                                                | `steromano/karbonite` |
| `image.pullPolicy`           | Override for the image pull policy                                                                                                                                               | `IfNotPresent`        |
| `image.tag`                  | Override for the image tag. If not specified, the image version set in the chart as `appVersion` will be used.                                                                   | `""`                  |
| `nameOverride`               | Name override                                                                                                                                                                    | `""`                  |
| `fullnameOverride`           | Fully qualified name override                                                                                                                                                    | `""`                  |
| `serviceAccount.name`        | Service account name override. If not specified, it will be automatically calculated starting from the fully qualified name                                                      | `""`                  |
| `serviceAccount.annotations` | Optional annotations for the service account                                                                                                                                     | `{}`                  |
| `serviceMonitor.enabled`     | Whether to enable or not the Prometheus ServiceMonitor CRD. Note: if the `ServiceMonitor` CRD is not defined in the cluster, the CRD will be skipped regardless of this setting. | `true`                |
| `podAnnotations`             | Optional pod annotations                                                                                                                                                         | `{}`                  |

*TODO:* complete the previous table