# Highly Available Kubernetes Infrastructure on AWS with Terraform

This project creates AWS infrastructure for a **kubeadm-based Kubernetes cluster** with **3 control-plane (master) EC2 instances**, **2 worker EC2 instances**, and an **internet-facing Network Load Balancer (NLB)** in front of the Kubernetes API servers.

> Terraform creates the AWS infrastructure. Kubernetes itself is installed/configured afterward with kubeadm. The scripts intentionally keep bootstrap separate so you can learn and control the Kubernetes installation process.

## Architecture

```text
                         Internet / kubectl
                                |
                         NLB : TCP/6443
                                |
                   Kubernetes API Target Group
                      /          |          \
                 master-1    master-2    master-3
                     |            |           |
                     +------------+-----------+
                                  |
                           Kubernetes network
                            /             \
                       worker-1         worker-2
```

The five EC2 nodes are distributed across two Availability Zones and public subnets for a training/lab-friendly setup. The NLB forwards TCP/6443 to all three control-plane nodes. EC2 instances use Ubuntu 24.04, encrypted gp3 root volumes, IMDSv2, and an IAM instance profile with AWS Systems Manager access.

## Files

- `versions.tf` — Terraform and AWS provider requirements.
- `variables.tf` — configurable inputs.
- `main.tf` — VPC, networking, security group, IAM, EC2, NLB, target group, attachments, and listener.
- `outputs.tf` — NLB endpoint and node addresses.
- `terraform.tfvars.example` — sample configuration.

## Prerequisites

Install Terraform, AWS CLI, and kubectl on your administration machine. Configure AWS credentials with an IAM identity allowed to create VPC, EC2, ELBv2, IAM role/profile, and related resources. If you want SSH access, create an EC2 key pair in the target region first.

```bash
aws configure
aws sts get-caller-identity
terraform version
```

## Step 1 — Configure variables

```bash
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars`. Most importantly, replace `admin_cidr` with your public IP in `/32` form and set `key_name` to an existing EC2 key pair. Do **not** leave SSH open to `0.0.0.0/0` outside a temporary lab.

## Step 2 — Initialize Terraform

```bash
terraform init
terraform fmt -recursive
terraform validate
```

## Step 3 — Review the execution plan

```bash
terraform plan -out=tfplan
```

Review all resources before applying.

## Step 4 — Create the infrastructure

```bash
terraform apply tfplan
```

Then inspect the generated addresses:

```bash
terraform output
terraform output -raw nlb_dns_name
```

At this stage AWS infrastructure exists, but the NLB targets will initially be unhealthy because nothing is listening on TCP/6443 until kube-apiserver starts.

## Step 5 — Prepare all five nodes

SSH to each node (or use AWS Systems Manager) and perform the Kubernetes prerequisites. On **every master and worker**, disable swap, load the required kernel modules, configure sysctl networking, install a CRI runtime such as containerd, and install kubelet/kubeadm/kubectl at mutually compatible Kubernetes versions.

Typical OS preparation includes:

```bash
sudo swapoff -a
sudo modprobe overlay
sudo modprobe br_netfilter
```

Configure sysctl values required for bridged Kubernetes networking and IPv4 forwarding, then install and configure containerd with the systemd cgroup driver. Install kubeadm, kubelet, and kubectl using the official Kubernetes package repository for the Kubernetes minor version you choose.

## Step 6 — Initialize master-1

Get the NLB DNS name:

```bash
NLB_DNS=$(terraform output -raw nlb_dns_name)
echo "$NLB_DNS"
```

On `master-1`, initialize the cluster using the NLB as the stable control-plane endpoint. Choose a pod CIDR compatible with your selected CNI. For example, a common lab value is `10.244.0.0/16`.

```bash
sudo kubeadm init \
  --control-plane-endpoint "<NLB-DNS>:6443" \
  --upload-certs \
  --pod-network-cidr=10.244.0.0/16
```

**Important:** replace `<NLB-DNS>` with the Terraform output. Save both commands printed by kubeadm: the control-plane join command and the worker join command. The certificate key from `--upload-certs` is sensitive and time-limited.

Configure kubectl for the administrative user on master-1 using the commands printed by kubeadm.

## Step 7 — Install a CNI plugin

Install exactly one supported Container Network Interface implementation, such as Calico, Cilium, or Flannel. Follow that CNI project's current installation documentation and ensure its pod CIDR/network configuration matches your kubeadm configuration where applicable.

Do not expect nodes to become fully `Ready` until a functioning CNI is installed.

## Step 8 — Join master-2 and master-3

Run the **control-plane join command generated by kubeadm** on each additional master. Conceptually it looks like:

```bash
sudo kubeadm join <NLB-DNS>:6443 \
  --token <TOKEN> \
  --discovery-token-ca-cert-hash sha256:<HASH> \
  --control-plane \
  --certificate-key <CERTIFICATE-KEY>
```

Do not copy the placeholders literally. Use values generated by your cluster.

After the API servers start on all masters, the NLB target group should report all three master instances healthy on TCP/6443.

## Step 9 — Join worker-1 and worker-2

Run the kubeadm **worker join command** on both workers:

```bash
sudo kubeadm join <NLB-DNS>:6443 \
  --token <TOKEN> \
  --discovery-token-ca-cert-hash sha256:<HASH>
```

Again, use the real values produced by kubeadm.

## Step 10 — Verify the cluster

From a machine with a valid cluster kubeconfig:

```bash
kubectl get nodes -o wide
kubectl get pods -A
```

Expected topology:

```text
master-1   control-plane
master-2   control-plane
master-3   control-plane
worker-1   worker
worker-2   worker
```

You can also verify the AWS target health:

```bash
aws elbv2 describe-target-health \
  --target-group-arn "$(terraform output -json | jq -r '. // empty' 2>/dev/null)"
```

For target health, it is usually simpler to inspect EC2 Console -> Target Groups -> the Kubernetes API target group, or obtain the target-group ARN with Terraform console/state.

## Security-group design

The node security group permits SSH from `admin_cidr`, Kubernetes API traffic on TCP/6443 from the VPC, etcd peer/client traffic on TCP/2379-2380 between cluster nodes, kubelet TCP/10250, control-plane TCP/10257-10259, NodePort TCP/UDP 30000-32767 inside the VPC, and node-to-node TCP/UDP/ICMP. Outbound traffic is allowed so nodes can obtain packages/images and communicate externally.

This is suitable for a **training environment**, not a final production security policy. Production clusters should use private subnets for nodes, controlled egress/NAT or VPC endpoints, narrowly scoped security-group rules, private management access, logging/monitoring, backups, and hardened IAM.

## Important HA behavior

The NLB is a TCP Layer-4 load balancer. It does not terminate Kubernetes API TLS; TLS terminates at kube-apiserver. The NLB DNS name is supplied to `kubeadm init` as `control-plane-endpoint`, giving clients and joining nodes one stable endpoint while the three control-plane nodes provide API-server redundancy.

A highly available control plane also depends on the stacked etcd members that kubeadm creates on the three control-plane nodes. Protect and back up etcd appropriately in real environments.

## Destroy the lab

When finished:

```bash
terraform destroy
```

Review the destroy plan before confirming. This permanently removes the Terraform-managed infrastructure.

## Production improvements

For production, place EC2 nodes in private subnets across three Availability Zones, restrict the API endpoint to trusted networks where possible, use SSM/bastion/VPN rather than public SSH, add NAT gateways or VPC endpoints, implement explicit CNI-specific security requirements, enable observability and audit logging, automate OS/Kubernetes bootstrap, configure etcd backup/restore procedures, and use remote Terraform state with locking and encryption.


aws ec2 create-key-pair \
  --key-name k8s-cluster-key \
  --region ap-south-1 \
  --query 'KeyMaterial' \
  --output text > k8s-cluster-key.pem

  chmod 400 k8s-cluster-key.pem
  

  key_name = "k8s-cluster-key"
