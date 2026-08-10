kubectl create nd demo-ns

kubectl apply -f nginx-pod.yaml

kubectl apply -f nginx-service.yaml

kubectl port-forward svc/nginx-service -n demo-ns 8080:8080


kubectl config current-context 
kubectl config get-contexts


kubectl delete ns demo-ns

