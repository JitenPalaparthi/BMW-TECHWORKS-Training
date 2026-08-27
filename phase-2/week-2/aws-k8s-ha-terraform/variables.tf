variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-south-1"
}

variable "cluster_name" {
  description = "Kubernetes cluster name"
  type        = string
  default     = "k8s-ha"
}

variable "vpc_cidr" {
  description = "VPC CIDR"
  type        = string
  default     = "10.0.0.0/16"
}

variable "ssh_allowed_cidr" {
  description = "CIDR allowed to SSH to nodes, preferably YOUR_PUBLIC_IP/32"
  type        = string
}

variable "api_allowed_cidr" {
  description = "CIDR allowed to reach the internet-facing Kubernetes API NLB on port 6443"
  type        = string
}

variable "key_name" {
  description = "Existing AWS EC2 key-pair name"
  type        = string
}

variable "master_instance_type" {
  type    = string
  default = "t3.medium"
}

variable "worker_instance_type" {
  type    = string
  default = "t3.medium"
}

variable "kubernetes_minor_version" {
  description = "Kubernetes package repository minor version, for example v1.36"
  type        = string
  default     = "v1.36"
}
