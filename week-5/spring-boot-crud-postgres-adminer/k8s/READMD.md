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


# Record why you are making the change
kubectl annotate deployment/java-demo-app -n java-demo \
  kubernetes.io/change-cause="Upgrade Java app from v0.0.2 to v0.0.3" \
  --overwrite

# Change version
kubectl set image deployment/java-demo-app -n java-demo \
  java-demo=jpalaparthi/bmw-k8s-java-demo:v0.0.3

# Watch rollout
kubectl rollout status deployment/java-demo-app -n java-demo

# See all revisions and CHANGE-CAUSE
kubectl rollout history deployment/java-demo-app -n java-demo

# Inspect one particular version/revision
kubectl rollout history deployment/java-demo-app \
  -n java-demo --revision=2

# Roll back one version
kubectl rollout undo deployment/java-demo-app -n java-demo

# Roll back to an exact revision
kubectl rollout undo deployment/java-demo-app \
  -n java-demo --to-revision=2

# See what your YAML would change BEFORE applying
kubectl diff -f 05-java-deployment.yaml

# Or compare the complete Kustomize application
kubectl diff -k .