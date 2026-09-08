# Project: Highly Available Kubernetes Cluster with etcd Backup, OpenEBS Storage, and Observability

## Project Objective

The objective of this project is to build a **highly available Kubernetes cluster in a local lab environment** and implement production-oriented capabilities such as:

- Multi-master Kubernetes architecture
- etcd high availability
- Automated etcd backup
- etcd disaster recovery and restoration
- Persistent storage using OpenEBS
- Metrics monitoring using Prometheus
- Visualization using Grafana
- Centralized logging using Loki

The complete environment will contain **six virtual machines/nodes**.

---

## Cluster Architecture

```text
                    Kubernetes Cluster

               +-------------------------+
               |     Control Plane       |
               +-------------------------+

        +-------------+ +-------------+ +-------------+
        |  Master-1   | |  Master-2   | |  Master-3   |
        | API Server  | | API Server  | | API Server  |
        | Scheduler   | | Scheduler   | | Scheduler   |
        | Controller  | | Controller  | | Controller  |
        | etcd        | | etcd        | | etcd        |
        +-------------+ +-------------+ +-------------+
               \             |             /
                \            |            /
                 +-----------------------+
                 |     etcd Cluster      |
                 +-----------------------+

                         |
                         |
                 Kubernetes Networking
                         |
                         |

        +-------------+ +-------------+ +-------------+
        |  Worker-1   | |  Worker-2   | |  Worker-3   |
        | kubelet     | | kubelet     | | kubelet     |
        | containerd  | | containerd  | | containerd  |
        | Workloads   | | Workloads   | | Workloads   |
        +-------------+ +-------------+ +-------------+
```

---

# 1. Local Infrastructure Setup

Create **six Linux virtual machines** on the local computers.

The nodes will be:

```text
master-1
master-2
master-3

worker-1
worker-2
worker-3
```

All six systems must be connected through the same network so that every node can communicate with the other nodes.

Verify connectivity using commands such as:

```bash
ping master-1
ping master-2
ping master-3
ping worker-1
ping worker-2
ping worker-3
```

Static IP addresses or consistent hostnames should preferably be configured.

---

# 2. Kubernetes High Availability Cluster

Build a Kubernetes cluster with:

```text
3 Control Plane Nodes
3 Worker Nodes
```

The control-plane nodes will run components including:

```text
kube-apiserver
kube-controller-manager
kube-scheduler
etcd
```

The worker nodes will primarily run:

```text
kubelet
container runtime
application workloads
```

A load balancer or virtual endpoint should ideally be placed in front of the Kubernetes API servers so that Kubernetes clients and worker nodes do not depend on a single master node.

Example architecture:

```text
kubectl / workers
        |
        v

Load Balancer / API Endpoint
        |
   ---------------------
   |        |          |
master-1 master-2  master-3
```

---

# 3. etcd High Availability

Because the cluster contains three control-plane nodes, etcd will operate as a **three-member cluster**.

Example:

```text
etcd-1
  |
  +------ etcd-2
  |
  +------ etcd-3
```

etcd stores critical Kubernetes cluster information, including:

```text
Pods
Deployments
Services
ConfigMaps
Secrets
Namespaces
RBAC configuration
Node information
Cluster configuration
```

Therefore, protecting etcd is one of the most important parts of the project.

---

# 4. Automated etcd Backup

Create an automated backup mechanism for etcd.

The backup must run periodically, for example:

```text
Every day at 2:00 AM
```

A Linux cron job can trigger the backup process.

Conceptually:

```text
Cron
 |
 |
 v
etcd backup script
 |
 |
 v
etcd snapshot
 |
 |
 v
Backup storage
```

The backup script should generate timestamped snapshots such as:

```text
etcd-snapshot-2026-09-04.db
etcd-snapshot-2026-09-05.db
etcd-snapshot-2026-09-06.db
```

The project should also define a backup retention policy.

For example:

```text
Keep the latest 7 backups
Delete backups older than 7 days
```

---

# 5. etcd Disaster Recovery

After backup functionality is verified, simulate an etcd disaster.

For example:

```text
1. Create Kubernetes resources.
2. Take an etcd snapshot.
3. Verify the snapshot.
4. Delete selected Kubernetes resources.

or, in an isolated lab recovery exercise:

5. Simulate loss/corruption of the etcd data directory.
6. Stop the affected control-plane components.
7. Restore etcd from the snapshot.
8. Restart the cluster.
9. Verify that Kubernetes state has been recovered.
```

The exercise demonstrates the complete disaster-recovery lifecycle:

```text
Running Cluster
      |
      v
etcd Snapshot
      |
      v
Failure Simulation
      |
      v
Restore Snapshot
      |
      v
Restart Control Plane
      |
      v
Cluster Recovered
```

Backup and restore operations should use the appropriate etcd utilities available with the installed etcd version, such as `etcdctl` for snapshot creation/status operations and the supported restore utility for that release.

---

# 6. OpenEBS Storage

Install **OpenEBS** as the Kubernetes storage platform.

Applications should use Kubernetes persistent storage through:

```text
StorageClass
     |
     v
PersistentVolumeClaim
     |
     v
PersistentVolume
     |
     v
OpenEBS Storage
```

Example application flow:

```text
Application Pod
      |
      v
PVC
      |
      v
StorageClass
      |
      v
OpenEBS
      |
      v
Underlying Node Storage
```

Students should deploy at least one stateful application and verify that its data survives pod recreation.

Possible examples include:

```text
PostgreSQL
MySQL
NGINX with persistent content
File-upload application
```

---

# 7. Prometheus Monitoring

Install Prometheus to collect metrics from the Kubernetes cluster.

Prometheus should monitor areas such as:

```text
Node CPU
Node memory
Disk usage
Network utilization
Pod status
Pod CPU
Pod memory
Kubernetes API server
etcd
Kubernetes control-plane components
```

Architecture:

```text
Kubernetes Components
       |
       v
Metrics Endpoints
       |
       v
Prometheus
       |
       v
Time-Series Database
```

---

# 8. Grafana Visualization

Install Grafana and integrate it with Prometheus.

Grafana dashboards should display important cluster information such as:

```text
Cluster CPU usage
Cluster memory usage
Worker-node utilization
Pod utilization
Disk/storage usage
API server metrics
etcd metrics
OpenEBS metrics where available
```

Architecture:

```text
Kubernetes
    |
    v
Prometheus
    |
    v
Grafana
    |
    v
Dashboards
```

---

# 9. Loki Centralized Logging

Install Loki to provide centralized Kubernetes logging.

A compatible log collector should collect container logs and send them to Loki.

Conceptually:

```text
Application Pods
       |
       v
Log Collector
       |
       v
Loki
       |
       v
Grafana
```

Using Grafana, users should be able to search logs based on attributes such as:

```text
Namespace
Pod
Container
Application
Node
```

---

# 10. Complete Project Architecture

```text
                     Developers / Administrators
                               |
                               v
                     Kubernetes API Endpoint
                               |
                    +----------------------+
                    |    Load Balancer     |
                    +----------------------+
                         /      |      \
                        /       |       \

                 Master-1   Master-2   Master-3
                    |          |          |
                  etcd-1     etcd-2     etcd-3
                     \         |         /
                      \        |        /
                       +---------------+
                       | etcd Cluster  |
                       +---------------+

                              |
                        Kubernetes
                              |
             ---------------------------------
             |               |               |
          Worker-1        Worker-2        Worker-3
             |               |               |
             +---------------+---------------+
                             |
                         Workloads
                             |
                         OpenEBS
                             |
                    Persistent Storage


                     Observability Stack

               +-------------------------+
               |       Prometheus        |
               +-------------------------+
                           |
                           v
               +-------------------------+
               |        Grafana          |
               +-------------------------+

Application Logs
      |
      v
Log Collector
      |
      v
Loki
      |
      v
Grafana
```

---

# Expected Deliverables

At the end of the project, the team should demonstrate:

1. A working **six-node Kubernetes cluster**.
2. Three working control-plane nodes.
3. Three working worker nodes.
4. A healthy three-member etcd cluster.
5. Successful deployment of applications.
6. OpenEBS dynamic persistent-volume provisioning.
7. Successful etcd snapshot creation.
8. Automatic daily etcd backups using cron.
9. Backup retention and cleanup.
10. Verification of etcd snapshots.
11. A controlled etcd disaster simulation.
12. Successful restoration of Kubernetes state from an etcd backup.
13. Prometheus collecting Kubernetes and node metrics.
14. Grafana dashboards displaying cluster metrics.
15. Loki collecting application/container logs.
16. Grafana querying and displaying Loki logs.

---

# Final Goal

The project should demonstrate more than simply installing Kubernetes.

It should demonstrate how a Kubernetes environment is **built, operated, monitored, protected, and recovered**.

The major areas covered are:

```text
High Availability
        +
Persistent Storage
        +
Backup
        +
Disaster Recovery
        +
Monitoring
        +
Logging
        =
Production-Oriented Kubernetes Platform
```

By completing the project, participants will gain practical experience operating the major building blocks required for a production-style Kubernetes environment.

---

## Important Design Note

**etcd backup is not the same as OpenEBS application storage backup.**

They should be treated as separate recovery domains:

- **etcd backup** protects the Kubernetes control-plane state.
- **OpenEBS storage** provides persistent application data and should have its own backup/recovery strategy.

Both should eventually have separate failure and recovery exercises.
