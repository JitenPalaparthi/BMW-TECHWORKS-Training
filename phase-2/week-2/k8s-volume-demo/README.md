# Go Kubernetes Volume Demo

## Run locally

go run .

Open http://localhost:8080

Upload with curl:

curl -F "file=@example.png" http://localhost:8080/upload

Fetch:

curl http://localhost:8080/files/example.png --output example-downloaded.png

## Docker

docker build -t go-volume-demo:1.0 .
docker run --rm -p 8080:8080 -v demo-data:/data go-volume-demo:1.0

## Minikube

Build directly into Minikube's Docker daemon:

eval $(minikube docker-env)
docker build -t go-volume-demo:1.0 .

Apply one manifest at a time:

kubectl apply -f k8s/01-emptydir.yaml
kubectl apply -f k8s/02-hostpath.yaml
kubectl apply -f k8s/03-pvc.yaml

Get URL, for example:

minikube service file-demo-emptydir --url

## What to demonstrate

- emptyDir: survives container restart inside the same Pod, but is deleted when the Pod is replaced/deleted.
- hostPath: data is stored on the Kubernetes node filesystem and can survive Pod recreation on that same node.
- PVC: storage lifecycle is decoupled from the Pod and is the normal persistent-storage model.
