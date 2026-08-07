# Docker Architecture -- Complete Guide

------------------------------------------------------------------------

# 1. What is Docker?

Docker is **not a single executable**. It is an ecosystem of cooperating
components that build, distribute, and run OCI-compatible containers.

At a high level:

``` text
                        User
                          │
                          ▼
                    Docker CLI
                          │
                    REST API (Unix socket/TCP)
                          │
                          ▼
                  Docker Engine (dockerd)
        ┌──────────────┬───────────────┬──────────────┐
        │              │               │              │
        ▼              ▼               ▼              ▼
    BuildKit      Network Manager  Volume Manager  Image Manager
        │
        ▼
    containerd
        │
        ▼
       runc
        │
        ▼
 Linux Kernel (Namespaces, cgroups, OverlayFS,
               Capabilities, Seccomp,
               AppArmor/SELinux)
```

------------------------------------------------------------------------

# 2. Components Written or Maintained by Docker

  -----------------------------------------------------------------------
  Component                               Purpose
  --------------------------------------- -------------------------------
  Docker CLI                              User commands (`docker run`,
                                          `build`, `pull`)

  Docker Engine (dockerd / Moby)          Main daemon, API server,
                                          orchestration

  BuildKit                                Modern image builder

  Buildx                                  Multi-platform builds

  Docker Compose v2                       Multi-container orchestration

  Docker Desktop                          Desktop application for
                                          Windows/macOS
  -----------------------------------------------------------------------

------------------------------------------------------------------------

# 3. Components that Already Existed

  Component            Owner
  -------------------- ---------------------------
  Linux Kernel         Linux Community
  Namespaces           Linux Kernel
  cgroups              Linux Kernel
  OverlayFS            Linux Kernel
  Seccomp              Linux Kernel
  AppArmor             Linux Kernel
  SELinux              Linux Community / NSA
  containerd           CNCF
  runc                 Open Container Initiative
  OCI Specifications   Open Container Initiative

Docker integrates these technologies rather than reimplementing them.

------------------------------------------------------------------------

# 4. Docker CLI

Responsibilities

-   Parses user commands
-   Validates arguments
-   Sends REST API requests to Docker Engine

The CLI never creates containers directly.

------------------------------------------------------------------------

# 5. Docker Engine (dockerd)

The central daemon.

Responsibilities

-   REST API server
-   Image management
-   Pull/Push
-   Container lifecycle
-   Networking
-   Volumes
-   Logging
-   Delegates execution to containerd

------------------------------------------------------------------------

# 6. BuildKit

Modern image builder.

Features

-   Parallel builds
-   Layer caching
-   Secret mounts
-   SSH forwarding
-   Multi-stage builds
-   Efficient dependency graph

------------------------------------------------------------------------

# 7. Image Manager

Maintains

-   OCI images
-   Layers
-   Manifests
-   Digests
-   Local image cache

------------------------------------------------------------------------

# 8. Network Manager

Creates

-   bridge networks
-   host mode
-   overlay
-   macvlan
-   none

Also configures DNS, NAT and port mapping.

------------------------------------------------------------------------

# 9. Volume Manager

Provides persistent storage.

Supports

-   local volumes
-   plugins
-   bind mounts

------------------------------------------------------------------------

# 10. Logging Drivers

Examples

-   json-file
-   journald
-   syslog
-   fluentd
-   splunk
-   awslogs

------------------------------------------------------------------------

# 11. containerd

Low-level container manager.

Responsibilities

-   Image unpacking
-   Snapshot management
-   Runtime management
-   Container lifecycle
-   Task management

Kubernetes commonly talks directly to containerd.

------------------------------------------------------------------------

# 12. runc

OCI runtime.

Responsibilities

-   Create namespaces
-   Configure cgroups
-   Mount root filesystem
-   Apply capabilities
-   Execute PID 1

This is the component that actually starts the container process.

------------------------------------------------------------------------

# 13. Linux Kernel Components

## Namespaces

Isolation of

-   PID
-   Network
-   Mount
-   IPC
-   UTS
-   User

## cgroups

Resource limits

-   CPU
-   Memory
-   Disk I/O
-   PIDs

## OverlayFS

Provides immutable image layers plus a writable container layer.

## Capabilities

Splits root privileges into fine-grained permissions.

## Seccomp

Filters allowed system calls.

## AppArmor / SELinux

Mandatory access control for additional isolation.

------------------------------------------------------------------------

# 14. Image Layers

Each image consists of immutable layers.

Container startup adds a writable layer.

``` text
Writable Container Layer
------------------------
Application Layer
Libraries
Base Image
```

------------------------------------------------------------------------

# 15. Daemon-Based vs Daemonless Containers

## Daemon-Based (Docker)

``` text
CLI
 │
 ▼
dockerd
 │
 ▼
containerd
 │
 ▼
runc
```

Advantages

-   Rich API
-   Easy UX
-   Plugins
-   Networking
-   Logging
-   Volume management

Disadvantages

-   Long-running daemon
-   Extra memory
-   Single management process

## Daemonless (Podman)

``` text
podman
   │
   ▼
conmon
   │
   ▼
runc / crun
```

Characteristics

-   No central daemon
-   Containers owned by user processes
-   Better rootless experience
-   Smaller attack surface

------------------------------------------------------------------------

# 16. Docker Run Flow

``` text
docker run nginx
      │
      ▼
Docker CLI
      │
REST API
      ▼
dockerd
      │
Image lookup
      │
Pull if missing
      │
Create writable layer
      │
Configure network
      │
Create mounts
      ▼
containerd
      ▼
runc
      ▼
Linux Kernel
      │
Namespaces
cgroups
Capabilities
Seccomp
OverlayFS
      ▼
Container Process
```

------------------------------------------------------------------------

# 17. Summary

Docker is an orchestration platform around Linux container primitives.
Docker contributes the developer tooling (CLI, Engine, BuildKit,
Compose, Desktop), while relying on standardized components such as
containerd, runc, OCI specifications, and Linux kernel features to build
and run containers efficiently and securely.
