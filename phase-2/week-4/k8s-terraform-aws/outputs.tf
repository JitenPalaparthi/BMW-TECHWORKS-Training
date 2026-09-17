output "ubuntu_ami_id" {
  description = "Ubuntu AMI used by the cluster"
  value       = data.aws_ami.ubuntu.id
}

output "vpc_id" {
  description = "Kubernetes VPC ID"
  value       = aws_vpc.k8s.id
}

output "master_public_ips" {
  description = "Public IP addresses of control plane nodes"
  value       = aws_instance.master[*].public_ip
}

output "master_private_ips" {
  description = "Private IP addresses of control plane nodes"
  value       = aws_instance.master[*].private_ip
}

output "worker_public_ips" {
  description = "Public IP addresses of worker nodes"
  value       = aws_instance.worker[*].public_ip
}

output "worker_private_ips" {
  description = "Private IP addresses of worker nodes"
  value       = aws_instance.worker[*].private_ip
}

output "nlb_dns_name" {
  description = "DNS name of Kubernetes API Network Load Balancer"
  value       = aws_lb.k8s_api.dns_name
}

output "kubernetes_api_endpoint" {
  description = "Kubernetes API endpoint to use with kubeadm"
  value       = "${aws_lb.k8s_api.dns_name}:6443"
}