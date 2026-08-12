# Java Demo Application Stack on Docker Desktop Kubernetes

This directory contains Kubernetes manifests for PostgreSQL 17, a Java Spring Boot backend, and Adminer.

## Quick Start

From inside this `k8s/` directory:

```bash
kubectl apply -k .
kubectl get all -n java-demo
```

Adminer port-forward:

```bash
kubectl port-forward svc/adminer-service -n java-demo 8080:8080
```

## Deployment strategy lab

See [`DEPLOYMENT-STRATEGIES.md`](DEPLOYMENT-STRATEGIES.md) for:

- RollingUpdate rollout status and image updates
- `kubernetes.io/change-cause` annotations
- rollout history and revision inspection
- `kubectl diff`
- rollback to previous or specific revisions
- Blue/Green deployment and instant Service-selector rollback
- Canary deployment using stable/canary replica ratios
- custom release-audit annotations

The Blue/Green and Canary manifests live under `examples/` and are intentionally not included in the root `kustomization.yaml`; run them separately as labs.
