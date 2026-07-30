# CI/CD Complete Guide

## Table of Contents

1.  Introduction
2.  What is CI?
3.  What is CD?
4.  CI vs CD
5.  Why CI/CD?
6.  Typical Workflow
7.  Benefits
8.  Drawbacks
9.  Use Cases
10. Common CI/CD Tools
11. Pipeline Stages
12. Best Practices
13. Sample GitHub Actions Workflow
14. Sample Jenkins Pipeline
15. CI/CD in Kubernetes
16. DevSecOps
17. Comparison Table
18. Interview Questions

# 1. Introduction

CI/CD stands for **Continuous Integration** and **Continuous
Delivery/Continuous Deployment**.

It automates building, testing, packaging, and releasing software.

------------------------------------------------------------------------

# 2. Continuous Integration (CI)

Developers frequently merge code into a shared repository.

Every commit automatically triggers:

-   Build
-   Unit tests
-   Static code analysis
-   Security scans
-   Packaging

Goal: - Detect bugs early - Prevent broken builds

------------------------------------------------------------------------

# 3. Continuous Delivery (CD)

Software that passes CI is automatically prepared for release.

Deployment to production requires manual approval.

Typical flow:

Developer → Git Push → Build → Test → Package → Staging → Manual
Approval → Production

------------------------------------------------------------------------

# 4. Continuous Deployment

Everything after successful tests is deployed automatically to
production.

No manual approval.

------------------------------------------------------------------------

# 5. CI vs CD

  CI                  Continuous Delivery   Continuous Deployment
  ------------------- --------------------- ------------------------
  Build/Test          Ready to Release      Automatically Released
  Developer focused   Release focused       Full automation

------------------------------------------------------------------------

# 6. Benefits

-   Faster feedback
-   Early bug detection
-   Higher software quality
-   Reduced integration issues
-   Faster releases
-   Repeatable deployments
-   Infrastructure consistency
-   Better collaboration
-   Reduced manual effort
-   Easier rollback

------------------------------------------------------------------------

# 7. Drawbacks

-   Initial setup complexity
-   Infrastructure cost
-   Learning curve
-   Flaky tests reduce confidence
-   Poor test coverage weakens value
-   Pipeline maintenance
-   Secret management challenges

------------------------------------------------------------------------

# 8. Typical CI/CD Workflow

1.  Developer writes code
2.  Commit to Git
3.  Push to remote repository
4.  Webhook triggers pipeline
5.  Checkout source
6.  Restore dependencies
7.  Compile
8.  Run unit tests
9.  Static analysis
10. Security scan
11. Build artifact/container
12. Publish artifact
13. Deploy to staging
14. Integration tests
15. Approval (Delivery)
16. Production deployment
17. Monitoring
18. Rollback if needed

------------------------------------------------------------------------

# 9. Use Cases

-   Microservices
-   Web applications
-   Mobile backend
-   Cloud-native applications
-   Kubernetes
-   Embedded software
-   APIs
-   SaaS products

------------------------------------------------------------------------

# 10. Popular CI/CD Tools

## Source Control

-   Git
-   GitHub
-   GitLab
-   Bitbucket
-   Azure Repos

## CI Servers

-   Jenkins
-   GitHub Actions
-   GitLab CI/CD
-   CircleCI
-   Travis CI
-   TeamCity
-   Bamboo
-   Azure Pipelines

## Artifact Repositories

-   Nexus
-   JFrog Artifactory
-   GitHub Packages

## Containers

-   Docker
-   Podman

## Orchestration

-   Kubernetes
-   OpenShift
-   Nomad

## Deployment

-   Argo CD
-   FluxCD
-   Spinnaker
-   Helm

## Configuration

-   Ansible
-   Puppet
-   Chef
-   SaltStack

## Infrastructure as Code

-   Terraform
-   OpenTofu
-   Pulumi

## Monitoring

-   Prometheus
-   Grafana
-   ELK
-   Loki
-   Jaeger

------------------------------------------------------------------------

# 11. Pipeline Stages

-   Checkout
-   Dependency Restore
-   Compile
-   Unit Testing
-   Code Coverage
-   Linting
-   Static Analysis
-   Security Scan
-   Package
-   Docker Build
-   Push Image
-   Deploy Dev
-   Integration Test
-   Deploy QA
-   Acceptance Test
-   Approval
-   Deploy Production
-   Smoke Test
-   Monitoring

------------------------------------------------------------------------

# 12. Best Practices

-   Keep builds fast
-   Automate everything possible
-   Use small commits
-   Maintain high test coverage
-   Fail fast
-   Version artifacts
-   Store secrets securely
-   Use immutable artifacts
-   Blue-Green or Canary deployments
-   Monitor deployments

------------------------------------------------------------------------

# 13. Sample GitHub Actions

``` yaml
name: CI

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-java@v4
        with:
          distribution: temurin
          java-version: 21

      - run: mvn clean verify
```

------------------------------------------------------------------------

# 14. Sample Jenkins Pipeline

``` groovy
pipeline {
  agent any

  stages {
    stage('Build') {
      steps {
        sh 'mvn clean package'
      }
    }

    stage('Test') {
      steps {
        sh 'mvn test'
      }
    }

    stage('Deploy') {
      steps {
        sh './deploy.sh'
      }
    }
  }
}
```

------------------------------------------------------------------------

# 15. CI/CD with Kubernetes

Git Push → CI Build → Docker Image → Registry → ArgoCD/Flux → Kubernetes
Cluster

------------------------------------------------------------------------

# 16. DevSecOps

Security should be integrated into every pipeline stage:

-   SAST
-   DAST
-   Dependency Scanning
-   Secret Scanning
-   Container Scanning
-   SBOM generation
-   Policy enforcement

------------------------------------------------------------------------

# 17. Tool Comparison

  Tool              Open Source   Strength
  ----------------- ------------- --------------------------
  Jenkins           Yes           Highly extensible
  GitHub Actions    No            Tight GitHub integration
  GitLab CI         Partial       Integrated platform
  CircleCI          No            Fast cloud CI
  Azure Pipelines   No            Microsoft ecosystem
  Argo CD           Yes           GitOps for Kubernetes
  FluxCD            Yes           Kubernetes-native GitOps

------------------------------------------------------------------------

# 18. Frequently Asked Interview Questions

1.  Difference between CI and CD?
2.  Delivery vs Deployment?
3.  Blue-Green deployment?
4.  Canary deployment?
5.  Rolling update?
6.  GitOps?
7.  Why artifact repository?
8.  Why immutable artifacts?
9.  What is pipeline as code?
10. What causes flaky tests?
11. Why shift-left testing?
12. Why infrastructure as code?

------------------------------------------------------------------------

# Summary

CI focuses on integrating, building, and testing code continuously.

Continuous Delivery prepares software for release with manual approval.

Continuous Deployment releases automatically after all quality gates
pass.

A mature CI/CD pipeline improves software quality, delivery speed,
reliability, and developer productivity while reducing deployment risk.
