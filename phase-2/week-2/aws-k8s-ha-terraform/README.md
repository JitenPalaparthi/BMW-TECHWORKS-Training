# AWS HA Kubernetes Cluster with Terraform and kubeadm

This project provisions the AWS infrastructure for a **high-availability kubeadm Kubernetes cluster** with:

- 3 control-plane EC2 instances
- 3 worker EC2 instances
- 3 public subnets across 3 Availability Zones
- Internet Gateway and public routing
- Internet-facing AWS Network Load Balancer
- TCP listener on port `6443`
- NLB target group containing only the 3 control-plane nodes
- Security groups for the NLB, control planes, and workers
- Ubuntu 24.04
- containerd
- kubeadm
- kubelet
- kubectl

Terraform creates the infrastructure. `kubeadm` creates the Kubernetes cluster.

---

## Architecture

```text
                           Internet
                              |
                              | TCP 6443
                              v
                  Internet-facing AWS NLB
                         :6443
                              |
                 +------------+------------+
                 |            |            |
                 v            v            v
              Master-1     Master-2     Master-3
              API Server   API Server   API Server
              etcd         etcd         etcd
                 \            |            /
                  \___________|___________/
                              |
                         Kubernetes
                              |
                 +------------+------------+
                 |            |            |
               Worker-1     Worker-2     Worker-3
```

The NLB itself is public because:

```hcl
internal = false
```

The NLB listens on TCP `6443`.

Only the CIDR configured in:

```hcl
api_allowed_cidr
```

can access the Kubernetes API.

For a lab you can technically set:

```hcl
api_allowed_cidr = "0.0.0.0/0"
```

but that exposes the Kubernetes API to the entire Internet and is not recommended.

Use your public IP instead:

```hcl
api_allowed_cidr = "49.37.10.20/32"
```

---

# 1. Prerequisites

Install:

- Terraform
- AWS CLI
- An AWS account
- An EC2 key pair

Check Terraform:

```bash
terraform version
```

Check AWS CLI:

```bash
aws --version
```

Configure AWS credentials:

```bash
aws configure
```

Verify authentication:

```bash
aws sts get-caller-identity
```

---

# 2. Create an EC2 key pair

If you already have one, skip this section.

AWS Console:

```text
EC2
  -> Network & Security
  -> Key Pairs
  -> Create key pair
```

Download the `.pem` file.

On macOS/Linux:

```bash
chmod 400 my-aws-key.pem
```

The Terraform variable must contain the AWS key-pair **name**, not the local filename.

Example:

```hcl
key_name = "my-aws-key"
```

---

# 3. Configure Terraform variables

Copy:

```bash
cp terraform.tfvars.example terraform.tfvars
```

Find your public IP:

```bash
curl ifconfig.me
```

Suppose the result is:

```text
49.37.10.20
```

Then edit `terraform.tfvars`:

```hcl
aws_region   = "ap-south-1"
cluster_name = "my-k8s-cluster"

key_name = "my-aws-key"

ssh_allowed_cidr = "49.37.10.20/32"
api_allowed_cidr = "49.37.10.20/32"

master_instance_type = "t3.medium"
worker_instance_type = "t3.medium"

kubernetes_minor_version = "v1.36"
```

---

# 4. Initialize Terraform

```bash
terraform init
```

Format:

```bash
terraform fmt -recursive
```

Validate:

```bash
terraform validate
```

Plan:

```bash
terraform plan
```

Create the infrastructure:

```bash
terraform apply
```

Enter:

```text
yes
```

---

# 5. View Terraform outputs

```bash
terraform output
```

Important output:

```bash
terraform output -raw nlb_dns_name
```

Example:

```text
my-k8s-cluster-nlb-123456.elb.ap-south-1.amazonaws.com
```

Also view node IPs:

```bash
terraform output master_public_ips
terraform output worker_public_ips
```

---

# 6. Why the NLB target group initially shows Unhealthy

Immediately after Terraform completes, the NLB targets may be:

```text
Unhealthy
```

This is expected.

The NLB checks:

```text
master-1:6443
master-2:6443
master-3:6443
```

But kube-apiserver is not running yet.

After `kubeadm init` and the other masters join, port `6443` starts listening and the targets become healthy.

---

# 7. SSH to Master-1

```bash
ssh -i my-aws-key.pem ubuntu@MASTER_1_PUBLIC_IP
```

Verify packages:

```bash
kubeadm version
kubectl version --client
containerd --version
```

---

# 8. Initialize Master-1

Get the NLB DNS name from your laptop:

```bash
terraform output -raw nlb_dns_name
```

Use that DNS name as kubeadm's control-plane endpoint.

On Master-1:

```bash
sudo kubeadm init \
  --control-plane-endpoint "NLB_DNS_NAME:6443" \
  --upload-certs \
  --pod-network-cidr=192.168.0.0/16
```

Example:

```bash
sudo kubeadm init \
  --control-plane-endpoint \
  "my-k8s-cluster-nlb-123456.elb.ap-south-1.amazonaws.com:6443" \
  --upload-certs \
  --pod-network-cidr=192.168.0.0/16
```

The NLB DNS name is critical. All control-plane nodes and workers should join through this endpoint.

---

# 9. Configure kubectl on Master-1

```bash
mkdir -p $HOME/.kube

sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config

sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

Check:

```bash
kubectl get nodes
```

The node will initially be `NotReady` until a CNI plugin is installed.

---

# 10. Install a CNI

This project uses:

```text
192.168.0.0/16
```

as the Pod CIDR.

Install a CNI that supports your chosen Kubernetes version.

For example, install Calico using the current official Calico installation instructions.

After installation:

```bash
kubectl get pods -n kube-system
kubectl get nodes
```

Master-1 should eventually become `Ready`.

---

# 11. Join Master-2 and Master-3

At the end of:

```bash
kubeadm init --upload-certs
```

kubeadm prints a control-plane join command similar to:

```bash
sudo kubeadm join NLB_DNS_NAME:6443 \
  --token TOKEN \
  --discovery-token-ca-cert-hash sha256:HASH \
  --control-plane \
  --certificate-key CERTIFICATE_KEY
```

SSH into Master-2 and run it.

Then SSH into Master-3 and run it.

Join control-plane nodes one at a time.

---

# 12. Check all control-plane nodes

On Master-1:

```bash
kubectl get nodes
```

Expected:

```text
NAME       STATUS   ROLES
master-1   Ready    control-plane
master-2   Ready    control-plane
master-3   Ready    control-plane
```

---

# 13. Create a worker join command

On Master-1:

```bash
kubeadm token create --print-join-command
```

Example:

```bash
kubeadm join NLB_DNS_NAME:6443 \
  --token TOKEN \
  --discovery-token-ca-cert-hash sha256:HASH
```

Notice that worker joins do not use:

```text
--control-plane
--certificate-key
```

---

# 14. Join all workers

SSH into Worker-1:

```bash
ssh -i my-aws-key.pem ubuntu@WORKER_1_PUBLIC_IP
```

Run the worker join command.

Repeat for:

```text
Worker-2
Worker-3
```

---

# 15. Verify the final cluster

On Master-1:

```bash
kubectl get nodes -o wide
```

Expected:

```text
NAME       STATUS   ROLES
master-1   Ready    control-plane
master-2   Ready    control-plane
master-3   Ready    control-plane
worker-1   Ready    <none>
worker-2   Ready    <none>
worker-3   Ready    <none>
```

---

# 16. Verify NLB health

In AWS:

```text
EC2
 -> Load Balancing
 -> Target Groups
 -> <cluster>-api
 -> Targets
```

You should eventually see:

```text
master-1    Healthy
master-2    Healthy
master-3    Healthy
```

The workers are deliberately not registered in the Kubernetes API NLB target group.

---

# 17. Test the public NLB endpoint

From your laptop:

```bash
nc -vz NLB_DNS_NAME 6443
```

or:

```bash
curl -k https://NLB_DNS_NAME:6443/version
```

Before kubeconfig/client authentication is configured, an HTTP response such as Unauthorized still proves that the public API endpoint is reachable.

---

# 18. Copy kubeconfig to your laptop

On Master-1:

```bash
cat ~/.kube/config
```

Securely copy it to your laptop, for example:

```bash
scp -i my-aws-key.pem \
  ubuntu@MASTER_1_PUBLIC_IP:/home/ubuntu/.kube/config \
  ./k8s-aws-config
```

Then:

```bash
export KUBECONFIG=$PWD/k8s-aws-config
```

Test:

```bash
kubectl get nodes
```

The kubeconfig server should point to:

```text
https://NLB_DNS_NAME:6443
```

because `kubeadm init` used the NLB as the control-plane endpoint.

---

# 19. Security notes

For learning, all six EC2 nodes are in public subnets.

For production, prefer:

```text
Internet
   |
Public NLB
   |
Private control-plane nodes
Private worker nodes
   |
NAT Gateway / VPC endpoints
```

Also consider:

- AWS Systems Manager instead of public SSH
- private node IPs only
- tightly restricted API CIDRs
- no `0.0.0.0/0` SSH
- external or carefully managed etcd
- automated certificate rotation
- AWS EBS CSI driver
- AWS cloud-controller-manager requirements
- centralized logging/monitoring
- backups
- explicit Kubernetes upgrade strategy

---

# 20. Terraform lifecycle

Show infrastructure:

```bash
terraform state list
```

See changes:

```bash
terraform plan
```

Destroy the entire lab:

```bash
terraform destroy
```

Enter:

```text
yes
```

This will remove the EC2 instances, NLB, networking, and security groups created by this project.

---

# 21. File overview

```text
aws-k8s-ha-terraform/
|
|-- versions.tf
|-- provider.tf
|-- variables.tf
|-- terraform.tfvars.example
|-- networking.tf
|-- security-groups.tf
|-- ec2.tf
|-- load-balancer.tf
|-- outputs.tf
|-- README.md
|
`-- scripts/
    `-- bootstrap.sh
```

---

# 22. Responsibility split

Terraform handles:

```text
VPC
Subnets
Internet Gateway
Routes
Security Groups
EC2
NLB
Target Group
Listener
AWS infrastructure bootstrap
```

kubeadm handles:

```text
Kubernetes control plane
Certificates
etcd
Joining control-plane nodes
Joining workers
```

The CNI handles:

```text
Pod networking
```

This separation makes troubleshooting much easier while learning.
