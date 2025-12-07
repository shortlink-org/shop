# oms Service Helm Chart

This Helm chart deploys the Shop oms service to a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.30+
- Helm 3.0+
- Access to the shortlink-template chart from `oci://ghcr.io/shortlink-org/charts`

## Installing the Chart

To install the chart with the release name `oms`:

```bash
helm dependency update
helm install oms . --namespace shop --create-namespace
```

## Uninstalling the Chart

To uninstall/delete the `oms` release:

```bash
helm delete oms --namespace shop
```

## Configuration

The following table lists the configurable parameters and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceAccount.create` | Create service account | `true` |
| `ingress.enabled` | Enable ingress | `true` |
| `deploy.image.repository` | Image repository | `registry.gitlab.com/shortlink-org/shop/oms` |
| `deploy.image.tag` | Image tag | `latest` |
| `deploy.resources.limits.cpu` | CPU limit | `200m` |
| `deploy.resources.limits.memory` | Memory limit | `128Mi` or `256Mi` |
| `monitoring.enabled` | Enable monitoring | `true` |

See `values.yaml` for full configuration options.
