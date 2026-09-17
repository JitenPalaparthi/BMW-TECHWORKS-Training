variable "aws_region" {
  description = "AWS region where the Kubernetes cluster will be created"
  type        = string
  default     = "ap-south-1"
}

variable "cluster_name" {
  description = "Kubernetes cluster name"
  type        = string
  default     = "k8s-cluster"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_1_cidr" {
  description = "CIDR for public subnet 1"
  type        = string
  default     = "10.0.1.0/24"
}

variable "public_subnet_2_cidr" {
  description = "CIDR for public subnet 2"
  type        = string
  default     = "10.0.2.0/24"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.medium"
}

variable "key_name" {
  description = "Existing AWS EC2 Key Pair name"
  type        = string
}

variable "allowed_ssh_cidr" {
  description = "CIDR allowed to SSH into Kubernetes nodes"
  type        = string
  default     = "0.0.0.0/0"
}

variable "allowed_k8s_api_cidr" {
  description = "CIDR allowed to access Kubernetes API server"
  type        = string
  default     = "0.0.0.0/0"
}