# Kubernetes Deployment Strategies and Rollback Lab

Namespace used below: `java-demo`.

## 1. Standard Deployment: inspect, change, and track a release

Check the current image and rollout status:

```bash
kubectl get deploy java-demo-app -n java-demo
kubectl get deploy java-demo-app -n java-demo -o wide
kubectl rollout status deployment/java-demo-app -n java-demo
kubectl get deploy java-demo-app -n java-demo -o jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
```

Before changing the image, attach a human-readable change cause and then update the Pod template:

```bash
kubectl annotate deployment/java-demo-app -n java-demo \
  kubernetes.io/change-cause="Upgrade Java app from v0.0.2 to v0.0.3" --overwrite

kubectl set image deployment/java-demo-app -n java-demo \
  java-demo=jpalaparthi/bmw-k8s-java-demo:v0.0.3

kubectl rollout status deployment/java-demo-app -n java-demo
```

`kubernetes.io/change-cause` is not automatically populated for you. Adding it explicitly makes `kubectl rollout history` much more useful.

## 2. See what changed between versions

List revisions and their recorded change causes:

```bash
kubectl rollout history deployment/java-demo-app -n java-demo
```

Inspect one specific revision:

```bash
kubectl rollout history deployment/java-demo-app -n java-demo --revision=2
```

See the current change-cause annotation:

```bash
kubectl get deployment/java-demo-app -n java-demo \
  -o jsonpath='{.metadata.annotations.kubernetes\.io/change-cause}{"\n"}'
```

See Kubernetes-managed ReplicaSet revision annotations:

```bash
kubectl get rs -n java-demo -l app=java-demo-app \
  -o custom-columns='NAME:.metadata.name,REVISION:.metadata.annotations.deployment\.kubernetes\.io/revision,IMAGE:.spec.template.spec.containers[0].image'
```

Preview local YAML changes against the live cluster before applying:

```bash
kubectl diff -f 05-java-deployment.yaml
```

Or preview the entire Kustomize directory:

```bash
kubectl diff -k .
```

Useful live inspection commands:

```bash
kubectl describe deployment java-demo-app -n java-demo
kubectl get deployment java-demo-app -n java-demo -o yaml
kubectl get rs -n java-demo -l app=java-demo-app
kubectl get pods -n java-demo -l app=java-demo-app --show-labels
```

## 3. Rollback

Rollback to the immediately previous Deployment revision:

```bash
kubectl rollout undo deployment/java-demo-app -n java-demo
kubectl rollout status deployment/java-demo-app -n java-demo
```

Rollback to a particular revision:

```bash
kubectl rollout history deployment/java-demo-app -n java-demo
kubectl rollout undo deployment/java-demo-app -n java-demo --to-revision=2
kubectl rollout status deployment/java-demo-app -n java-demo
```

Pause / resume an in-progress Deployment rollout:

```bash
kubectl rollout pause deployment/java-demo-app -n java-demo
kubectl rollout resume deployment/java-demo-app -n java-demo
```

Restart without changing the image tag:

```bash
kubectl rollout restart deployment/java-demo-app -n java-demo
kubectl rollout status deployment/java-demo-app -n java-demo
```

## 4. Blue/Green Deployment

Blue = current production version. Green = new candidate version. The Service selects only one color at a time.

Deploy both environments:

```bash
kubectl apply -f examples/blue-green/blue-deployment.yaml
kubectl apply -f examples/blue-green/green-deployment.yaml
kubectl apply -f examples/blue-green/service.yaml
kubectl get pods -n java-demo -l app=java-demo-app --show-labels
```

Initially the Service selects `version=blue`:

```bash
kubectl get svc java-demo-bg-service -n java-demo -o jsonpath='{.spec.selector}{"\n"}'
kubectl get endpoints java-demo-bg-service -n java-demo
```

Test Green directly before production cutover:

```bash
kubectl port-forward deployment/java-demo-green -n java-demo 18080:8080
```

Switch 100% of Service traffic from Blue to Green:

```bash
kubectl patch service java-demo-bg-service -n java-demo \
  -p '{"spec":{"selector":{"app":"java-demo-app","version":"green"}}}'
```

Verify:

```bash
kubectl get svc java-demo-bg-service -n java-demo -o jsonpath='{.spec.selector}{"\n"}'
kubectl get endpoints java-demo-bg-service -n java-demo
```

Immediate Blue/Green rollback: switch the Service selector back to Blue:

```bash
kubectl patch service java-demo-bg-service -n java-demo \
  -p '{"spec":{"selector":{"app":"java-demo-app","version":"blue"}}}'
```

After Green has been stable long enough, Blue can be scaled down instead of deleted immediately:

```bash
kubectl scale deployment java-demo-blue -n java-demo --replicas=0
```

## 5. Canary Deployment with plain Kubernetes

This example uses 4 stable Pods and 1 canary Pod. Both groups match the same Service selector, so the canary gets an approximate share of connections. This is not precise percentage-based routing; use an ingress/service mesh/Gateway implementation when exact weighted traffic is required.

Deploy it:

```bash
kubectl apply -f examples/canary/stable-deployment.yaml
kubectl apply -f examples/canary/canary-deployment.yaml
kubectl apply -f examples/canary/service.yaml
kubectl get pods -n java-demo -l app=java-demo-canary --show-labels
```

Approximate traffic ratios by changing replica counts:

```bash
# 4 stable + 1 canary => roughly 20% of endpoints are canary
kubectl scale deployment/java-demo-stable -n java-demo --replicas=4
kubectl scale deployment/java-demo-canary -n java-demo --replicas=1

# 3 stable + 2 canary => roughly 40% of endpoints are canary
kubectl scale deployment/java-demo-stable -n java-demo --replicas=3
kubectl scale deployment/java-demo-canary -n java-demo --replicas=2

# Promote candidate: all Pods run the new version
kubectl scale deployment/java-demo-stable -n java-demo --replicas=0
kubectl scale deployment/java-demo-canary -n java-demo --replicas=5
```

Abort the Canary immediately:

```bash
kubectl scale deployment/java-demo-canary -n java-demo --replicas=0
kubectl scale deployment/java-demo-stable -n java-demo --replicas=5
```

Update the canary image while recording why:

```bash
kubectl annotate deployment/java-demo-canary -n java-demo \
  kubernetes.io/change-cause="Canary test of v0.0.4" --overwrite
kubectl set image deployment/java-demo-canary -n java-demo \
  java-demo=jpalaparthi/bmw-k8s-java-demo:v0.0.4
kubectl rollout status deployment/java-demo-canary -n java-demo
kubectl rollout history deployment/java-demo-canary -n java-demo
```

## 6. Useful annotations for release auditing

You can attach your own release metadata as annotations:

```bash
kubectl annotate deployment/java-demo-app -n java-demo \
  release.example.com/version="v0.0.3" \
  release.example.com/git-commit="abc1234" \
  release.example.com/ticket="APP-142" \
  release.example.com/deployed-by="jiten" \
  --overwrite
```

Inspect all annotations:

```bash
kubectl get deployment/java-demo-app -n java-demo -o jsonpath='{.metadata.annotations}'
echo
```

Remove an annotation by appending `-` to its key:

```bash
kubectl annotate deployment/java-demo-app -n java-demo release.example.com/ticket-
```

## 7. Recommended training flow

1. Deploy v0.0.2.
2. Annotate the intended change.
3. Change to v0.0.3 with `kubectl set image`.
4. Check `rollout status` and `rollout history`.
5. Inspect a revision.
6. Roll back to the previous revision.
7. Run the Blue/Green example and switch the Service selector.
8. Roll it back by switching the selector to Blue.
9. Run the Canary example at 4:1, then 3:2, then promote or abort.
10. Use `kubectl diff` before future applies.
