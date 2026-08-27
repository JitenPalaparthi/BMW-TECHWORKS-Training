output "nlb_dns_name" {
  description = "Public DNS name of the internet-facing Kubernetes API NLB"
  value       = aws_lb.k8s_api.dns_name
}

output "kubernetes_api_endpoint" {
  description = "Kubernetes API endpoint"
  value       = "https://${aws_lb.k8s_api.dns_name}:6443"
}

output "master_public_ips" {
  value = aws_instance.master[*].public_ip
}

output "master_private_ips" {
  value = aws_instance.master[*].private_ip
}

output "worker_public_ips" {
  value = aws_instance.worker[*].public_ip
}

output "worker_private_ips" {
  value = aws_instance.worker[*].private_ip
}
