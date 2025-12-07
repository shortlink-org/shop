# Helm Charts for Shop Services

This document describes the Helm charts available for deploying Shop services to Kubernetes clusters.

## Overview

Each service in the Shop Boundary has its own Helm chart located in `<service>/ops/<service>/`. These charts follow the pattern established in the [shortlink-org/shortlink](https://github.com/shortlink-org/shortlink) repository and use the `shortlink-template` chart as a dependency.

## Available Charts

| Service | Chart Location | Description | Port(s) |
|---------|----------------|-------------|---------|
| admin | `admin/ops/admin/` | Shop admin service | 8000 |
| bff | `bff/ops/bff/` | API Gateway (Wundergraph) | 9991 |
| courier-emulation | `courier-emulation/ops/courier-emulation/` | Courier emulation service | 50051, 9090 |
| delivery | `delivery/ops/delivery/` | Delivery service (Rust) | 50051, 9090 |
| email-subscription | `email-subscription/ops/email-subscription/` | Email subscription service | 8000 |
| feed | `feed/ops/feed/` | Feed service (Go) | 50051, 9090 |
| geolocation | `geolocation/ops/geolocation/` | Geolocation service (Go) | 50051, 9090 |
| merch | `merch/ops/merch/` | Merchandise service | 50051, 9090 |
| oms | `oms/ops/oms/` | Order Management Service (Temporal) | 50051, 9090 |
| oms-graphql | `oms-graphql/ops/oms-graphql/` | GraphQL API Bridge | 8080 |
| pricer | `pricer/ops/pricer/` | Price service (Go) | 50051, 9090 |
| support | `support/ops/support/` | Support service (PHP) | 8080 |
| ui | `ui/ops/ui/` | Shop UI (NextJS) | 3000 |

## Prerequisites

- Kubernetes 1.30.0 or higher
- Helm 3.0 or higher
- Access to `oci://ghcr.io/shortlink-org/charts` registry for the `shortlink-template` dependency

## Chart Structure

Each chart follows this structure:

```
<service>/ops/<service>/
├── Chart.yaml          # Chart metadata and dependencies
├── values.yaml         # Default configuration values
├── README.md          # Service-specific documentation
├── .helmignore        # Files to ignore when packaging
├── templates/
│   ├── base.yaml      # Main template that includes common resources
│   └── NOTES.txt      # Post-installation notes
└── charts/            # Downloaded dependencies (gitignored)
```

## Installation

### Installing a Single Service

1. Navigate to the service's chart directory:
   ```bash
   cd <service>/ops/<service>/
   ```

2. Update chart dependencies:
   ```bash
   helm dependency update
   ```

3. Install the chart:
   ```bash
   helm install <release-name> . --namespace shop --create-namespace
   ```

### Example: Installing the Feed Service

```bash
cd feed/ops/feed/
helm dependency update
helm install shop-feed . --namespace shop --create-namespace
```

## Configuration

Each chart can be configured through `values.yaml`. Common configuration options include:

### Service Account
```yaml
serviceAccount:
  create: true  # Create a service account for the service
```

### Ingress Configuration

For gRPC services (using Istio):
```yaml
ingress:
  enabled: true
  ingressClassName: istio
  istio:
    match:
      - uri:
          prefix: /shop.service.v1.ServiceName/
    route:
      destination:
        port: 50051
```

For HTTP services (using NGINX):
```yaml
ingress:
  enabled: true
  ingressClassName: nginx
  annotations:
    cert-manager.io/cluster-issuer: cert-manager-production
  hostname: shop.shortlink.best
  paths:
    - path: /service-path
      service:
        name: shop-service
        port: 8080
```

### Deployment Configuration

```yaml
deploy:
  type: Deployment  # Can be: Deployment, Rollout, or StatefulSet
  
  image:
    repository: registry.gitlab.com/shortlink-org/shop/<service>
    tag: latest
    pullPolicy: Always
  
  resources:
    limits:
      cpu: 200m
      memory: 128Mi
    requests:
      cpu: 50m
      memory: 64Mi
  
  env:
    TRACER_URI: grafana-tempo.grafana:4317
    # Service-specific environment variables
  
  livenessProbe:
    enabled: true
    httpGet:
      path: /live
      port: 9090
  
  readinessProbe:
    enabled: true
    httpGet:
      path: /ready
      port: 9090
```

### Service Configuration

```yaml
service:
  type: ClusterIP
  ports:
    - name: grpc  # or http
      port: 50051
      protocol: TCP
      public: true
    - name: metrics
      port: 9090
      protocol: TCP
      public: true
```

### Monitoring

```yaml
monitoring:
  enabled: true  # Enable Prometheus ServiceMonitor
```

### Network Policy

```yaml
networkPolicy:
  enabled: false  # Enable network policies
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: shortlink
  policyTypes:
    - Ingress
```

### Pod Disruption Budget

```yaml
podDisruptionBudget:
  enabled: false  # Enable PDB for high availability
```

## Overriding Values

You can override values in several ways:

### Using --set flag
```bash
helm install shop-feed feed/ops/feed/ \
  --set deploy.image.tag=v1.2.3 \
  --set deploy.replicaCount=3
```

### Using a custom values file
```bash
# Create custom-values.yaml
cat > custom-values.yaml << EOF
deploy:
  image:
    tag: v1.2.3
  replicaCount: 3
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
EOF

helm install shop-feed feed/ops/feed/ -f custom-values.yaml
```

## Upgrading a Release

```bash
helm upgrade shop-feed feed/ops/feed/ \
  --set deploy.image.tag=v1.2.4
```

## Uninstalling a Release

```bash
helm uninstall shop-feed --namespace shop
```

## Validating Charts

Before deploying, you can validate the chart:

### Lint the chart
```bash
cd <service>/ops/<service>/
helm dependency update
helm lint .
```

### Dry-run installation
```bash
helm install shop-feed . --dry-run --debug --namespace shop
```

### Template rendering
```bash
helm template shop-feed . --namespace shop
```

## Common Issues and Solutions

### Issue: Dependencies not found
**Error**: `chart directory is missing these dependencies: shortlink-template`

**Solution**: Run `helm dependency update` in the chart directory.

### Issue: Cannot pull dependency
**Error**: `failed to download "oci://ghcr.io/shortlink-org/charts/shortlink-template"`

**Solution**: Ensure you have access to the GitHub Container Registry and that you're authenticated if required.

### Issue: Template errors
**Error**: `template: no template "shortlink-common.ServiceAccount"`

**Solution**: This usually means dependencies weren't downloaded. Run `helm dependency update`.

## Best Practices

1. **Always update dependencies** before installing or upgrading a chart
2. **Use specific image tags** instead of `latest` in production
3. **Enable monitoring** in production environments
4. **Configure resource limits** appropriately for your workload
5. **Use network policies** in production for security
6. **Enable PodDisruptionBudget** for critical services
7. **Test charts** in a development environment before deploying to production
8. **Use separate namespaces** for different environments (dev, staging, production)

## Related Documentation

- [Helm Documentation](https://helm.sh/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [shortlink-org/shortlink Repository](https://github.com/shortlink-org/shortlink)
- [shortlink-template Chart](https://github.com/shortlink-org/shortlink/tree/main/ops/Helm/shortlink-template)

## Contributing

When creating or modifying Helm charts:

1. Follow the existing pattern and structure
2. Use the `shortlink-template` dependency for common resources
3. Document any custom values or configurations in the service's README
4. Test charts with `helm lint` and `helm install --dry-run`
5. Update this documentation if adding new charts or significant features
