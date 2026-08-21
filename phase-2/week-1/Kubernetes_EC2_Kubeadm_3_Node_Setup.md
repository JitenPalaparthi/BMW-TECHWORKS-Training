# Kubernetes 3-Node Cluster on AWS EC2

kubeadm + containerd + Calico • Ubuntu 24.04 • Kubernetes v1.34

This guide builds a Kubernetes cluster with one EC2 control-plane node and two EC2 worker nodes. The EC2 instances communicate with each other using their private IPv4 addresses. Public IP addresses are used only when you SSH from your laptop to the instances.

# 1. Target architecture

Note: Replace the example 172.31.x.x addresses throughout this document with the actual private IPv4 addresses shown in your EC2 console.

# 2. AWS prerequisites

- Create three Ubuntu 24.04 EC2 instances in the same VPC. For a lab, place them in the same subnet when convenient.
- Recommended minimum for the control plane: 2 vCPU and 4 GB RAM. Workers should have enough resources for the workloads you plan to run.
- Attach a security group that allows the Kubernetes nodes to communicate with one another over the required cluster ports.
- Keep SSH (TCP 22) restricted to your own public IP instead of opening it to the whole internet.
# 3. Recommended EC2 Security Group rules

For a learning environment, the simplest safe arrangement is to attach the same Kubernetes security group to all three nodes and allow inbound traffic from that same security group. This permits private node-to-node communication without exposing all Kubernetes ports publicly.

For a stricter production-style configuration, explicitly allow the required Kubernetes ports instead of all traffic. Important defaults include control-plane TCP 6443 for the API server, TCP 2379-2380 for etcd, TCP 10250 for the kubelet API, TCP 10257 for the controller manager, TCP 10259 for the scheduler, and worker NodePort range TCP/UDP 30000-32767 when NodePort services are used.

# 4. SSH into each EC2 instance

```bash
chmod 400 my-key.pem
```

What it does: Restricts permissions on the EC2 private key. OpenSSH rejects private keys that are readable by other users.

```bash
ssh -i my-key.pem ubuntu@<PUBLIC-IP>
```

What it does: Connects from your laptop to an Ubuntu EC2 instance. Use the instance's public IP here, because the connection originates outside the VPC.

# 5. Perform the following preparation on ALL THREE nodes

## 5.1 Update the operating system

```bash
sudo apt-get update
```

What it does: Downloads the latest package metadata from configured Ubuntu repositories.

```bash
sudo apt-get upgrade -y
```

What it does: Installs available package upgrades without prompting for confirmation.

## 5.2 Set hostnames

Run one appropriate command on each EC2 instance:

```bash
# EC2-1
sudo hostnamectl set-hostname k8s-master

# EC2-2
sudo hostnamectl set-hostname k8s-worker1

# EC2-3
sudo hostnamectl set-hostname k8s-worker2
```

What it does: Gives each node a clear name. Kubernetes will normally use the machine hostname as the node name.

## 5.3 Disable swap

```bash
sudo swapoff -a
```

What it does: Disables active swap immediately. The standard kubelet setup expects swap to be disabled unless swap support is deliberately configured.

```bash
swapon --show
```

What it does: Displays active swap devices/files. No output means swap is disabled.

Note: Most Ubuntu EC2 images do not have swap enabled by default, but running these checks is harmless and makes the prerequisite explicit.

## 5.4 Load required kernel modules

```bash
sudo modprobe overlay
```

What it does: Loads the overlay filesystem kernel module used by container runtimes for layered container filesystems.

```bash
sudo modprobe br_netfilter
```

What it does: Allows Linux bridge traffic to be visible to netfilter/iptables processing, which Kubernetes networking commonly requires.

```bash
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF
```

What it does: Makes these modules part of the node's persistent Kubernetes-related module configuration.

## 5.5 Configure kernel networking

```bash
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
```

What it does: Enables packet forwarding and ensures bridged IPv4/IPv6 traffic is processed by the host's packet-filtering path. IP forwarding is required because Kubernetes nodes route traffic between Pods and networks.

```bash
sudo sysctl --system
```

What it does: Reloads sysctl configuration files so the networking settings take effect now.

```bash
sysctl net.ipv4.ip_forward
```

What it does: Verifies IPv4 forwarding. The expected value is net.ipv4.ip_forward = 1.

## 5.6 Install and configure containerd

```bash
sudo apt-get install -y containerd
```

What it does: Installs containerd, the container runtime Kubernetes will use through the Container Runtime Interface.

```bash
sudo mkdir -p /etc/containerd
```

What it does: Creates containerd's configuration directory if it does not already exist.

```bash
containerd config default | sudo tee /etc/containerd/config.toml
```

What it does: Generates a complete default containerd configuration and stores it in /etc/containerd/config.toml.

```bash
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
```

What it does: Makes containerd use the systemd cgroup driver. Matching the kubelet and runtime cgroup driver avoids resource-management problems.

```bash
sudo systemctl restart containerd
```

What it does: Restarts containerd so the new configuration is loaded.

```bash
sudo systemctl enable containerd
```

What it does: Configures containerd to start automatically whenever the EC2 instance boots.

```bash
sudo systemctl status containerd --no-pager
```

What it does: Checks whether containerd is running. Look for active (running).

## 5.7 Install kubelet, kubeadm and kubectl

```bash
sudo apt-get install -y apt-transport-https ca-certificates curl gpg
```

What it does: Installs utilities and certificates needed to securely configure the Kubernetes package repository.

```bash
sudo mkdir -p -m 755 /etc/apt/keyrings
```

What it does: Creates the directory used for repository signing keys.

```bash
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.34/deb/Release.key \
  | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
```

What it does: Downloads the official Kubernetes repository signing key and stores it in APT keyring format.

```bash
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.34/deb/ /' \
  | sudo tee /etc/apt/sources.list.d/kubernetes.list
```

What it does: Adds the official Kubernetes v1.34 package repository.

```bash
sudo apt-get update
```

What it does: Refreshes APT metadata again, now including the Kubernetes repository.

```bash
sudo apt-get install -y kubelet kubeadm kubectl
```

What it does: Installs the node agent (kubelet), cluster bootstrap utility (kubeadm), and Kubernetes command-line client (kubectl).

```bash
sudo apt-mark hold kubelet kubeadm kubectl
```

What it does: Prevents unattended package upgrades from unexpectedly changing Kubernetes minor versions.

```bash
sudo systemctl enable --now kubelet
```

What it does: Enables kubelet at boot and starts it. Before kubeadm initializes or joins the node, kubelet may restart repeatedly; that is expected.

```bash
kubeadm version
kubectl version --client
```

What it does: Confirms the Kubernetes bootstrap and client binaries are installed.

# 6. Initialize ONLY the control-plane node (EC2-1)

First determine EC2-1's private IPv4 address:

```bash
hostname -I
```

What it does: Shows the addresses assigned to the instance. Confirm the EC2 private IPv4 address in the AWS console as well.

```bash
sudo kubeadm init \
  --apiserver-advertise-address=172.31.10.20 \
  --pod-network-cidr=192.168.0.0/16
```

What it does: Initializes the Kubernetes control plane. --apiserver-advertise-address tells the API server which node address other cluster members should use; on EC2 nodes in the same VPC, use EC2-1's PRIVATE IPv4 address. --pod-network-cidr reserves 192.168.0.0/16 for Pod networking, which matches the Calico configuration used later.

Note: Run kubeadm init only on EC2-1. Do not run it on worker nodes.

# 7. Configure kubectl on the control-plane node

```bash
mkdir -p $HOME/.kube
```

What it does: Creates the current user's Kubernetes client configuration directory.

```bash
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
```

What it does: Copies the administrative kubeconfig generated by kubeadm into your user's default kubectl location.

```bash
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

What it does: Changes ownership of the kubeconfig to your current user so kubectl can read it without sudo.

```bash
kubectl get nodes
```

What it does: Queries the API server for registered nodes. The control plane may initially show NotReady because no CNI has been installed yet.

# 8. Install Calico networking on EC2-1

Kubernetes requires a CNI plugin for Pod networking. The commands below use Calico v3.32.1 and a 192.168.0.0/16 Pod CIDR.

```bash
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.1/manifests/v1_crd_projectcalico_org.yaml

kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.1/manifests/tigera-operator.yaml
```

What it does: Installs the Calico custom resource definitions and Tigera operator.

```bash
curl -O https://raw.githubusercontent.com/projectcalico/calico/v3.32.1/manifests/custom-resources.yaml
grep -n "cidr:" custom-resources.yaml
```

What it does: Downloads Calico's installation custom resources and shows the configured IP pool CIDR. Confirm that the CIDR is 192.168.0.0/16 before creating it.

```bash
kubectl create -f custom-resources.yaml
```

What it does: Creates the Calico installation resources. Calico then configures Pod networking across the cluster.

```bash
watch kubectl get tigerastatus
```

What it does: Watches Calico component health. Exit with Ctrl+C after components become available.

```bash
kubectl get nodes
```

What it does: Checks node readiness. The control-plane node should become Ready after networking is operational.

```bash
kubectl get pods -A
```

What it does: Displays Pods in all namespaces so you can verify Kubernetes system components and Calico are running.

# 9. Obtain the worker join command

At the end of kubeadm init, Kubernetes prints a join command. If you did not save it, generate a new one on EC2-1:

```bash
kubeadm token create --print-join-command
```

What it does: Creates a bootstrap token and prints a complete kubeadm join command, including the CA certificate hash required to securely discover the control plane.

```bash
sudo kubeadm join 172.31.10.20:6443 \
  --token <REAL-TOKEN> \
  --discovery-token-ca-cert-hash sha256:<REAL-HASH>
```

The 172.31.10.20 address is the CONTROL PLANE'S PRIVATE IP. Do not replace it with each worker's address. Both workers connect to the same API server endpoint. Replace the placeholder token and hash with the real values printed by kubeadm.

# 10. Join EC2-2 and EC2-3 as workers

After completing all preparation steps in Section 5 on EC2-2, run the real join command there:

```bash
sudo kubeadm join 172.31.10.20:6443 \
  --token <REAL-TOKEN> \
  --discovery-token-ca-cert-hash sha256:<REAL-HASH>
```

Run the same join command on EC2-3. The same valid bootstrap token can be used to join multiple worker nodes.

Note: If the token expires, simply run kubeadm token create --print-join-command again on EC2-1 and use the newly generated command.

# 11. Verify the complete cluster

Return to EC2-1 and run:

```bash
kubectl get nodes -o wide
```

What it does: Shows all nodes, their roles, Kubernetes version, internal/private IP addresses, operating system information, and container runtime.

```bash
# Typical result
NAME          STATUS   ROLES           INTERNAL-IP
k8s-master    Ready    control-plane   172.31.10.20
k8s-worker1   Ready    <none>          172.31.10.21
k8s-worker2   Ready    <none>          172.31.10.22
```

```bash
kubectl get pods -A -o wide
```

What it does: Shows all cluster Pods and the nodes on which they are running. Use this to verify CoreDNS, kube-proxy, Calico and control-plane components.

# 12. Test scheduling with an application

```bash
kubectl create deployment nginx --image=nginx
```

What it does: Creates a Deployment that asks Kubernetes to run an NGINX Pod.

```bash
kubectl get pods -o wide
```

What it does: Shows where the NGINX Pod was scheduled. With workers available, it should normally run on a worker.

```bash
kubectl expose deployment nginx --type=NodePort --port=80
```

What it does: Creates a NodePort Service so NGINX can be reached through a node port. If you access this externally, the relevant NodePort must also be permitted by the EC2 security group.

```bash
kubectl get svc nginx
```

What it does: Displays the allocated NodePort and service details.

# 13. Public IP vs private IP: the rule to remember

# 14. Common problems and troubleshooting

# 15. Resetting a node when you need to start over

```bash
sudo kubeadm reset -f
```

What it does: Removes Kubernetes state created by kubeadm from that node. Use this when a failed init/join attempt must be cleaned before retrying.

Note: A kubeadm reset does not necessarily remove all CNI network artifacts or firewall rules. For a disposable lab, recreating the EC2 instance is often the cleanest recovery.

# 16. Compact command sequence

All nodes — base setup:

```bash
sudo apt-get update
sudo apt-get upgrade -y
sudo swapoff -a
sudo modprobe overlay
sudo modprobe br_netfilter

cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

sudo sysctl --system

sudo apt-get install -y containerd
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd

sudo apt-get install -y apt-transport-https ca-certificates curl gpg
sudo mkdir -p -m 755 /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.34/deb/Release.key \
  | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.34/deb/ /' \
  | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
sudo systemctl enable --now kubelet
```

Control-plane only:

```bash
sudo kubeadm init \
  --apiserver-advertise-address=<CONTROL-PLANE-PRIVATE-IP> \
  --pod-network-cidr=192.168.0.0/16

mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.1/manifests/v1_crd_projectcalico_org.yaml
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.1/manifests/tigera-operator.yaml
curl -O https://raw.githubusercontent.com/projectcalico/calico/v3.32.1/manifests/custom-resources.yaml
kubectl create -f custom-resources.yaml

kubeadm token create --print-join-command
```

Each worker:

```bash
sudo kubeadm join <CONTROL-PLANE-PRIVATE-IP>:6443 \
  --token <REAL-TOKEN> \
  --discovery-token-ca-cert-hash sha256:<REAL-HASH>
```

# 17. References

- Kubernetes — Installing kubeadm (v1.34): https://v1-34.docs.kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
- Kubernetes — Ports and Protocols: https://kubernetes.io/docs/reference/networking/ports-and-protocols/
- Kubernetes — Control Plane / Node Communication: https://kubernetes.io/docs/concepts/architecture/control-plane-node-communication/
- Calico — Kubernetes documentation: https://docs.tigera.io/calico/latest/getting-started/kubernetes/
## Reference tables

### Node layout

| Node | Role | Example private IP | Used for |
|---|---|---|---|
| EC2-1 | Control Plane | 172.31.10.20 | Kubernetes API server and control-plane components |
| EC2-2 | Worker | 172.31.10.21 | Runs application Pods |
| EC2-3 | Worker | 172.31.10.22 | Runs application Pods |

### Public vs private IP

| Connection | Address to use | Reason |
|---|---|---|
| Laptop → EC2 via SSH | EC2 public IP | Your laptop is outside the AWS VPC |
| Worker → Kubernetes API server | Control-plane private IP | Cluster traffic stays inside the VPC |
| Node → node / Pod networking | Private network | Avoids routing cluster-internal traffic over the public internet |
