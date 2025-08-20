<!-- PROJECT SHIELDS -->
<a id="readme-top"></a>

<!-- PROJECT TITLE -->
<div align="center">
  <h3 align="center">🚀 Orion Platform</h3>
  <p align="center">
    A cloud‑native developer platform built with Go and Kubernetes. Declaratively deploy full‑stack applications with one YAML while the operator provisions the right infrastructure for each environment.
    <br/>
    <a href="https://github.com/virtual457/Orion-platform"><strong>Explore the docs »</strong></a>
    <br/><br/>
    <a href="https://github.com/virtual457/Orion-platform/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    ·
    <a href="https://github.com/virtual457/Orion-platform/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>

## About The Project

Orion Platform is a Kubernetes operator and set of resources that streamline the complete application lifecycle:

- Single‑file application definition using a custom `Application` resource
- Smart environment selection (local containers vs. cloud services)
- Automated provisioning for PostgreSQL, Redis, and S3‑compatible storage
- Status reporting, health checks, and lifecycle management

This project focuses on clarity, reliability, and an approachable developer experience for platform engineering.

### Key Features

- ✅ Custom Resource Definition for applications
- ✅ Controller with event‑driven reconciliation loop
- ✅ Environment‑aware infrastructure provisioning
- ✅ Persistent storage for stateful services
- ✅ Service discovery and configuration injection
- ✅ Ownership and automatic cleanup of dependent resources

## Architecture

```
Developer YAML → Kubernetes API → Orion Controller → Infrastructure Creation
       ↓              ↓                    ↓                    ↓
   kubectl apply → etcd storage → Event notification → PostgreSQL/Redis/MinIO
                                                 → Application Deployment
                                                 → Services
                                                 → Status update
```

### Components

- Custom Resource: `Application` (image, replicas, infra requirements)
- Controller: Reconciles desired state, provisions infra, deploys app
- Infrastructure Layer: Local (containers) and cloud‑ready integration points

## Getting Started

### Prerequisites
- Go 1.21+
- Docker
- A Kubernetes cluster (Kind/Minikube or remote)
- kubectl configured to your cluster

### Build & Run
```bash
# Build the operator
make build

# Run against your current kubeconfig
make run

# Apply the CRD
kubectl apply -f config/crd/application-crd.yaml

# Deploy the controller (manifests)
kubectl apply -f deploy/controller.yaml
```

### Example: Minimal Application
```yaml
apiVersion: platform.orion.dev/v1alpha1
kind: Application
metadata:
  name: simple-nginx
spec:
  image: nginx:latest
  replicas: 3
```

## Roadmap

- [x] Project and CRD setup
- [x] Core reconciliation loop
- [x] Local infra provisioning (PostgreSQL/Redis/MinIO)
- [ ] Cloud integrations (RDS/ElastiCache/S3)
- [ ] Observability (metrics/dashboards)
- [ ] CI/CD and testing

## Built With

- Go (controller‑runtime)
- Kubernetes (CRDs, RBAC, Deployments/StatefulSets/Services)
- Docker / Kind / Minikube

## Contributing

Contributions are welcome! Please open an issue to discuss changes or submit a PR following conventional guidelines.

## License

Distributed under the MIT License. See `LICENSE` for details.

## Contact

Chandan Gowda K S – gowdakeelarashivan.c@northeastern.edu

Project Link: https://github.com/virtual457/Orion-platform
