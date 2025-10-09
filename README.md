<!-- PROJECT SHIELDS -->
<!-- *** I'm using markdown "reference style" links for readability. *** Reference links are enclosed in brackets [ ] instead of parentheses ( ). *** See the bottom of this document for the declaration of the reference variables *** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use. *** https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]

<a id="readme-top"></a>

<!-- PROJECT TITLE -->
<div align="center">
  <h3 align="center">🚀 Orion Platform</h3>
  <p align="center">
    A cloud‑native developer platform built with Go and Kubernetes. Declaratively deploy full‑stack applications with one YAML while the operator provisions the right infrastructure for each environment.
    <br/>
    <a href="https://github.com/virtual457/Orion-platform"><strong>Explore the docs »</strong></a>
    <br/><br/>
    <a href="https://github.com/virtual457/Orion-platform">View Demo</a>
    ·
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

<p align="right">(<a href="#readme-top">back to top</a>)</p>

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

<p align="right">(<a href="#readme-top">back to top</a>)</p>

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

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Built With

- Go (controller‑runtime)
- Kubernetes (CRDs, RBAC, Deployments/StatefulSets/Services)
- Docker / Kind / Minikube

## Contributing

Contributions are welcome! Please open an issue to discuss changes or submit a PR following conventional guidelines.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## License

Distributed under the MIT License. See `LICENSE` for details.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Contact

Chandan Gowda K S – chandan.keelara@gmail.com

Project Link: https://github.com/virtual457/Orion-platform

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/virtual457/Orion-platform.svg?style=for-the-badge
[forks-shield]: https://img.shields.io/github/forks/virtual457/Orion-platform.svg?style=for-the-badge
[stars-shield]: https://img.shields.io/github/stars/virtual457/Orion-platform.svg?style=for-the-badge
[issues-shield]: https://img.shields.io/github/issues/virtual457/Orion-platform.svg?style=for-the-badge
[license-shield]: https://img.shields.io/github/license/virtual457/Orion-platform.svg?style=for-the-badge
[contributors-url]: https://github.com/virtual457/Orion-platform/graphs/contributors
[forks-url]: https://github.com/virtual457/Orion-platform/network/members
[stars-url]: https://github.com/virtual457/Orion-platform/stargazers
[issues-url]: https://github.com/virtual457/Orion-platform/issues
[license-url]: https://github.com/virtual457/Orion-platform/blob/master/LICENSE
