# Kubernetes Cluster on AWS EC2: Two Control Planes, Two Workers and an NLB

**Stack:** kubeadm + containerd + Calico  
**Operating system:** Ubuntu Server 24.04 LTS  
**Kubernetes:** v1.34  
**Calico:** v3.32.1  
**API endpoint:** Internet-facing AWS Network Load Balancer, TCP 6443

This guide creates four Kubernetes nodes on AWS EC2:

- Two control-plane nodes (`k8s-master1` and `k8s-master2`)
- Two worker nodes (`k8s-worker1` and `k8s-worker2`)
- One AWS Network Load Balancer in front of both Kubernetes API servers
- External `kubectl` access from a laptop through the NLB

The EC2 nodes communicate through private IPv4 addresses. Their public IP addresses are used only for SSH in this lab. Kubernetes nodes and external administrators use the NLB DNS name as the stable Kubernetes API endpoint.

> **Production warning:** Two control-plane nodes with stacked etcd are load-balanced, but they are not truly highly available. A two-member etcd cluster requires both members for quorum after the second member joins. Losing either control-plane node can prevent Kubernetes API writes. Kubernetes recommends three or more control-plane nodes, preferably an odd number. Use this two-control-plane design for learning/testing. For production, add a third control-plane node or use a separate three-member external etcd cluster.

---

## 1. Target architecture

```text
External laptop
     |
     | kubectl HTTPS/TCP 6443
     v
Internet-facing AWS Network Load Balancer
     | TCP 6443                  | TCP 6443
     v                           v
k8s-master1                  k8s-master2
172.31.10.20                 172.31.20.20
AZ-A                         AZ-B
     |                           |
     +-------------+-------------+
                   |
           Private VPC network
             |             |
             v             v
       k8s-worker1     k8s-worker2
       172.31.10.21    172.31.20.21
       AZ-A            AZ-B
```

Replace every example address with the actual values from your AWS account.

### Example node layout

| EC2 instance | Kubernetes role | Availability Zone | Example private IP |
|---|---|---|---|
| `k8s-master1` | First control plane | AZ-A | `172.31.10.20` |
| `k8s-master2` | Second control plane | AZ-B | `172.31.20.20` |
| `k8s-worker1` | Worker | AZ-A | `172.31.10.21` |
| `k8s-worker2` | Worker | AZ-B | `172.31.20.21` |

### Address-selection rule

| Connection | Address to use |
|---|---|
| Laptop to EC2 using SSH | EC2 public IP |
| Kubernetes node to Kubernetes API | NLB DNS name |
| Laptop `kubectl` to Kubernetes API | NLB DNS name |
| Node-to-node and Calico traffic | EC2 private IP network |
| API server advertise address on each master | That master's private IP |

---

## 2. Values to collect before starting

Keep these values in a local note:

```text
AWS_REGION=
VPC_ID=
VPC_CIDR=
PUBLIC_SUBNET_A=
PUBLIC_SUBNET_B=
MY_PUBLIC_IP=
MASTER1_PUBLIC_IP=
MASTER1_PRIVATE_IP=
MASTER2_PUBLIC_IP=
MASTER2_PRIVATE_IP=
WORKER1_PUBLIC_IP=
WORKER1_PRIVATE_IP=
WORKER2_PUBLIC_IP=
WORKER2_PRIVATE_IP=
NLB_DNS=
```

Find your laptop's current public IPv4 address:

```bash
curl -4 ifconfig.me
```

If the result is `203.0.113.10`, use `203.0.113.10/32` in security-group rules.

---

## 3. AWS networking prerequisites

You need:

1. A VPC with DNS resolution and DNS hostnames enabled.
2. Two public subnets in two different Availability Zones.
3. An internet gateway attached to the VPC.
4. A route from both public subnets to the internet gateway.
5. Four EC2 instances in the same VPC.

For a straightforward lab, place each instance in a public subnet and assign it a public IPv4 address. A production design normally places the Kubernetes nodes in private subnets and uses a bastion host, VPN or another controlled administration path.

Suggested lab placement:

| Subnet | AZ | Instances |
|---|---|---|
| Public subnet A | AZ-A | `k8s-master1`, `k8s-worker1` |
| Public subnet B | AZ-B | `k8s-master2`, `k8s-worker2` |

---

## 4. Create the security groups

Create two security groups:

- `k8s-api-nlb-sg` for the NLB
- `k8s-nodes-sg` for all four EC2 instances

### 4.1 NLB security group: `k8s-api-nlb-sg`

Inbound rules:

| Type | Protocol | Port | Source | Purpose |
|---|---|---:|---|---|
| Custom TCP | TCP | 6443 | Your laptop public IP `/32` | External `kubectl` |
| Custom TCP | TCP | 6443 | VPC CIDR, such as `172.31.0.0/16` | Node join and internal API access through the NLB |

Outbound rule:

| Type | Protocol | Port | Destination |
|---|---|---:|---|
| Custom TCP | TCP | 6443 | `k8s-nodes-sg` |

Attach the security group while creating the NLB. AWS does not allow a security group to be added later to an NLB that was originally created without one.

### 4.2 EC2 node security group: `k8s-nodes-sg`

Inbound rules:

| Type | Protocol | Port | Source | Purpose |
|---|---|---:|---|---|
| SSH | TCP | 22 | Your laptop public IP `/32` | SSH administration |
| Custom TCP | TCP | 6443 | `k8s-api-nlb-sg` | Kubernetes API traffic from the NLB |
| All traffic | All | All | `k8s-nodes-sg` | Private communication among cluster nodes |

Outbound rule:

```text
All traffic -> 0.0.0.0/0
```

The self-referencing `All traffic` rule is convenient for this lab and permits etcd, kubelet, Calico and other private cluster communication. Do not use `0.0.0.0/0` as the source for internal Kubernetes ports.

Important Kubernetes default ports include:

- TCP 6443: Kubernetes API server
- TCP 2379-2380: etcd
- TCP 10250: kubelet API
- TCP 10257: controller manager
- TCP 10259: scheduler
- TCP/UDP 30000-32767: NodePort Services, only when needed

---

## 5. Create four EC2 instances

Create four instances with these suggested settings:

```text
AMI: Ubuntu Server 24.04 LTS
Instance type: t3.medium or larger
Storage: 30 GB gp3
Key pair: Same key pair for all four instances
Security group: k8s-nodes-sg
Public IP: Enabled for this lab
```

Recommended names:

```text
k8s-master1
k8s-master2
k8s-worker1
k8s-worker2
```

Record each instance's private and public IPv4 addresses from the EC2 console.

---

## 6. Create the NLB target group

Create the target group before initializing Kubernetes.

1. Open **AWS Console -> EC2 -> Target Groups**.
2. Select **Create target group**.
3. Configure:

```text
Choose a target type: Instances
Target group name: k8s-api-targets
Protocol: TCP
Port: 6443
IP address type: IPv4
VPC: Same VPC as the four EC2 instances
```

4. Configure the health check:

```text
Health-check protocol: TCP
Health-check port: Traffic port
```

5. Register these targets on port 6443:

```text
k8s-master1
k8s-master2
```

6. Create the target group.

Both targets will initially be unhealthy because the API servers are not running yet. This is expected.

---

## 7. Create the internet-facing Network Load Balancer

1. Open **AWS Console -> EC2 -> Load Balancers**.
2. Select **Create load balancer -> Network Load Balancer**.
3. Configure:

```text
Load balancer name: k8s-api-nlb
Scheme: Internet-facing
IP address type: IPv4
VPC: Same VPC as the EC2 instances
Mappings: Select public subnet A and public subnet B
Security group: k8s-api-nlb-sg
```

4. Create the listener:

```text
Protocol: TCP
Port: 6443
Default action: Forward to k8s-api-targets
```

5. Create the NLB.
6. Optionally enable cross-zone load balancing.
7. Copy the NLB DNS name.

Example:

```text
k8s-api-nlb-1234567890.ap-south-1.elb.amazonaws.com
```

From now on, call this value `<NLB-DNS>`.

Do not place `https://` in `--control-plane-endpoint`; kubeadm expects only the DNS name and port.

---

## 8. Test the empty load-balancer path

From your laptop:

```bash
nc -zv -w 3 <NLB-DNS> 6443
```

Before Kubernetes starts, a refusal or failure is expected because no API server is listening. A persistent timeout after Kubernetes is initialized normally indicates an NLB, route, target-group or security-group problem.

---

## 9. SSH into all four instances

On your laptop:

```bash
chmod 400 my-key.pem
```

Connect to each instance using its public IP:

```bash
ssh -i my-key.pem ubuntu@<MASTER1-PUBLIC-IP>
ssh -i my-key.pem ubuntu@<MASTER2-PUBLIC-IP>
ssh -i my-key.pem ubuntu@<WORKER1-PUBLIC-IP>
ssh -i my-key.pem ubuntu@<WORKER2-PUBLIC-IP>
```

Use separate terminal windows to reduce confusion.

---

## 10. Configure hostnames

Run one command on the appropriate node.

On master 1:

```bash
sudo hostnamectl set-hostname k8s-master1
```

On master 2:

```bash
sudo hostnamectl set-hostname k8s-master2
```

On worker 1:

```bash
sudo hostnamectl set-hostname k8s-worker1
```

On worker 2:

```bash
sudo hostnamectl set-hostname k8s-worker2
```

Log out and reconnect after changing the hostname.

---

## 11. Perform the base preparation on all four nodes

Run every command in this section on `k8s-master1`, `k8s-master2`, `k8s-worker1` and `k8s-worker2`.

### 11.1 Update Ubuntu

```bash
sudo apt-get update
sudo apt-get upgrade -y
```

If Ubuntu reports that a reboot is required, reboot and reconnect before continuing:

```bash
sudo reboot
```

### 11.2 Disable swap

```bash
sudo swapoff -a
sudo sed -i '/[[:space:]]swap[[:space:]]/s/^/#/' /etc/fstab
swapon --show
```

`swapon --show` should produce no output.

### 11.3 Load the required kernel modules

```bash
sudo modprobe overlay
sudo modprobe br_netfilter
```

Make them persistent:

```bash
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF
```

### 11.4 Configure kernel networking

```bash
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
```

Apply and verify:

```bash
sudo sysctl --system
sysctl net.ipv4.ip_forward
```

Expected:

```text
net.ipv4.ip_forward = 1
```

### 11.5 Install and configure containerd

```bash
sudo apt-get install -y containerd
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml >/dev/null
```

Enable the systemd cgroup driver:

```bash
sudo sed -i \
  's/SystemdCgroup = false/SystemdCgroup = true/' \
  /etc/containerd/config.toml
```

Restart and enable containerd:

```bash
sudo systemctl restart containerd
sudo systemctl enable containerd
sudo systemctl status containerd --no-pager
```

Look for `active (running)`.

### 11.6 Install Kubernetes v1.34

```bash
sudo apt-get install -y apt-transport-https ca-certificates curl gpg
sudo mkdir -p -m 755 /etc/apt/keyrings
```

Install the Kubernetes repository key:

```bash
curl -fsSL \
  https://pkgs.k8s.io/core:/stable:/v1.34/deb/Release.key \
  | sudo gpg --dearmor \
  -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
```

Add the v1.34 repository:

```bash
echo \
  'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.34/deb/ /' \
  | sudo tee /etc/apt/sources.list.d/kubernetes.list
```

Install and hold the packages:

```bash
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
sudo systemctl enable --now kubelet
```

Kubelet may restart repeatedly before the node is initialized or joined. That is expected.

Verify:

```bash
kubeadm version
kubelet --version
kubectl version --client
```

Check containerd's cgroup setting:

```bash
grep -n 'SystemdCgroup' /etc/containerd/config.toml
```

Expected:

```text
SystemdCgroup = true
```

---

## 12. Initialize only the first control-plane node

Run this section only on `k8s-master1`.

Confirm its private IP:

```bash
hostname -I
```

Set variables with your actual values:

```bash
MASTER1_PRIVATE_IP="172.31.10.20"
NLB_DNS="k8s-api-nlb-1234567890.ap-south-1.elb.amazonaws.com"
```

Optionally pull Kubernetes images first:

```bash
sudo kubeadm config images pull
```

Initialize the first control plane:

```bash
sudo kubeadm init \
  --apiserver-advertise-address="${MASTER1_PRIVATE_IP}" \
  --control-plane-endpoint="${NLB_DNS}:6443" \
  --apiserver-cert-extra-sans="${NLB_DNS}" \
  --pod-network-cidr="192.168.0.0/16" \
  --upload-certs
```

The flags mean:

- `--apiserver-advertise-address`: master 1's private IP, used by its local API server.
- `--control-plane-endpoint`: the shared NLB DNS name and port used by all clients and joining nodes.
- `--apiserver-cert-extra-sans`: explicitly adds the NLB DNS name to the API-server TLS certificate.
- `--pod-network-cidr`: the Calico Pod address pool.
- `--upload-certs`: securely makes shared control-plane certificates available for the next control-plane join.

At the end, kubeadm prints two different join commands:

1. A control-plane join command containing `--control-plane` and `--certificate-key`.
2. A worker join command without those flags.

Save both commands securely. Do not post the token, CA hash or certificate key publicly.

---

## 13. Configure kubectl on master 1

On `k8s-master1`, as the `ubuntu` user:

```bash
mkdir -p "$HOME/.kube"
sudo cp /etc/kubernetes/admin.conf "$HOME/.kube/config"
sudo chown "$(id -u):$(id -g)" "$HOME/.kube/config"
chmod 600 "$HOME/.kube/config"
```

Verify the configured API endpoint:

```bash
kubectl config view --minify | grep server
```

Expected pattern:

```text
server: https://<NLB-DNS>:6443
```

Check the node:

```bash
kubectl get nodes
```

Master 1 will probably be `NotReady` until Calico is installed.

---

## 14. Verify NLB target health after master 1 initialization

Open:

```text
AWS Console -> EC2 -> Target Groups -> k8s-api-targets -> Targets
```

Expected at this point:

| Target | Expected state |
|---|---|
| `k8s-master1` | Healthy after a short delay |
| `k8s-master2` | Unhealthy because it has not joined yet |

From your laptop, test the API port:

```bash
nc -zv <NLB-DNS> 6443
```

The connection should now succeed.

---

## 15. Install Calico on master 1

Run only from `k8s-master1`, where kubectl is configured.

Install the Calico CRDs and Tigera operator:

```bash
kubectl create -f \
  https://raw.githubusercontent.com/projectcalico/calico/v3.32.1/manifests/v1_crd_projectcalico_org.yaml

kubectl create -f \
  https://raw.githubusercontent.com/projectcalico/calico/v3.32.1/manifests/tigera-operator.yaml
```

Download the custom resources:

```bash
curl -O \
  https://raw.githubusercontent.com/projectcalico/calico/v3.32.1/manifests/custom-resources.yaml
```

Confirm the CIDR:

```bash
grep -n "cidr:" custom-resources.yaml
```

It should show `192.168.0.0/16`.

Create the Calico resources:

```bash
kubectl create -f custom-resources.yaml
```

Watch status:

```bash
watch kubectl get tigerastatus
```

Exit with `Ctrl+C` when the components become available.

Check Kubernetes:

```bash
kubectl get nodes
kubectl get pods -A
```

---

## 16. Join the second control-plane node

Run the control-plane join command printed by `kubeadm init` on `k8s-master2`.

It will look like this:

```bash
sudo kubeadm join <NLB-DNS>:6443 \
  --token <REAL-TOKEN> \
  --discovery-token-ca-cert-hash sha256:<REAL-CA-HASH> \
  --control-plane \
  --certificate-key <REAL-CERTIFICATE-KEY> \
  --apiserver-advertise-address=<MASTER2-PRIVATE-IP>
```

Example structure only:

```bash
sudo kubeadm join \
  k8s-api-nlb-1234567890.ap-south-1.elb.amazonaws.com:6443 \
  --token abcdef.0123456789abcdef \
  --discovery-token-ca-cert-hash sha256:0123456789abcdef... \
  --control-plane \
  --certificate-key 0123456789abcdef... \
  --apiserver-advertise-address=172.31.20.20
```

Use the real values printed by your first control plane. Never copy the example token or hashes.

### If the control-plane certificate key expired

On master 1, upload the certificates again:

```bash
sudo kubeadm init phase upload-certs --upload-certs
```

Save the new certificate key.

Generate a fresh base join command:

```bash
kubeadm token create --print-join-command
```

Append these options to that command before running it on master 2:

```text
--control-plane
--certificate-key <NEW-CERTIFICATE-KEY>
--apiserver-advertise-address=<MASTER2-PRIVATE-IP>
```

### Configure kubectl on master 2

After the join completes, run on master 2:

```bash
mkdir -p "$HOME/.kube"
sudo cp /etc/kubernetes/admin.conf "$HOME/.kube/config"
sudo chown "$(id -u):$(id -g)" "$HOME/.kube/config"
chmod 600 "$HOME/.kube/config"
```

Verify:

```bash
kubectl get nodes -o wide
```

After a short delay, the NLB target group should show both masters as healthy.

---

## 17. Generate or retrieve the worker join command

If you saved the worker command printed by `kubeadm init`, use it.

Otherwise, run on either control-plane node:

```bash
kubeadm token create --print-join-command
```

It should use the NLB endpoint and resemble:

```bash
sudo kubeadm join <NLB-DNS>:6443 \
  --token <REAL-TOKEN> \
  --discovery-token-ca-cert-hash sha256:<REAL-CA-HASH>
```

If the printed command unexpectedly contains a master's private IP, replace only the endpoint with `<NLB-DNS>:6443`. The token and CA hash remain unchanged.

---

## 18. Join worker 1

Run the real worker join command on `k8s-worker1`:

```bash
sudo kubeadm join <NLB-DNS>:6443 \
  --token <REAL-TOKEN> \
  --discovery-token-ca-cert-hash sha256:<REAL-CA-HASH>
```

Do not add `--control-plane` or `--certificate-key` to a worker join.

---

## 19. Join worker 2

Run the same valid worker join command on `k8s-worker2`:

```bash
sudo kubeadm join <NLB-DNS>:6443 \
  --token <REAL-TOKEN> \
  --discovery-token-ca-cert-hash sha256:<REAL-CA-HASH>
```

The same bootstrap token can join more than one worker while the token remains valid.

---

## 20. Verify the complete cluster

Run on either control-plane node:

```bash
kubectl get nodes -o wide
```

Expected structure:

```text
NAME          STATUS   ROLES           INTERNAL-IP
k8s-master1   Ready    control-plane   172.31.10.20
k8s-master2   Ready    control-plane   172.31.20.20
k8s-worker1   Ready    <none>          172.31.10.21
k8s-worker2   Ready    <none>          172.31.20.21
```

Check all system Pods:

```bash
kubectl get pods -A -o wide
```

Check control-plane readiness through the NLB:

```bash
kubectl get --raw='/readyz?verbose'
```

Check cluster endpoints:

```bash
kubectl cluster-info
```

---

## 21. Configure external kubectl access on your laptop

Because the cluster was initialized using the NLB as `--control-plane-endpoint`, `/etc/kubernetes/admin.conf` should already contain the NLB endpoint.

### 21.1 Prepare a copy on master 1

Run on master 1:

```bash
sudo cp /etc/kubernetes/admin.conf /home/ubuntu/aws-kubeconfig
sudo chown ubuntu:ubuntu /home/ubuntu/aws-kubeconfig
sudo chmod 600 /home/ubuntu/aws-kubeconfig
```

### 21.2 Install kubectl on a Mac

```bash
brew install kubectl
```

### 21.3 Copy kubeconfig to the laptop

Run on the laptop:

```bash
mkdir -p "$HOME/.kube"

scp -i my-key.pem \
  ubuntu@<MASTER1-PUBLIC-IP>:/home/ubuntu/aws-kubeconfig \
  "$HOME/.kube/aws-kubeconfig"

chmod 600 "$HOME/.kube/aws-kubeconfig"
```

### 21.4 Use the new context

For the current terminal:

```bash
export KUBECONFIG="$HOME/.kube/aws-kubeconfig"
```

Or use it explicitly:

```bash
kubectl --kubeconfig "$HOME/.kube/aws-kubeconfig" get nodes
```

Verify the configured server:

```bash
kubectl --kubeconfig "$HOME/.kube/aws-kubeconfig" \
  config view --minify | grep server
```

Expected:

```text
server: https://<NLB-DNS>:6443
```

Test external access:

```bash
kubectl --kubeconfig "$HOME/.kube/aws-kubeconfig" cluster-info
kubectl --kubeconfig "$HOME/.kube/aws-kubeconfig" get nodes
kubectl --kubeconfig "$HOME/.kube/aws-kubeconfig" get pods -A
```

The communication flow is:

```text
Laptop kubectl -> NLB TCP 6443 -> healthy master API server TCP 6443
```

> `admin.conf` provides cluster-admin privileges. Do not share it with multiple users. For a team, configure individual authentication identities and Kubernetes RBAC permissions.

---

## 22. Test scheduling with NGINX

```bash
kubectl create deployment nginx --image=nginx
kubectl get pods -o wide
```

Expose it internally first:

```bash
kubectl expose deployment nginx --port=80 --target-port=80
kubectl get svc nginx
```

The NLB created in this guide is only for the Kubernetes API. Do not reuse it for application traffic. Expose applications separately using an Ingress controller, an application NLB/ALB, or another suitable Kubernetes Service design.

If you deliberately expose NGINX using NodePort:

```bash
kubectl expose deployment nginx \
  --name=nginx-nodeport \
  --type=NodePort \
  --port=80

kubectl get svc nginx-nodeport
```

Only open the allocated NodePort in an appropriate security group if external NodePort access is actually required.

---

## 23. NLB and external-access troubleshooting

### 23.1 Both NLB targets are unhealthy

On each master:

```bash
sudo ss -lntp | grep 6443
sudo crictl ps | grep kube-apiserver
sudo journalctl -u kubelet --no-pager -n 100
```

Verify:

- The target group uses TCP 6443.
- Both master instances are registered on port 6443.
- The NLB listener uses TCP 6443.
- `k8s-nodes-sg` accepts TCP 6443 from `k8s-api-nlb-sg`.
- The NLB security group's outbound rules permit TCP 6443 to the nodes.
- The NLB includes the Availability Zones where the masters run.

### 23.2 External kubectl times out

From the laptop:

```bash
nc -zv <NLB-DNS> 6443
```

Check:

- Your laptop's public IP has not changed.
- NLB security-group ingress allows your current public IP `/32`.
- At least one NLB target is healthy.
- Network ACLs and route tables permit the traffic.

### 23.3 Nodes cannot join through the NLB

From the joining node:

```bash
nc -zv <NLB-DNS> 6443
curl -k https://<NLB-DNS>:6443/version
```

Check that NLB ingress permits TCP 6443 from the VPC CIDR.

### 23.4 TLS certificate error

Example:

```text
x509: certificate is valid for ... not for <NLB-DNS>
```

This indicates that the cluster was initialized without the NLB DNS name in the API-server certificate. The correct initial command must include:

```text
--control-plane-endpoint=<NLB-DNS>:6443
--apiserver-cert-extra-sans=<NLB-DNS>
```

For a new lab cluster, the cleanest correction is usually to reset all nodes and initialize again with the correct endpoint.

### 23.5 Worker remains NotReady

```bash
kubectl describe node k8s-worker1
kubectl get pods -A -o wide
kubectl get tigerastatus
```

Check Calico Pods and confirm the self-referencing node security-group rule permits node-to-node traffic.

### 23.6 Control-plane certificate key expired

On master 1:

```bash
sudo kubeadm init phase upload-certs --upload-certs
kubeadm token create --print-join-command
```

Use the new certificate key and fresh join token when joining master 2.

### 23.7 Token expired

On either existing control-plane node:

```bash
kubeadm token create --print-join-command
```

Run the new command on the worker.

---

## 24. Resetting a node

Run this only when you intentionally want to remove kubeadm state from that node:

```bash
sudo kubeadm reset -f
```

Optional CNI cleanup for a disposable lab node:

```bash
sudo rm -rf /etc/cni/net.d
sudo systemctl restart containerd
sudo systemctl restart kubelet
```

If you reset a control-plane node that already joined stacked etcd, also remove the old node and etcd membership correctly from the surviving cluster before trying to rejoin. For disposable labs, recreating the affected EC2 instance is frequently simpler and safer.

---

## 25. Compact execution order

Use this order to avoid endpoint and certificate problems:

1. Create VPC/subnets or select existing ones.
2. Create `k8s-api-nlb-sg` and `k8s-nodes-sg`.
3. Create four EC2 instances.
4. Create the TCP 6443 target group and register both masters.
5. Create the internet-facing NLB and copy its DNS name.
6. Prepare all four nodes: swap, kernel modules, sysctl and containerd.
7. Install Kubernetes v1.34 on all four nodes.
8. Initialize master 1 with the NLB `--control-plane-endpoint` and `--upload-certs`.
9. Configure kubectl on master 1.
10. Install Calico.
11. Join master 2 using the control-plane join command.
12. Join both workers using the worker join command.
13. Confirm both masters are healthy in the NLB target group.
14. Copy kubeconfig to the laptop.
15. Verify external `kubectl` through the NLB.

---

## 26. Security recommendations

- Restrict NLB TCP 6443 ingress to trusted public IPs or a corporate network.
- Restrict SSH TCP 22 to your own public IP or use a bastion/VPN.
- Never expose etcd ports 2379-2380 to the internet.
- Do not distribute `/etc/kubernetes/admin.conf` to a team.
- Use individual identities and RBAC for multiple administrators.
- Use three control-plane nodes for production stacked-etcd quorum.
- Place production nodes in private subnets.
- Use a separate ingress/application load balancer for workloads.
- Back up etcd and test recovery procedures.
- Plan Kubernetes upgrades deliberately; do not unhold packages and upgrade all nodes simultaneously.

---

## 27. References

- Kubernetes v1.34 — Installing kubeadm: https://v1-34.docs.kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
- Kubernetes v1.34 — Highly available clusters with kubeadm: https://v1-34.docs.kubernetes.io/docs/setup/production-environment/tools/kubeadm/high-availability/
- Kubernetes — Ports and protocols: https://kubernetes.io/docs/reference/networking/ports-and-protocols/
- Kubernetes — Creating a cluster with kubeadm: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/
- Kubernetes — Container runtimes: https://kubernetes.io/docs/setup/production-environment/container-runtimes/
- Calico — Self-managed Kubernetes installation: https://docs.tigera.io/calico/latest/getting-started/kubernetes/self-managed-onprem/onpremises
- Calico — System requirements: https://docs.tigera.io/calico/latest/getting-started/kubernetes/requirements
- AWS — Network Load Balancers: https://docs.aws.amazon.com/elasticloadbalancing/latest/network/introduction.html
- AWS — NLB target groups: https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html
- AWS — NLB security groups: https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-security-groups.html

