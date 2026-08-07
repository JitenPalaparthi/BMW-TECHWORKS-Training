
# Docker Container Creation Flow with Network and Volume

This document explains, step by step, what happens internally when you create a Docker network, create a Docker volume, and start a container attached to both.

The example uses the following commands:

```bash
docker network create app-net

docker volume create app-data

docker run -d \
  --name app-container \
  --network app-net \
  --mount source=app-data,target=/data \
  -p 8080:80 \
  nginx:latest
```

---

# 1. High-Level Flow

```text
User
  |
  | docker run ...
  v
Docker CLI
  |
  | Docker Engine API
  v
dockerd
  |
  +--> Image Manager
  |
  +--> Network Manager
  |
  +--> Volume Manager
  |
  +--> containerd
          |
          v
        shim
          |
          v
        runc
          |
          v
Linux Kernel
  |
  +--> namespaces
  +--> cgroups
  +--> mounts
  +--> veth pair
  +--> bridge
  +--> routing/NAT
  +--> security policies
  |
  v
Container process starts
```

---

# 2. Step 1: The User Runs `docker network create`

```bash
docker network create app-net
```

## 2.1 Docker CLI parses the command

The Docker CLI:

1. Reads the command-line arguments.
2. Identifies the requested operation as network creation.
3. Creates a Docker Engine API request.
4. Sends the request to the Docker daemon.

On Linux, the CLI usually communicates through:

```text
/var/run/docker.sock
```

This is a Unix domain socket.

The request is conceptually similar to:

```http
POST /networks/create
```

with a payload containing:

```json
{
  "Name": "app-net",
  "Driver": "bridge"
}
```

If no network driver is specified, Docker normally creates a user-defined bridge network.

---

# 3. Step 2: `dockerd` Receives the Network Request

The Docker daemon, `dockerd`, receives the API request.

It performs the following operations:

1. Validates the network name.
2. Checks whether a network with that name already exists.
3. Chooses the network driver.
4. Allocates an IP subnet.
5. Creates network metadata.
6. Configures host networking resources.

For a bridge network, Docker may create a Linux bridge device similar to:

```text
br-<network-id>
```

Example:

```text
br-a13f02c7d880
```

The exact identifier depends on the network ID.

---

# 4. Step 3: Docker Configures the Bridge Network

For a user-defined bridge network, Docker creates or configures:

- A Linux bridge
- An IP subnet
- A gateway IP
- Routing rules
- Firewall/NAT rules
- Internal DNS support

Example logical configuration:

```text
Network: app-net
Subnet: 172.20.0.0/16
Gateway: 172.20.0.1
```

The host bridge may receive:

```text
172.20.0.1
```

Containers attached to the network may receive addresses such as:

```text
172.20.0.2
172.20.0.3
172.20.0.4
```

Docker also stores the network's metadata in its local state.

Inspect it with:

```bash
docker network inspect app-net
```

---

# 5. Step 4: The User Creates a Volume

```bash
docker volume create app-data
```

The Docker CLI sends another API request to `dockerd`.

Conceptually:

```http
POST /volumes/create
```

with:

```json
{
  "Name": "app-data",
  "Driver": "local"
}
```

---

# 6. Step 5: Docker Volume Manager Creates the Volume

The Docker volume manager handles the request.

For the default local volume driver, Docker:

1. Validates the volume name.
2. Creates volume metadata.
3. Creates a directory on the host.
4. Associates the directory with the volume name.

On a typical Linux Docker installation, the host path is similar to:

```text
/var/lib/docker/volumes/app-data/_data
```

Important:

```text
Volume name:
app-data

Host storage path:
/var/lib/docker/volumes/app-data/_data

Container mount point:
/data
```

The volume exists independently of the container.

Deleting the container does not automatically delete the volume.

Inspect it with:

```bash
docker volume inspect app-data
```

---

# 7. Step 6: The User Runs the Container

```bash
docker run -d \
  --name app-container \
  --network app-net \
  --mount source=app-data,target=/data \
  -p 8080:80 \
  nginx:latest
```

The CLI parses:

```text
-d
Run in detached mode

--name app-container
Assign a container name

--network app-net
Attach the container to app-net

--mount source=app-data,target=/data
Mount app-data at /data

-p 8080:80
Publish host port 8080 to container port 80

nginx:latest
Use the nginx image
```

The Docker CLI then communicates with Docker Engine through the API.

Logically, `docker run` performs two major operations:

```text
docker create
docker start
```

---

# 8. Step 7: Docker Creates the Container Object

The CLI first sends a container creation request.

Conceptually:

```http
POST /containers/create?name=app-container
```

The request includes configuration such as:

```json
{
  "Image": "nginx:latest",
  "HostConfig": {
    "NetworkMode": "app-net",
    "PortBindings": {
      "80/tcp": [
        {
          "HostPort": "8080"
        }
      ]
    },
    "Mounts": [
      {
        "Type": "volume",
        "Source": "app-data",
        "Target": "/data"
      }
    ]
  }
}
```

At this stage, the container process has not yet started.

Docker creates an internal container configuration object.

---

# 9. Step 8: Docker Checks Whether the Image Exists

Docker checks its local image store for:

```text
nginx:latest
```

Two possibilities exist.

## Case 1: Image exists locally

Docker uses the existing image metadata and layers.

## Case 2: Image does not exist locally

Docker contacts a registry, usually Docker Hub.

The flow becomes:

```text
dockerd
   |
   v
Registry API
   |
   v
Manifest
   |
   v
Image configuration
   |
   v
Image layers
```

Docker verifies the content using digests.

For example:

```text
sha256:<digest>
```

Each layer is content-addressed.

---

# 10. Step 9: Image Layers Are Unpacked

The image consists of immutable read-only layers.

Example:

```text
Layer 1: Base filesystem
Layer 2: Nginx packages
Layer 3: Configuration files
Layer 4: Entrypoint scripts
```

Docker delegates lower-level image and snapshot handling to `containerd`.

The storage driver or snapshotter prepares the root filesystem.

On Linux, this is commonly based on OverlayFS.

Conceptually:

```text
LowerDir:
  Image read-only layers

UpperDir:
  Container writable layer

WorkDir:
  OverlayFS internal working directory

MergedDir:
  Final filesystem visible to the container
```

The container sees one merged filesystem.

---

# 11. Step 10: Docker Creates the Writable Container Layer

The image is immutable.

When the container is created, Docker adds a writable layer above the image layers.

```text
Container writable layer
------------------------
Nginx image layer
Linux library layer
Base filesystem layer
```

Any file created or modified inside the container is written to this writable layer unless that path is backed by a volume or bind mount.

For example:

```bash
echo hello > /tmp/test.txt
```

is normally written into the writable container layer.

But:

```bash
echo hello > /data/test.txt
```

is written into the Docker volume because `/data` is mounted from `app-data`.

---

# 12. Step 11: Docker Resolves the Volume Mount

Docker validates:

```text
Source volume: app-data
Target path: /data
```

It confirms that the volume exists.

Docker then prepares the mount mapping:

```text
Host:
/var/lib/docker/volumes/app-data/_data

Container:
/data
```

When the container mount namespace is created, the host volume directory is mounted at `/data`.

From inside the container:

```bash
ls /data
```

shows the contents of the Docker volume.

From the host:

```bash
sudo ls /var/lib/docker/volumes/app-data/_data
```

shows the same underlying data.

---

# 13. Step 12: Docker Prepares the Container Network

Docker looks up:

```text
app-net
```

It obtains:

- Network ID
- Subnet
- Gateway
- Driver
- DNS settings
- Existing endpoints

Docker allocates an available container IP address.

Example:

```text
Container IP: 172.20.0.2
Gateway: 172.20.0.1
```

Docker creates a network endpoint for the container.

---

# 14. Step 13: Docker Creates the Network Namespace

A network namespace gives the container an isolated network stack.

The container gets its own:

- Network interfaces
- IP addresses
- Routing table
- ARP table
- Firewall view
- Loopback interface
- Port space

The host and container do not directly share the same network namespace.

Conceptually:

```text
Host network namespace
  |
  +--> eth0
  +--> docker0 or br-xxxx
  +--> veth-host
           |
           | virtual Ethernet connection
           |
         veth-container
              |
              v
Container network namespace
  |
  +--> eth0
  +--> lo
```

---

# 15. Step 14: Docker Creates a `veth` Pair

Docker creates a virtual Ethernet pair.

A `veth` pair works like a virtual cable:

```text
veth-host <----------------> veth-container
```

One end remains on the host.

The other end is moved into the container's network namespace.

Inside the container, the interface is usually renamed:

```text
eth0
```

The host-side interface is attached to the Docker bridge:

```text
br-<network-id>
```

---

# 16. Step 15: Docker Assigns the Container IP Address

Inside the container's network namespace, Docker configures:

```text
Interface: eth0
IP: 172.20.0.2
Subnet: 172.20.0.0/16
Gateway: 172.20.0.1
```

It also brings up:

```text
lo
eth0
```

The container routing table may logically contain:

```text
default via 172.20.0.1 dev eth0
172.20.0.0/16 dev eth0
```

---

# 17. Step 16: Docker Configures Container DNS

User-defined Docker networks provide built-in container name resolution.

Docker configures DNS so that containers on the same network can resolve each other by name.

For example, another container on `app-net` may be able to resolve:

```text
app-container
```

to:

```text
172.20.0.2
```

Inside the container, `/etc/resolv.conf` is configured to use Docker's embedded DNS service.

A common internal DNS address is:

```text
127.0.0.11
```

---

# 18. Step 17: Docker Configures Port Publishing

The command contains:

```bash
-p 8080:80
```

This means:

```text
Host port: 8080
Container port: 80
```

Traffic arriving at:

```text
HostIP:8080
```

is forwarded to:

```text
172.20.0.2:80
```

Docker configures host networking rules using mechanisms such as:

- iptables
- nftables
- NAT
- routing
- Docker userland proxy in some configurations

Logical flow:

```text
Client
  |
  v
Host:8080
  |
  v
NAT / forwarding rule
  |
  v
Container:172.20.0.2:80
```

---

# 19. Step 18: Docker Builds the OCI Runtime Configuration

Docker and `containerd` create an OCI runtime bundle.

The bundle contains:

```text
config.json
root filesystem reference
runtime settings
```

The OCI runtime configuration contains details such as:

- Command
- Environment variables
- Working directory
- User
- Root filesystem
- Namespace configuration
- cgroup settings
- Mounts
- Capabilities
- Seccomp rules
- Hostname

A simplified configuration may conceptually look like:

```json
{
  "process": {
    "args": [
      "/docker-entrypoint.sh",
      "nginx",
      "-g",
      "daemon off;"
    ],
    "cwd": "/",
    "capabilities": {}
  },
  "root": {
    "path": "rootfs"
  },
  "mounts": [
    {
      "destination": "/data",
      "type": "bind",
      "source": "/var/lib/docker/volumes/app-data/_data"
    }
  ],
  "linux": {
    "namespaces": [
      { "type": "pid" },
      { "type": "network" },
      { "type": "mount" },
      { "type": "ipc" },
      { "type": "uts" }
    ]
  }
}
```

---

# 20. Step 19: Docker Hands the Container to `containerd`

`dockerd` does not directly call Linux system calls for every low-level container operation.

It delegates the execution lifecycle to `containerd`.

The flow is:

```text
dockerd
   |
   v
containerd
```

`containerd` manages:

- Container metadata
- Images
- Snapshots
- Tasks
- Runtime selection
- Container lifecycle

A container definition and a running task are different concepts.

```text
Container:
Metadata and configuration

Task:
Running operating-system process
```

---

# 21. Step 20: `containerd` Starts a Shim

`containerd` starts a per-container runtime shim.

Conceptually:

```text
containerd
   |
   v
containerd-shim-runc-v2
```

The shim is important because it:

- Keeps container standard input/output connected
- Collects the exit status
- Allows the container to continue running if `containerd` restarts
- Acts as an intermediary between `containerd` and the OCI runtime
- Manages the running process after `runc` exits

The shim is a long-running process.

`runc` is normally short-lived.

---

# 22. Step 21: The Shim Invokes `runc`

The shim invokes the OCI runtime.

By default, this is usually:

```text
runc
```

Other possible runtimes include:

```text
crun
Kata Containers runtime
gVisor runsc
NVIDIA runtime integration
```

`runc` reads the OCI runtime bundle and creates the container.

---

# 23. Step 22: `runc` Requests Kernel Isolation

`runc` uses Linux system calls and kernel features to create the container environment.

The kernel creates the required namespaces.

## PID namespace

Provides isolated process IDs.

Inside the container, the Nginx process may appear as:

```text
PID 1
```

On the host, the same process may have a different PID:

```text
PID 24571
```

## Mount namespace

Provides an isolated mount table.

The container sees:

- Its merged image filesystem
- `/data` mounted from the Docker volume
- `/proc`
- `/sys`
- `/dev`

## Network namespace

Provides isolated network interfaces and routing.

## UTS namespace

Provides an isolated hostname.

## IPC namespace

Provides isolated shared memory, semaphores, and message queues.

## User namespace

If configured, maps container users to different host user IDs.

---

# 24. Step 23: `runc` Configures cgroups

Linux cgroups control and account for resource usage.

Docker may configure:

- CPU quota
- CPU shares
- Memory limit
- PIDs limit
- Block I/O controls
- Device access

Example command:

```bash
docker run \
  --memory 512m \
  --cpus 1.5 \
  nginx
```

Docker translates these settings into cgroup configuration.

The kernel enforces them.

Without explicit limits, the container may use host resources according to the default cgroup settings.

---

# 25. Step 24: Security Policies Are Applied

Before the process starts, Docker and the runtime apply security controls.

## Linux capabilities

Docker drops many root capabilities by default.

The container receives only a limited capability set.

## Seccomp

Docker applies a seccomp profile to restrict dangerous system calls.

## AppArmor

On supported systems, Docker may apply an AppArmor profile.

## SELinux

On SELinux-enabled systems, labels may be assigned to container processes and files.

## Read-only masks and protected paths

Sensitive kernel paths may be:

- Read-only
- Hidden
- Masked
- Restricted

---

# 26. Step 25: The Root Filesystem Is Mounted

The container's mount namespace is populated.

Conceptually:

```text
/
├── bin
├── etc
├── usr
├── var
├── proc
├── sys
├── dev
└── data
```

The root filesystem comes from the merged image layers plus the writable layer.

The `/data` path comes from the Docker volume.

The container sees a unified filesystem, even though it is assembled from multiple sources.

---

# 27. Step 26: The Container Process Starts

After all namespaces, mounts, networking, cgroups, and security settings are ready, `runc` starts the configured process.

For the Nginx image, the effective command is similar to:

```bash
/docker-entrypoint.sh nginx -g "daemon off;"
```

The main process becomes PID 1 inside the container.

This is the actual container workload.

A container is considered running while its main process remains running.

If PID 1 exits, the container stops.

---

# 28. Step 27: `runc` Exits but the Container Continues

After creating the container and starting the process, `runc` normally exits.

The running hierarchy becomes conceptually:

```text
dockerd
  |
  v
containerd
  |
  v
containerd-shim-runc-v2
  |
  v
nginx process
```

The shim remains.

This allows the process to continue independently of the short-lived `runc` command.

---

# 29. Step 28: Docker Updates Container State

Docker records the container as running.

The container state may include:

```text
Status: running
PID: host PID
StartedAt: timestamp
IP address: 172.20.0.2
Network: app-net
Volume: app-data
Published port: 8080 -> 80
```

You can inspect it with:

```bash
docker inspect app-container
```

You can verify the running process with:

```bash
docker ps
```

---

# 30. Step 29: A Request Reaches the Container

Suppose a browser sends:

```text
http://localhost:8080
```

The request flow is:

```text
Browser
  |
  v
Host port 8080
  |
  v
Host NAT/forwarding rules
  |
  v
Docker bridge
  |
  v
Host-side veth
  |
  v
Container-side eth0
  |
  v
Container IP 172.20.0.2:80
  |
  v
Nginx
```

Nginx responds.

The response returns through the reverse path.

---

# 31. Step 30: Writing Data to the Volume

Inside the container:

```bash
docker exec app-container \
  sh -c 'echo "persistent data" > /data/message.txt'
```

The write does not go to the container writable layer.

It goes to:

```text
/var/lib/docker/volumes/app-data/_data/message.txt
```

The data remains after deleting the container.

Example:

```bash
docker rm -f app-container
```

The volume still exists:

```bash
docker volume ls
```

A new container can mount the same volume and read the same data.

---

# 32. End-to-End Low-Level Architecture

```text
+------------------------------------------------------------+
| User                                                       |
| docker run -d --name app-container                         |
|   --network app-net                                        |
|   --mount source=app-data,target=/data                     |
|   -p 8080:80 nginx:latest                                  |
+------------------------------+-----------------------------+
                               |
                               v
+------------------------------------------------------------+
| Docker CLI                                                 |
| - parses command                                           |
| - prepares Engine API request                              |
+------------------------------+-----------------------------+
                               |
                               | Unix socket or TCP
                               v
+------------------------------------------------------------+
| dockerd                                                    |
|                                                            |
| +------------------+  +------------------+                 |
| | Image Manager    |  | Volume Manager   |                 |
| | pull/check image |  | resolve app-data |                 |
| +------------------+  +------------------+                 |
|                                                            |
| +------------------+  +------------------+                 |
| | Network Manager  |  | Port Publisher   |                 |
| | attach app-net   |  | 8080 -> 80       |                 |
| +------------------+  +------------------+                 |
|                                                            |
| +--------------------------------------------------------+ |
| | Container configuration and Engine metadata           | |
| +--------------------------------------------------------+ |
+------------------------------+-----------------------------+
                               |
                               v
+------------------------------------------------------------+
| containerd                                                 |
| - image content                                            |
| - snapshot                                                 |
| - task lifecycle                                           |
| - runtime selection                                        |
+------------------------------+-----------------------------+
                               |
                               v
+------------------------------------------------------------+
| containerd-shim-runc-v2                                    |
| - process supervision                                      |
| - exit status                                              |
| - stdio                                                    |
+------------------------------+-----------------------------+
                               |
                               v
+------------------------------------------------------------+
| runc                                                       |
| - reads OCI bundle                                         |
| - creates namespaces                                       |
| - configures cgroups                                       |
| - mounts rootfs and volume                                 |
| - applies capabilities/seccomp                             |
| - starts PID 1                                             |
+------------------------------+-----------------------------+
                               |
                               v
+------------------------------------------------------------+
| Linux Kernel                                               |
|                                                            |
| Namespaces:                                                |
| - PID                                                      |
| - Mount                                                    |
| - Network                                                  |
| - IPC                                                      |
| - UTS                                                      |
| - User                                                     |
|                                                            |
| cgroups:                                                   |
| - CPU                                                      |
| - Memory                                                   |
| - PIDs                                                     |
| - I/O                                                      |
|                                                            |
| Storage:                                                   |
| - OverlayFS                                                |
| - volume bind mount                                        |
|                                                            |
| Network:                                                   |
| - bridge                                                   |
| - veth pair                                                |
| - IP routing                                               |
| - NAT/firewall rules                                       |
|                                                            |
| Security:                                                  |
| - capabilities                                             |
| - seccomp                                                  |
| - AppArmor/SELinux                                         |
+------------------------------+-----------------------------+
                               |
                               v
+------------------------------------------------------------+
| Running Container                                          |
|                                                            |
| PID 1: nginx                                               |
| IP: 172.20.0.2                                             |
| Port: 80                                                   |
| Mounted volume: /data                                      |
| Root filesystem: image layers + writable layer             |
+------------------------------------------------------------+
```

---

# 33. Exact Responsibility of Every Major Component

| Component | Responsibility |
|---|---|
| Docker CLI | Reads user commands and calls the Docker Engine API |
| Docker API | Defines requests and responses between clients and `dockerd` |
| `dockerd` | Coordinates images, containers, networks, volumes, plugins, and API operations |
| Image manager | Resolves image names, manifests, configurations, and layers |
| Volume manager | Creates and mounts persistent storage |
| Network manager | Creates networks, allocates IPs, creates endpoints, configures connectivity |
| Port publisher | Maps host ports to container ports |
| `containerd` | Manages image content, snapshots, runtime tasks, and container lifecycle |
| containerd shim | Keeps the container process supervised and reports its status |
| `runc` | Creates the actual OCI container using Linux kernel features |
| Linux namespaces | Isolate process, network, mounts, hostname, users, and IPC |
| Linux cgroups | Limit and account for CPU, memory, PIDs, and I/O |
| OverlayFS | Combines image layers with a writable container layer |
| Linux bridge | Connects container virtual interfaces on the host |
| `veth` pair | Provides a virtual Ethernet link between host and container namespace |
| iptables/nftables | Implements forwarding, NAT, filtering, and port publishing |
| Seccomp | Restricts system calls |
| Capabilities | Restrict privileged kernel operations |
| AppArmor/SELinux | Enforce mandatory access-control policies |
| Container PID 1 | The main process whose lifetime determines container lifetime |

---

# 34. Useful Commands to Observe the Flow

## Inspect the network

```bash
docker network inspect app-net
```

## Inspect the volume

```bash
docker volume inspect app-data
```

## Inspect the container

```bash
docker inspect app-container
```

## See the container IP

```bash
docker inspect \
  -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' \
  app-container
```

## See the host PID

```bash
docker inspect \
  -f '{{.State.Pid}}' \
  app-container
```

## Enter the container

```bash
docker exec -it app-container sh
```

## Check mounts inside the container

```bash
docker exec app-container mount
```

## Check network interfaces

```bash
docker exec app-container ip addr
```

The standard Nginx image may not contain the `ip` command.

## Check routes

```bash
docker exec app-container ip route
```

## Check the Docker bridge on the host

```bash
ip link
```

## Check namespaces associated with the process

```bash
PID=$(docker inspect -f '{{.State.Pid}}' app-container)

sudo ls -l /proc/$PID/ns
```

## Enter the container network namespace from the host

```bash
PID=$(docker inspect -f '{{.State.Pid}}' app-container)

sudo nsenter -t "$PID" -n ip addr
```

## Enter all major namespaces

```bash
PID=$(docker inspect -f '{{.State.Pid}}' app-container)

sudo nsenter -t "$PID" \
  -m -u -i -n -p \
  sh
```

---

# 35. Key Takeaway

When you execute:

```bash
docker run \
  --network app-net \
  --mount source=app-data,target=/data \
  -p 8080:80 \
  nginx
```

Docker does not create an independent operating system.

It coordinates several layers:

```text
Docker CLI
   |
   v
dockerd
   |
   v
containerd
   |
   v
containerd shim
   |
   v
runc
   |
   v
Linux kernel
```

The Linux kernel creates the actual isolation and resource controls.

Docker primarily provides:

- A user-friendly CLI
- An API
- Image management
- Network orchestration
- Volume orchestration
- Runtime coordination
- Port publishing
- Logging and lifecycle management

The final running container is an isolated Linux process with:

- Its own namespaces
- Controlled resources through cgroups
- A layered root filesystem
- A mounted persistent volume
- A virtual network interface
- An assigned IP address
- Host-to-container port forwarding
- Applied security restrictions
