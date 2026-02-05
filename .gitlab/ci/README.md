# GitLab CI/CD

This directory contains GitLab CI/CD configuration for the Shop project.

## Structure

```
.gitlab/ci/
├── .gitlab-ci.yml           # Main CI configuration
├── stages/
│   └── build.yml            # Build stage with Docker and Helm jobs
├── workflows/
│   ├── matrix_build_docker.yml  # Docker build matrix
│   └── matrix_build_helm.yml    # Helm chart build matrix
├── pipelines/
│   └── docker.yml           # Docker build pipeline
└── README.md
```

## Services

### Docker Images

| Service | Description | Language |
|---------|-------------|----------|
| oms | Order Management System | Go |
| admin | Admin Panel | Python/Django |
| bff | Backend for Frontend | TypeScript |
| pricer | Pricer Service | Go |
| delivery | Delivery Service | Rust |
| ui | Shop Frontend | React/Next.js |

### Helm Charts

| Chart | Path | Description |
|-------|------|-------------|
| common | ops/Helm/common | Common infrastructure |
| temporal | ops/Helm/temporal | Temporal workflow engine |
| oms | oms/ops/oms | OMS deployment |
| admin | admin/ops/Helm | Admin panel deployment |
| bff | bff/ops/Helm | BFF deployment |

## Components

This CI configuration uses shared GitLab components from `shortlink-org/gitlab-templates`:

- `docker_pipeline` - Docker build and push pipeline
- `helm_publish` - Helm chart packaging and publishing
- `common` - Common templates and utilities

## Usage

The pipeline automatically triggers on:
- Push to branches (except `renovate/*` and `*-draft`)
- Merge request events

### Manual Trigger

You can manually trigger specific jobs from the GitLab UI or using the API.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `CI_REGISTRY_IMAGE` | Container registry path |
| `CI_SERVER_FQDN` | GitLab server FQDN |
