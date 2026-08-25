# Minikube Kubernetes RBAC Complete Guide

## Create Human Users and Give Only Selected Role / ClusterRole Access

This hands-on guide shows how to create multiple human users in a Minikube Kubernetes cluster and give each user only the permissions they actually need.

The lab uses:

- Minikube
- kubectl
- OpenSSL
- Kubernetes X.509 client certificates
- Kubernetes `CertificateSigningRequest` objects
- `Role`
- `RoleBinding`
- `ClusterRole`
- `ClusterRoleBinding`
- Group-based RBAC
- Per-user kubeconfig files
- `kubectl auth can-i` verification

---

# 1. Target Architecture

We will build this example:

```text
                         MINIKUBE CLUSTER
                                |
       +------------------------+-------------------------+
       |                        |                         |
       |                        |                         |
 development namespace     qa namespace             Cluster Scope
       |                        |                         |
       |                        |                         |
 developer1                  qauser1                   auditor1
       |                        |                         |
 RoleBinding                RoleBinding          ClusterRoleBinding
       |                        |                         |
 developer-role              qa-viewer             cluster-auditor
       |                        |                         |
 Pods/Deployments            Read only              Nodes +
 in development              QA resources            namespaces +
 only                        only                    pods cluster-wide
```

The important objective is **least privilege**.

For example:

```text
developer1
    CAN:
      get/list/watch Pods in development
      create/update/delete Deployments in development

    CANNOT:
      access kube-system
      access qa
      list Nodes
      create ClusterRoles
      modify cluster-wide RBAC
```

---

# 2. Authentication vs Authorization

Kubernetes first authenticates a user and then authorizes the API request.

```text
User
 |
 | certificate
 v
Kubernetes API Server
 |
 +---- Authentication ----> Who are you?
 |
 +---- Authorization -----> What are you allowed to do?
                              |
                              v
                             RBAC
```

A valid certificate does **not** automatically give the user permissions.

For example:

```text
Certificate says:
CN=developer1

Kubernetes authenticates:
Username = developer1

RBAC then checks:
Can developer1 list Pods in namespace development?
```

---

# 3. Kubernetes Does Not Have a Built-In User Object

There is no normal Kubernetes resource like:

```yaml
apiVersion: v1
kind: User
metadata:
  name: developer1
```

This command also does not exist:

```bash
kubectl create user developer1
```

Human users are normally authenticated using mechanisms such as:

- X.509 client certificates
- OIDC
- External identity providers
- Authentication proxy
- Webhook authentication

For this Minikube lab, we will use X.509 client certificates.

---

# 4. RBAC Objects

Kubernetes RBAC has four primary objects:

| Object | Scope | Purpose |
|---|---|---|
| `Role` | Namespace | Defines permissions in one namespace |
| `RoleBinding` | Namespace | Assigns a Role or ClusterRole inside one namespace |
| `ClusterRole` | Cluster | Defines reusable or cluster-scoped permissions |
| `ClusterRoleBinding` | Cluster | Assigns ClusterRole permissions cluster-wide |

Remember:

```text
Role + RoleBinding
        |
        v
One namespace
```

```text
ClusterRole + RoleBinding
        |
        v
One namespace
```

```text
ClusterRole + ClusterRoleBinding
        |
        v
Whole cluster
```

---

# 5. Prerequisites

You need:

```text
minikube
kubectl
openssl
```

Verify:

```bash
minikube version
kubectl version --client
openssl version
```

Docker is also commonly used as the Minikube driver.

Verify:

```bash
docker version
```

---

# 6. Start Minikube

Start a cluster:

```bash
minikube start
```

Or explicitly use Docker:

```bash
minikube start --driver=docker
```

Check:

```bash
minikube status
```

Check Kubernetes nodes:

```bash
kubectl get nodes
```

Example:

```text
NAME       STATUS   ROLES           AGE   VERSION
minikube   Ready    control-plane   ...   ...
```

Verify your admin context:

```bash
kubectl config current-context
```

Expected:

```text
minikube
```

Check identity:

```bash
kubectl auth whoami
```

---

# 7. Create Namespaces

Create development and QA namespaces:

```bash
kubectl create namespace development
kubectl create namespace qa
```

Verify:

```bash
kubectl get namespaces
```

---

# 8. Create Test Applications

Create an application in development:

```bash
kubectl create deployment nginx-dev \
  --image=nginx \
  -n development
```

Create an application in QA:

```bash
kubectl create deployment nginx-qa \
  --image=nginx \
  -n qa
```

Check:

```bash
kubectl get pods -n development
kubectl get pods -n qa
```

---

# 9. Create a Working Directory

Create:

```bash
mkdir -p ~/minikube-rbac-lab
cd ~/minikube-rbac-lab
```

We will store:

```text
developer1.key
developer1.csr
developer1.crt
developer1.kubeconfig

qauser1.key
qauser1.csr
qauser1.crt
qauser1.kubeconfig

auditor1.key
auditor1.csr
auditor1.crt
auditor1.kubeconfig
```

---

# 10. User 1: developer1

We want:

```text
developer1

Namespace:
development

Permissions:
Pods:
  get
  list
  watch

Deployments:
  get
  list
  watch
  create
  update
  patch
  delete
```

But the user must NOT have access to:

```text
qa
kube-system
nodes
clusterroles
clusterrolebindings
```

---

# 11. Generate developer1 Private Key

Run:

```bash
openssl genrsa \
  -out developer1.key \
  3072
```

Protect the key:

```bash
chmod 600 developer1.key
```

---

# 12. Generate developer1 Certificate Signing Request

Run:

```bash
openssl req \
  -new \
  -key developer1.key \
  -out developer1.csr \
  -subj "/CN=developer1/O=developers"
```

Meaning:

```text
CN = developer1
```

becomes the Kubernetes username.

```text
O = developers
```

becomes a Kubernetes group.

Conceptually:

```text
Certificate subject

CN=developer1
O=developers

        |
        v

Kubernetes identity

Username: developer1
Group: developers
```

---

# 13. Create a Kubernetes CertificateSigningRequest

First encode the CSR.

## macOS

```bash
CSR_BASE64=$(cat developer1.csr | base64 | tr -d '\n')
```

## Linux

```bash
CSR_BASE64=$(base64 -w 0 developer1.csr)
```

Now create the Kubernetes CSR resource:

```bash
cat > developer1-k8s-csr.yaml <<EOF
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: developer1
spec:
  request: ${CSR_BASE64}
  signerName: kubernetes.io/kube-apiserver-client
  expirationSeconds: 31536000
  usages:
  - client auth
EOF
```

Apply:

```bash
kubectl apply \
  -f developer1-k8s-csr.yaml
```

Check:

```bash
kubectl get csr
```

Initially you should see something similar to:

```text
NAME         AGE   SIGNERNAME                            REQUESTOR   CONDITION
developer1   ...   kubernetes.io/kube-apiserver-client  ...         Pending
```

---

# 14. Inspect the CSR Before Approving

As cluster administrator:

```bash
kubectl describe csr developer1
```

Before approving a certificate in a real environment, verify:

```text
username
group membership
requested usages
requestor
```

Do not blindly approve arbitrary certificate requests.

---

# 15. Approve developer1 CSR

Approve:

```bash
kubectl certificate approve developer1
```

Check:

```bash
kubectl get csr developer1
```

Expected condition:

```text
Approved,Issued
```

---

# 16. Download developer1 Certificate

Retrieve the certificate:

```bash
kubectl get csr developer1 \
  -o jsonpath='{.status.certificate}' \
  | base64 --decode \
  > developer1.crt
```

On systems where `base64 --decode` is unavailable, use the appropriate local equivalent such as:

```bash
base64 -d
```

Verify:

```bash
openssl x509 \
  -in developer1.crt \
  -noout \
  -subject \
  -issuer \
  -dates
```

You should see the subject containing:

```text
CN = developer1
O = developers
```

---

# 17. Get Minikube Cluster Information

Get API server:

```bash
SERVER=$(kubectl config view \
  --minify \
  -o jsonpath='{.clusters[0].cluster.server}')
```

Check:

```bash
echo "$SERVER"
```

Get CA data:

```bash
CA_DATA=$(kubectl config view \
  --raw \
  --minify \
  -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')
```

If your kubeconfig uses a CA file path instead of embedded CA data, this variable may be empty.

A portable approach is to create the user kubeconfig by copying cluster connection information from the current Minikube configuration as shown below.

---

# 18. Create developer1 Kubeconfig

First extract the Minikube CA certificate.

Try:

```bash
kubectl config view \
  --raw \
  --minify \
  -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' \
  | base64 --decode \
  > minikube-ca.crt
```

Verify:

```bash
openssl x509 \
  -in minikube-ca.crt \
  -noout \
  -subject
```

Now configure the cluster:

```bash
kubectl config \
  --kubeconfig=developer1.kubeconfig \
  set-cluster minikube \
  --server="$SERVER" \
  --certificate-authority=minikube-ca.crt \
  --embed-certs=true
```

Configure credentials:

```bash
kubectl config \
  --kubeconfig=developer1.kubeconfig \
  set-credentials developer1 \
  --client-certificate=developer1.crt \
  --client-key=developer1.key \
  --embed-certs=true
```

Create context:

```bash
kubectl config \
  --kubeconfig=developer1.kubeconfig \
  set-context developer1@minikube \
  --cluster=minikube \
  --user=developer1 \
  --namespace=development
```

Activate context:

```bash
kubectl config \
  --kubeconfig=developer1.kubeconfig \
  use-context developer1@minikube
```

Inspect:

```bash
kubectl config \
  --kubeconfig=developer1.kubeconfig \
  view
```

---

# 19. Test Authentication Before RBAC

Run:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth whoami
```

Expected identity should include:

```text
developer1
developers
system:authenticated
```

Now try:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  get pods \
  -n development
```

Expected:

```text
Forbidden
```

This is correct.

The user is authenticated:

```text
Authentication = SUCCESS
```

but has not yet been authorized:

```text
Authorization = DENIED
```

---

# 20. Create developer Role

Create:

```bash
cat > developer-role.yaml <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: developer-role
  namespace: development
rules:

- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch

- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
EOF
```

Apply:

```bash
kubectl apply \
  -f developer-role.yaml
```

Check:

```bash
kubectl get role \
  developer-role \
  -n development
```

Describe:

```bash
kubectl describe role \
  developer-role \
  -n development
```

---

# 21. Role Alone Does Not Give Access

Try again:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  get pods \
  -n development
```

It should still return:

```text
Forbidden
```

Why?

Because:

```text
Role
=
permission definition
```

It has not yet been assigned to developer1.

We need:

```text
developer1
     |
     v
RoleBinding
     |
     v
developer-role
```

---

# 22. Create developer1 RoleBinding

Create:

```bash
cat > developer1-rolebinding.yaml <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: developer1-binding
  namespace: development
subjects:
- kind: User
  name: developer1
  apiGroup: rbac.authorization.k8s.io

roleRef:
  kind: Role
  name: developer-role
  apiGroup: rbac.authorization.k8s.io
EOF
```

Apply:

```bash
kubectl apply \
  -f developer1-rolebinding.yaml
```

Check:

```bash
kubectl get rolebinding \
  developer1-binding \
  -n development
```

---

# 23. Test developer1 Allowed Actions

List Pods:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  get pods \
  -n development
```

Should work.

List Deployments:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  get deployments \
  -n development
```

Should work.

Create a Deployment:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  create deployment app1 \
  --image=nginx \
  -n development
```

Should work.

Delete Deployment:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  delete deployment app1 \
  -n development
```

Should work.

---

# 24. Test developer1 Denied Actions

Try QA Pods:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  get pods \
  -n qa
```

Expected:

```text
Forbidden
```

Try kube-system:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  get pods \
  -n kube-system
```

Expected:

```text
Forbidden
```

Try Nodes:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  get nodes
```

Expected:

```text
Forbidden
```

Try ClusterRoles:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  get clusterroles
```

Expected:

```text
Forbidden
```

---

# 25. Verify with kubectl auth can-i

Can developer1 list Pods?

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth can-i list pods \
  -n development
```

Expected:

```text
yes
```

Can developer1 create Deployments?

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth can-i create deployments.apps \
  -n development
```

Expected:

```text
yes
```

Can developer1 delete Pods?

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth can-i delete pods \
  -n development
```

Expected:

```text
no
```

Can developer1 list Nodes?

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth can-i list nodes
```

Expected:

```text
no
```

Can developer1 access QA?

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth can-i list pods \
  -n qa
```

Expected:

```text
no
```

---

# 26. User 2: qauser1

Now create a second user who should have read-only access to the `qa` namespace.

Required:

```text
qauser1

Namespace:
qa

CAN:
get/list/watch Pods
get/list/watch Deployments
get/list/watch Services

CANNOT:
create
update
patch
delete
access development
access Nodes
```

---

# 27. Generate qauser1 Certificate

Private key:

```bash
openssl genrsa \
  -out qauser1.key \
  3072
```

CSR:

```bash
openssl req \
  -new \
  -key qauser1.key \
  -out qauser1.csr \
  -subj "/CN=qauser1/O=qa-team"
```

Encode CSR on macOS:

```bash
CSR_BASE64=$(cat qauser1.csr | base64 | tr -d '\n')
```

On Linux:

```bash
CSR_BASE64=$(base64 -w 0 qauser1.csr)
```

Create Kubernetes CSR:

```bash
cat > qauser1-k8s-csr.yaml <<EOF
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: qauser1
spec:
  request: ${CSR_BASE64}
  signerName: kubernetes.io/kube-apiserver-client
  expirationSeconds: 31536000
  usages:
  - client auth
EOF
```

Apply:

```bash
kubectl apply \
  -f qauser1-k8s-csr.yaml
```

Approve:

```bash
kubectl certificate approve qauser1
```

Download:

```bash
kubectl get csr qauser1 \
  -o jsonpath='{.status.certificate}' \
  | base64 --decode \
  > qauser1.crt
```

---

# 28. Create qauser1 Kubeconfig

```bash
kubectl config \
  --kubeconfig=qauser1.kubeconfig \
  set-cluster minikube \
  --server="$SERVER" \
  --certificate-authority=minikube-ca.crt \
  --embed-certs=true
```

Credentials:

```bash
kubectl config \
  --kubeconfig=qauser1.kubeconfig \
  set-credentials qauser1 \
  --client-certificate=qauser1.crt \
  --client-key=qauser1.key \
  --embed-certs=true
```

Context:

```bash
kubectl config \
  --kubeconfig=qauser1.kubeconfig \
  set-context qauser1@minikube \
  --cluster=minikube \
  --user=qauser1 \
  --namespace=qa
```

Use:

```bash
kubectl config \
  --kubeconfig=qauser1.kubeconfig \
  use-context qauser1@minikube
```

---

# 29. Create QA Read-Only Role

Create:

```bash
cat > qa-viewer-role.yaml <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: qa-viewer
  namespace: qa
rules:

- apiGroups:
  - ""
  resources:
  - pods
  - services
  verbs:
  - get
  - list
  - watch

- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - get
  - list
  - watch
EOF
```

Apply:

```bash
kubectl apply \
  -f qa-viewer-role.yaml
```

---

# 30. Bind QA Role to qauser1

```bash
cat > qauser1-rolebinding.yaml <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: qauser1-viewer-binding
  namespace: qa

subjects:
- kind: User
  name: qauser1
  apiGroup: rbac.authorization.k8s.io

roleRef:
  kind: Role
  name: qa-viewer
  apiGroup: rbac.authorization.k8s.io
EOF
```

Apply:

```bash
kubectl apply \
  -f qauser1-rolebinding.yaml
```

---

# 31. Test qauser1

Allowed:

```bash
kubectl \
  --kubeconfig=qauser1.kubeconfig \
  get pods \
  -n qa
```

Allowed:

```bash
kubectl \
  --kubeconfig=qauser1.kubeconfig \
  get deployments \
  -n qa
```

Denied:

```bash
kubectl \
  --kubeconfig=qauser1.kubeconfig \
  create deployment test \
  --image=nginx \
  -n qa
```

Denied:

```bash
kubectl \
  --kubeconfig=qauser1.kubeconfig \
  get pods \
  -n development
```

Denied:

```bash
kubectl \
  --kubeconfig=qauser1.kubeconfig \
  get nodes
```

---

# 32. Cluster-Wide User: auditor1

Now create:

```text
auditor1
```

Requirements:

```text
CAN:
list/get/watch Pods in all namespaces
list/get/watch Nodes
list/get/watch Namespaces

CANNOT:
create workloads
delete workloads
modify RBAC
modify Nodes
```

This requires:

```text
ClusterRole
+
ClusterRoleBinding
```

---

# 33. Generate auditor1 Certificate

```bash
openssl genrsa \
  -out auditor1.key \
  3072
```

```bash
openssl req \
  -new \
  -key auditor1.key \
  -out auditor1.csr \
  -subj "/CN=auditor1/O=auditors"
```

macOS:

```bash
CSR_BASE64=$(cat auditor1.csr | base64 | tr -d '\n')
```

Linux:

```bash
CSR_BASE64=$(base64 -w 0 auditor1.csr)
```

CSR object:

```bash
cat > auditor1-k8s-csr.yaml <<EOF
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: auditor1
spec:
  request: ${CSR_BASE64}
  signerName: kubernetes.io/kube-apiserver-client
  expirationSeconds: 31536000
  usages:
  - client auth
EOF
```

Apply:

```bash
kubectl apply \
  -f auditor1-k8s-csr.yaml
```

Approve:

```bash
kubectl certificate approve auditor1
```

Download certificate:

```bash
kubectl get csr auditor1 \
  -o jsonpath='{.status.certificate}' \
  | base64 --decode \
  > auditor1.crt
```

---

# 34. Create auditor1 Kubeconfig

```bash
kubectl config \
  --kubeconfig=auditor1.kubeconfig \
  set-cluster minikube \
  --server="$SERVER" \
  --certificate-authority=minikube-ca.crt \
  --embed-certs=true
```

```bash
kubectl config \
  --kubeconfig=auditor1.kubeconfig \
  set-credentials auditor1 \
  --client-certificate=auditor1.crt \
  --client-key=auditor1.key \
  --embed-certs=true
```

```bash
kubectl config \
  --kubeconfig=auditor1.kubeconfig \
  set-context auditor1@minikube \
  --cluster=minikube \
  --user=auditor1
```

```bash
kubectl config \
  --kubeconfig=auditor1.kubeconfig \
  use-context auditor1@minikube
```

---

# 35. Create cluster-auditor ClusterRole

Create:

```bash
cat > cluster-auditor-role.yaml <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-auditor
rules:

- apiGroups:
  - ""
  resources:
  - pods
  - namespaces
  - nodes
  verbs:
  - get
  - list
  - watch

- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - get
  - list
  - watch
EOF
```

Apply:

```bash
kubectl apply \
  -f cluster-auditor-role.yaml
```

Check:

```bash
kubectl get clusterrole \
  cluster-auditor
```

---

# 36. Create ClusterRoleBinding

```bash
cat > auditor1-clusterrolebinding.yaml <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: auditor1-binding

subjects:
- kind: User
  name: auditor1
  apiGroup: rbac.authorization.k8s.io

roleRef:
  kind: ClusterRole
  name: cluster-auditor
  apiGroup: rbac.authorization.k8s.io
EOF
```

Apply:

```bash
kubectl apply \
  -f auditor1-clusterrolebinding.yaml
```

---

# 37. Test auditor1

Allowed:

```bash
kubectl \
  --kubeconfig=auditor1.kubeconfig \
  get nodes
```

Allowed:

```bash
kubectl \
  --kubeconfig=auditor1.kubeconfig \
  get namespaces
```

Allowed:

```bash
kubectl \
  --kubeconfig=auditor1.kubeconfig \
  get pods -A
```

Allowed:

```bash
kubectl \
  --kubeconfig=auditor1.kubeconfig \
  get deployments -A
```

Denied:

```bash
kubectl \
  --kubeconfig=auditor1.kubeconfig \
  create deployment test \
  --image=nginx \
  -n development
```

Denied:

```bash
kubectl \
  --kubeconfig=auditor1.kubeconfig \
  delete namespace qa
```

Denied:

```bash
kubectl \
  --kubeconfig=auditor1.kubeconfig \
  get clusterroles
```

unless you explicitly add RBAC resources to the ClusterRole.

---

# 38. Very Important Difference

Namespace-based access:

```text
developer1
     |
     v
RoleBinding
     |
     v
Role
     |
     v
development namespace only
```

Cluster-wide access:

```text
auditor1
     |
     v
ClusterRoleBinding
     |
     v
ClusterRole
     |
     v
whole cluster
```

---

# 39. ClusterRole Can Also Be Used Inside One Namespace

A powerful RBAC pattern is:

```text
ClusterRole
+
RoleBinding
```

Example:

```text
ClusterRole: view
      |
      v
RoleBinding
namespace=development
      |
      v
developer1
```

Although `view` is a ClusterRole, the RoleBinding limits the effect to the `development` namespace.

---

# 40. Built-In Kubernetes Roles

Check:

```bash
kubectl get clusterroles
```

Common built-in roles include:

```text
view
edit
admin
cluster-admin
```

Inspect:

```bash
kubectl describe clusterrole view
```

```bash
kubectl describe clusterrole edit
```

```bash
kubectl describe clusterrole admin
```

```bash
kubectl describe clusterrole cluster-admin
```

---

# 41. Give Built-In view Role Only in development

Example:

```bash
kubectl create rolebinding \
  developer1-view \
  --clusterrole=view \
  --user=developer1 \
  --namespace=development
```

Architecture:

```text
developer1
    |
    v
RoleBinding
development/developer1-view
    |
    v
ClusterRole
view

Effective scope:
development only
```

---

# 42. Cluster-Wide view Access

This is different:

```bash
kubectl create clusterrolebinding \
  developer1-global-view \
  --clusterrole=view \
  --user=developer1
```

Now `developer1` gets the `view` ClusterRole cluster-wide.

Remove it after testing:

```bash
kubectl delete clusterrolebinding \
  developer1-global-view
```

---

# 43. RBAC Combination Table

| Role Definition | Binding | Effective Scope |
|---|---|---|
| Role | RoleBinding | One namespace |
| ClusterRole | RoleBinding | One namespace |
| ClusterRole | ClusterRoleBinding | Entire cluster |
| Role | ClusterRoleBinding | Not valid |

---

# 44. Group-Based Access

Individual user bindings become difficult to maintain:

```text
developer1
developer2
developer3
developer4
```

A better model is:

```text
developers group
       |
       v
RoleBinding
       |
       v
developer-role
```

Remember that we created developer1 using:

```text
/CN=developer1/O=developers
```

Therefore the certificate contains:

```text
Username = developer1
Group    = developers
```

---

# 45. Bind the developers Group

Create:

```bash
cat > developers-group-binding.yaml <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: developers-group-binding
  namespace: development

subjects:
- kind: Group
  name: developers
  apiGroup: rbac.authorization.k8s.io

roleRef:
  kind: Role
  name: developer-role
  apiGroup: rbac.authorization.k8s.io
EOF
```

Apply:

```bash
kubectl apply \
  -f developers-group-binding.yaml
```

Now anyone authenticated as a member of the `developers` group receives the Role permissions in `development`.

---

# 46. Create developer2 and Reuse the Group

Private key:

```bash
openssl genrsa \
  -out developer2.key \
  3072
```

CSR:

```bash
openssl req \
  -new \
  -key developer2.key \
  -out developer2.csr \
  -subj "/CN=developer2/O=developers"
```

Notice:

```text
O=developers
```

The group is the same.

After issuing a certificate and creating developer2's kubeconfig, you do **not** need another user-specific RoleBinding if the group binding already exists.

Architecture:

```text
developer1 -----\
developer2 ------+---- developers
developer3 -----/          |
                          RoleBinding
                             |
                             v
                       developer-role
                             |
                             v
                       development
```

---

# 47. Multiple Groups

A client certificate can contain more than one organization entry.

For example:

```bash
openssl req \
  -new \
  -key user.key \
  -out user.csr \
  -subj "/CN=user1/O=developers/O=project-a"
```

Conceptually:

```text
Username:
user1

Groups:
developers
project-a
system:authenticated
```

You can then bind different groups to different permissions.

---

# 48. Example Team Design

A useful organizational structure:

```text
developers
   |
   v
developer-role
development namespace

qa-team
   |
   v
qa-viewer
qa namespace

operations
   |
   v
operations ClusterRole
selected namespaces or cluster

auditors
   |
   v
cluster-auditor
cluster-wide read only

cluster-admins
   |
   v
cluster-admin
whole cluster
```

---

# 49. Restrict Access to Specific Resources

Suppose a user should only read ConfigMaps:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: configmap-reader
  namespace: development
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
```

---

# 50. Restrict Access to One Named Resource

You can use `resourceNames`.

Example:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: selected-config-reader
  namespace: development
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  resourceNames:
  - application-config
  verbs:
  - get
```

This grants access only to:

```text
ConfigMap/application-config
```

for operations compatible with `resourceNames`.

---

# 51. Secrets Need Special Care

Do not casually give users:

```yaml
resources:
- secrets
verbs:
- get
- list
```

Reading Secrets can expose credentials, tokens, API keys and other sensitive values.

Follow least privilege.

---

# 52. Be Careful with exec Access

Giving access to:

```text
pods/exec
```

allows users to execute commands inside containers.

Example rule:

```yaml
- apiGroups:
  - ""
  resources:
  - pods/exec
  verbs:
  - create
```

Only grant this when necessary.

---

# 53. Be Careful with Logs

To allow logs:

```yaml
- apiGroups:
  - ""
  resources:
  - pods/log
  verbs:
  - get
```

This is separate from normal Pod permissions.

---

# 54. Example Developer Role with Logs

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: developer-role
  namespace: development
rules:

- apiGroups: [""]
  resources:
  - pods
  verbs:
  - get
  - list
  - watch

- apiGroups: [""]
  resources:
  - pods/log
  verbs:
  - get

- apiGroups: ["apps"]
  resources:
  - deployments
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
```

---

# 55. Example Developer Role with exec

If explicitly required:

```yaml
- apiGroups:
  - ""
  resources:
  - pods/exec
  verbs:
  - create
```

Then:

```bash
kubectl exec -it <pod> \
  -n development \
  -- /bin/sh
```

may be authorized.

Do not automatically grant `pods/exec`.

---

# 56. Service Access

To permit Services:

```yaml
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
```

---

# 57. ConfigMap Access

```yaml
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
```

---

# 58. Read-Only Namespace Role Example

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: namespace-reader
  namespace: development
rules:

- apiGroups:
  - ""
  resources:
  - pods
  - services
  - configmaps
  verbs:
  - get
  - list
  - watch

- apiGroups:
  - apps
  resources:
  - deployments
  - replicasets
  verbs:
  - get
  - list
  - watch
```

---

# 59. Full Developer Namespace Role Example

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: app-developer
  namespace: development
rules:

- apiGroups:
  - ""
  resources:
  - pods
  - services
  - configmaps
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete

- apiGroups:
  - ""
  resources:
  - pods/log
  verbs:
  - get

- apiGroups:
  - apps
  resources:
  - deployments
  - replicasets
  - statefulsets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
```

Add Secrets only if required.

---

# 60. Verify All Permissions for a User

Using the user's kubeconfig:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth can-i --list \
  -n development
```

This is one of the best RBAC troubleshooting commands.

---

# 61. Admin Impersonation

As administrator, test another identity without switching kubeconfigs:

```bash
kubectl auth can-i \
  list pods \
  --as=developer1 \
  -n development
```

Test QA:

```bash
kubectl auth can-i \
  list pods \
  --as=developer1 \
  -n qa
```

Test auditor:

```bash
kubectl auth can-i \
  list nodes \
  --as=auditor1
```

---

# 62. Test Group Permissions with Impersonation

```bash
kubectl auth can-i \
  list pods \
  --as=developer2 \
  --as-group=developers \
  -n development
```

Expected:

```text
yes
```

---

# 63. List RBAC Resources

Roles:

```bash
kubectl get roles -A
```

RoleBindings:

```bash
kubectl get rolebindings -A
```

ClusterRoles:

```bash
kubectl get clusterroles
```

ClusterRoleBindings:

```bash
kubectl get clusterrolebindings
```

---

# 64. Inspect Bindings

```bash
kubectl describe rolebinding \
  developer1-binding \
  -n development
```

```bash
kubectl describe rolebinding \
  qauser1-viewer-binding \
  -n qa
```

```bash
kubectl describe clusterrolebinding \
  auditor1-binding
```

---

# 65. Check Kubernetes CSRs

```bash
kubectl get csr
```

Detailed:

```bash
kubectl describe csr developer1
```

---

# 66. Certificate Expiration

Check:

```bash
openssl x509 \
  -in developer1.crt \
  -noout \
  -dates
```

Example:

```text
notBefore=...
notAfter=...
```

Client certificates should not be treated as permanent credentials.

For a production human-user system, OIDC is generally easier to manage than manually issuing long-lived certificates.

---

# 67. Revoking Certificate-Based Access

Kubernetes does not provide a simple CRL-based human-user lifecycle through RBAC for these lab certificates.

The easiest immediate authorization revocation is to remove the user's RBAC binding.

Example:

```bash
kubectl delete rolebinding \
  developer1-binding \
  -n development
```

After this:

```text
developer1
```

may still successfully authenticate with the certificate, but Kubernetes RBAC no longer grants those Role permissions.

Authentication:

```text
SUCCESS
```

Authorization:

```text
DENIED
```

If group-based bindings grant permissions, remove the relevant group membership at the identity layer or modify the bindings accordingly.

---

# 68. Remove Cluster-Wide Access

Delete:

```bash
kubectl delete clusterrolebinding \
  auditor1-binding
```

Now auditor1 should lose cluster-wide access.

Test:

```bash
kubectl \
  --kubeconfig=auditor1.kubeconfig \
  get nodes
```

Expected:

```text
Forbidden
```

---

# 69. Do Not Give cluster-admin to Normal Users

Avoid:

```bash
kubectl create clusterrolebinding \
  developer1-admin \
  --clusterrole=cluster-admin \
  --user=developer1
```

This gives extremely broad cluster permissions.

Only genuine cluster administrators should receive such privileges.

---

# 70. Least Privilege Model

Instead of:

```text
developer
    |
    v
cluster-admin
```

prefer:

```text
developer
    |
    v
RoleBinding
    |
    v
Application Role
    |
    v
One namespace
```

---

# 71. Example Access Matrix

| User | development | qa | Nodes | Cluster RBAC |
|---|---|---|---|---|
| developer1 | Developer | No | No | No |
| qauser1 | No | Read only | No | No |
| auditor1 | Read only | Read only | Read only | No modification |
| admin | Full | Full | Full | Full |

---

# 72. Recommended Permission Matrix

## Developer

```text
Pods:
get/list/watch

Pod logs:
get

Deployments:
get/list/watch/create/update/patch/delete

Services:
get/list/watch/create/update/patch/delete

ConfigMaps:
get/list/watch/create/update/patch/delete
```

Do not automatically allow:

```text
Secrets
RBAC modifications
Nodes
Namespaces
pods/exec
```

---

# 73. QA User

Recommended:

```text
Pods:
get/list/watch

Pod logs:
get

Deployments:
get/list/watch

Services:
get/list/watch
```

No create/update/delete.

---

# 74. Auditor

Recommended:

```text
get
list
watch
```

for selected cluster resources.

Avoid:

```text
create
update
patch
delete
```

---

# 75. RBAC Request Evaluation

Suppose developer1 runs:

```bash
kubectl get pods \
  -n development \
  --kubeconfig=developer1.kubeconfig
```

The flow is:

```text
developer1.kubeconfig
       |
       | certificate
       v
API Server
       |
       v
Authentication
       |
       | CN=developer1
       | O=developers
       v
Identity
       |
       v
RBAC Authorization
       |
       v
Find RoleBindings
       |
       v
development/developer1-binding
       |
       v
developer-role
       |
       v
Check:

verb      = list
resource  = pods
namespace = development
       |
       v
ALLOW
```

---

# 76. Same User Accessing QA

```bash
kubectl get pods \
  -n qa \
  --kubeconfig=developer1.kubeconfig
```

Evaluation:

```text
user      = developer1
verb      = list
resource  = pods
namespace = qa
```

The RoleBinding exists in:

```text
development
```

not:

```text
qa
```

Therefore:

```text
DENY
```

---

# 77. Important Security Note About Bind/Escalate

Be careful when allowing users to create or modify:

```text
Roles
ClusterRoles
RoleBindings
ClusterRoleBindings
```

RBAC itself includes protections against privilege escalation, including restrictions related to binding roles and escalating role permissions.

Do not give ordinary application developers broad RBAC administration permissions unless intentionally required.

---

# 78. Cleanup developer1

```bash
kubectl delete rolebinding \
  developer1-binding \
  -n development
```

```bash
kubectl delete role \
  developer-role \
  -n development
```

```bash
kubectl delete csr developer1
```

Local files:

```bash
rm -f developer1.key
rm -f developer1.csr
rm -f developer1.crt
rm -f developer1.kubeconfig
```

---

# 79. Cleanup qauser1

```bash
kubectl delete rolebinding \
  qauser1-viewer-binding \
  -n qa
```

```bash
kubectl delete role \
  qa-viewer \
  -n qa
```

```bash
kubectl delete csr qauser1
```

---

# 80. Cleanup auditor1

```bash
kubectl delete clusterrolebinding \
  auditor1-binding
```

```bash
kubectl delete clusterrole \
  cluster-auditor
```

```bash
kubectl delete csr auditor1
```

---

# 81. Delete Namespaces

```bash
kubectl delete namespace development
kubectl delete namespace qa
```

---

# 82. Delete Minikube

To remove the entire cluster:

```bash
minikube delete
```

Check:

```bash
minikube status
```

---

# 83. Final RBAC Mental Model

```text
                 AUTHENTICATION

Certificate / OIDC / Token
           |
           v
      User Identity
           |
           v
      Authorization
           |
           v
          RBAC
```

Namespace permissions:

```text
User / Group
     |
     v
RoleBinding
     |
     v
Role
     |
     v
Namespace
```

Reusable namespace permissions:

```text
User / Group
     |
     v
RoleBinding
     |
     v
ClusterRole
     |
     v
RoleBinding namespace only
```

Cluster-wide permissions:

```text
User / Group
     |
     v
ClusterRoleBinding
     |
     v
ClusterRole
     |
     v
Entire Cluster
```

---

# 84. Recommended Real-World Design

For learning:

```text
X.509 certificates
+
RBAC
```

is excellent because the full authentication and authorization flow is visible.

For a larger organization, prefer:

```text
Identity Provider
Keycloak / Microsoft Entra ID / Okta / another OIDC provider
        |
        v
OIDC
        |
        v
Groups
        |
        v
Kubernetes RBAC
```

Example:

```text
developers
    -> development RoleBinding

qa-team
    -> qa RoleBinding

auditors
    -> read-only ClusterRoleBinding

cluster-admins
    -> carefully controlled administrative privileges
```

---

# 85. Most Important Commands

Create user key:

```bash
openssl genrsa -out developer1.key 3072
```

Create CSR:

```bash
openssl req \
  -new \
  -key developer1.key \
  -out developer1.csr \
  -subj "/CN=developer1/O=developers"
```

Create Kubernetes CSR:

```bash
kubectl apply -f developer1-k8s-csr.yaml
```

Approve:

```bash
kubectl certificate approve developer1
```

Get certificate:

```bash
kubectl get csr developer1 \
  -o jsonpath='{.status.certificate}' \
  | base64 --decode \
  > developer1.crt
```

Check identity:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth whoami
```

Test permissions:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth can-i --list \
  -n development
```

Test one permission:

```bash
kubectl \
  --kubeconfig=developer1.kubeconfig \
  auth can-i create deployments.apps \
  -n development
```

---

# 86. Complete Lab Summary

You created:

```text
developer1
  |
  +-- certificate authentication
  |
  +-- developer1.kubeconfig
  |
  +-- RoleBinding
  |
  +-- developer-role
  |
  +-- development namespace only
```

```text
qauser1
  |
  +-- certificate authentication
  |
  +-- qauser1.kubeconfig
  |
  +-- RoleBinding
  |
  +-- qa-viewer
  |
  +-- qa namespace read-only
```

```text
auditor1
  |
  +-- certificate authentication
  |
  +-- auditor1.kubeconfig
  |
  +-- ClusterRoleBinding
  |
  +-- cluster-auditor
  |
  +-- cluster-wide read-only access
```

This is the essential RBAC pattern for controlling user access in Kubernetes.

---

# 87. Quick Reference

```text
Role
    Defines permissions in a namespace

RoleBinding
    Assigns Role/ClusterRole permissions in one namespace

ClusterRole
    Defines reusable or cluster-scoped permissions

ClusterRoleBinding
    Assigns ClusterRole permissions across the cluster

CN in client certificate
    Kubernetes username

O in client certificate
    Kubernetes group

kubectl auth whoami
    Shows authenticated identity

kubectl auth can-i
    Tests authorization
```

---

# 88. Recommended Training Order

```text
1. Start Minikube
2. Create namespaces
3. Create developer1 private key
4. Create developer1 CSR
5. Create Kubernetes CertificateSigningRequest
6. Admin approves CSR
7. Download issued certificate
8. Create dedicated kubeconfig
9. Verify authentication
10. Observe Forbidden before RBAC
11. Create Role
12. Observe Role alone is insufficient
13. Create RoleBinding
14. Verify allowed operations
15. Verify denied namespace access
16. Create qauser1
17. Give QA read-only access
18. Create auditor1
19. Create ClusterRole
20. Create ClusterRoleBinding
21. Verify cluster-wide read-only access
22. Create group-based bindings
23. Use kubectl auth can-i
24. Remove bindings and observe immediate loss of authorization
```

That sequence makes the distinction between authentication, authorization, namespace scope, and cluster scope very clear.
