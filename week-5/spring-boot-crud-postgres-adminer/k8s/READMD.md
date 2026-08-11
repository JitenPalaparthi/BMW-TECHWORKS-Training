# 1. Create the namespace if it doesn't exist
kubectl create namespace java-demo --dry-run=client -o yaml | kubectl apply -f -

# 2. Apply Configs, Secrets, and Storage
kubectl apply -f postgres-config.yaml -f postgres-pvc.yaml -f java-config.yaml

# 3. Apply PostgreSQL Workload & Service
kubectl apply -f postgres-deployment.yaml -f postgres-service.yaml

# 4. Apply Java App Workload & Service
kubectl apply -f java-deployment.yaml -f java-service.yaml

# 5. Apply Adminer Deployment and Service 
kubectl apply -f adminer-deployment.yaml

# 6. Port forward
kubectl port-forward svc/adminer-service -n java-demo 8080:8080

# Java Demo Application Stack on Docker Desktop Kubernetes

This directory contains the Kubernetes manifests to run PostgreSQL 17, a Java Spring Boot backend, and Adminer DB web client inside Docker Desktop Kubernetes on macOS.

## Quick Start (Deploying Everything)

Run Kustomize directly through `kubectl` from inside the `k8s/` directory:

```bash
kubectl apply -k .