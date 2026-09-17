provider "aws" {
  region = var.aws_region
}

locals {
  common_tags = {
    Project   = var.cluster_name
    ManagedBy = "Terraform"
  }
}

# -------------------------------------------------------
# Availability Zones
# -------------------------------------------------------

data "aws_availability_zones" "available" {
  state = "available"
}

# -------------------------------------------------------
# Ubuntu AMI
# -------------------------------------------------------

data "aws_ami" "ubuntu" {
  most_recent = true

  owners = ["099720109477"]

  filter {
    name = "name"

    values = [
      "ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"
    ]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }
}

# -------------------------------------------------------
# VPC
# -------------------------------------------------------

resource "aws_vpc" "k8s" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-vpc"
    }
  )
}

# -------------------------------------------------------
# Internet Gateway
# -------------------------------------------------------

resource "aws_internet_gateway" "k8s" {
  vpc_id = aws_vpc.k8s.id

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-igw"
    }
  )
}

# -------------------------------------------------------
# Public Subnet 1
# -------------------------------------------------------

resource "aws_subnet" "public_1" {
  vpc_id                  = aws_vpc.k8s.id
  cidr_block              = var.public_subnet_1_cidr
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-public-subnet-1"
    }
  )
}

# -------------------------------------------------------
# Public Subnet 2
# -------------------------------------------------------

resource "aws_subnet" "public_2" {
  vpc_id                  = aws_vpc.k8s.id
  cidr_block              = var.public_subnet_2_cidr
  availability_zone       = data.aws_availability_zones.available.names[1]
  map_public_ip_on_launch = true

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-public-subnet-2"
    }
  )
}

# -------------------------------------------------------
# Route Table
# -------------------------------------------------------

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.k8s.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.k8s.id
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-public-route-table"
    }
  )
}

# -------------------------------------------------------
# Route Table Associations
# -------------------------------------------------------

resource "aws_route_table_association" "public_1" {
  subnet_id      = aws_subnet.public_1.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_2" {
  subnet_id      = aws_subnet.public_2.id
  route_table_id = aws_route_table.public.id
}

# -------------------------------------------------------
# Security Group
# -------------------------------------------------------

resource "aws_security_group" "k8s_nodes" {
  name        = "${var.cluster_name}-nodes-sg"
  description = "Security group for Kubernetes control plane and worker nodes"
  vpc_id      = aws_vpc.k8s.id

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-nodes-sg"
    }
  )
}

# -------------------------------------------------------
# Allow SSH
# -------------------------------------------------------

resource "aws_vpc_security_group_ingress_rule" "ssh" {
  security_group_id = aws_security_group.k8s_nodes.id

  cidr_ipv4   = var.allowed_ssh_cidr
  from_port   = 22
  to_port     = 22
  ip_protocol = "tcp"

  description = "SSH access"
}

# -------------------------------------------------------
# Kubernetes API Server
# -------------------------------------------------------

resource "aws_vpc_security_group_ingress_rule" "k8s_api" {
  security_group_id = aws_security_group.k8s_nodes.id

  cidr_ipv4   = var.allowed_k8s_api_cidr
  from_port   = 6443
  to_port     = 6443
  ip_protocol = "tcp"

  description = "Kubernetes API server"
}

# -------------------------------------------------------
# Internal Kubernetes Communication
# -------------------------------------------------------

resource "aws_vpc_security_group_ingress_rule" "internal" {
  security_group_id = aws_security_group.k8s_nodes.id

  referenced_security_group_id = aws_security_group.k8s_nodes.id
  ip_protocol                  = "-1"

  description = "Allow all communication between Kubernetes nodes"
}

# -------------------------------------------------------
# NodePort Services
# -------------------------------------------------------

resource "aws_vpc_security_group_ingress_rule" "nodeport" {
  security_group_id = aws_security_group.k8s_nodes.id

  cidr_ipv4   = var.allowed_k8s_api_cidr
  from_port   = 30000
  to_port     = 32767
  ip_protocol = "tcp"

  description = "Kubernetes NodePort services"
}

# -------------------------------------------------------
# Outbound Traffic
# -------------------------------------------------------

resource "aws_vpc_security_group_egress_rule" "all" {
  security_group_id = aws_security_group.k8s_nodes.id

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"

  description = "Allow all outbound traffic"
}

# -------------------------------------------------------
# Control Plane Nodes
# 3 EC2 instances
# -------------------------------------------------------

resource "aws_instance" "master" {
  count = 3

  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type
  key_name      = var.key_name

  subnet_id = count.index % 2 == 0 ? (
    aws_subnet.public_1.id
    ) : (
    aws_subnet.public_2.id
  )

  vpc_security_group_ids = [
    aws_security_group.k8s_nodes.id
  ]

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-master-${count.index + 1}"
      Role = "control-plane"
    }
  )
}

# -------------------------------------------------------
# Worker Nodes
# 2 EC2 instances
# -------------------------------------------------------

resource "aws_instance" "worker" {
  count = 2

  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type
  key_name      = var.key_name

  subnet_id = count.index % 2 == 0 ? (
    aws_subnet.public_1.id
    ) : (
    aws_subnet.public_2.id
  )

  vpc_security_group_ids = [
    aws_security_group.k8s_nodes.id
  ]

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-worker-${count.index + 1}"
      Role = "worker"
    }
  )
}

# =======================================================
# Network Load Balancer
# =======================================================

resource "aws_lb" "k8s_api" {
  name               = "${var.cluster_name}-nlb"
  internal           = false
  load_balancer_type = "network"

  subnets = [
    aws_subnet.public_1.id,
    aws_subnet.public_2.id
  ]

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-nlb"
    }
  )
}

# -------------------------------------------------------
# NLB Target Group
# -------------------------------------------------------

resource "aws_lb_target_group" "k8s_api" {
  name = "${var.cluster_name}-api-tg"

  port     = 6443
  protocol = "TCP"
  vpc_id   = aws_vpc.k8s.id

  target_type = "instance"

  health_check {
    protocol            = "TCP"
    port                = "6443"
    healthy_threshold   = 3
    unhealthy_threshold = 3
    interval            = 30
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${var.cluster_name}-api-tg"
    }
  )
}

# -------------------------------------------------------
# Register Master 1
# -------------------------------------------------------

resource "aws_lb_target_group_attachment" "master_1" {
  target_group_arn = aws_lb_target_group.k8s_api.arn
  target_id        = aws_instance.master[0].id
  port             = 6443
}

# -------------------------------------------------------
# Register Master 2
# -------------------------------------------------------

resource "aws_lb_target_group_attachment" "master_2" {
  target_group_arn = aws_lb_target_group.k8s_api.arn
  target_id        = aws_instance.master[1].id
  port             = 6443
}

# -------------------------------------------------------
# Register Master 3
# -------------------------------------------------------

resource "aws_lb_target_group_attachment" "master_3" {
  target_group_arn = aws_lb_target_group.k8s_api.arn
  target_id        = aws_instance.master[2].id
  port             = 6443
}

# -------------------------------------------------------
# NLB Listener
# -------------------------------------------------------

resource "aws_lb_listener" "k8s_api" {
  load_balancer_arn = aws_lb.k8s_api.arn

  port     = 6443
  protocol = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.k8s_api.arn
  }
}