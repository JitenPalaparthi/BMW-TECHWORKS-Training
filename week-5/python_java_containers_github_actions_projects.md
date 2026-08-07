# 10 End-to-End Projects Using Python, Java, Containers, GitHub Actions, and Docker Images

## Overview

This document contains **10 practical end-to-end projects** designed to teach software development, containerization, CI/CD, image publishing, and deployment.

Each project includes:

- Project objective
- Business use case
- Suggested architecture
- Python responsibilities
- Java responsibilities
- Database or messaging components
- REST APIs
- Docker/containerization
- Docker Compose where required
- GitHub repository structure
- GitHub Actions CI/CD workflow
- Unit and integration testing
- Security scanning
- Docker image build and tagging
- Docker Hub or GHCR publishing
- Deployment strategy
- Production improvements
- Suggested milestones

The projects progress from intermediate to advanced level.

---

# Common Technology Stack

## Backend Languages

- Python 3.12+
- Java 21+
- Spring Boot 3.x
- FastAPI or Flask

## Databases

- PostgreSQL
- Redis
- MongoDB where appropriate

## Messaging

- Apache Kafka or RabbitMQ for event-driven projects

## Container Platform

- Docker
- Docker Compose
- Optional Podman

## CI/CD

- GitHub Actions

## Container Registries

- Docker Hub
- GitHub Container Registry (GHCR)
- Nexus Repository Manager as an optional private registry

## Testing

### Python

- pytest
- unittest
- coverage.py

### Java

- JUnit 5
- Mockito
- Testcontainers
- JaCoCo

## Security and Quality

- Trivy
- Dependabot
- CodeQL
- SonarQube/SonarCloud
- pip-audit
- OWASP Dependency Check

---

# Project 1 — Employee Management Platform

## Difficulty

Intermediate

## Objective

Build an employee management platform in which Java provides the primary business APIs and Python provides reporting and analytics APIs.

## Business Requirements

The system should allow administrators to:

- Create employees
- Update employee details
- Delete employees
- Search employees
- Assign departments
- View employee statistics
- Generate salary and department reports

## Architecture

```text
Client
  |
  v
Java Spring Boot API
  |
  +---- PostgreSQL
  |
  +---- Python Reporting Service
             |
             +---- PostgreSQL Read Access
```

## Java Service Responsibilities

Use Spring Boot for:

- Employee CRUD operations
- Department CRUD operations
- Validation
- Database transactions
- REST API exposure

Suggested endpoints:

```text
POST   /api/employees
GET    /api/employees
GET    /api/employees/{id}
PUT    /api/employees/{id}
DELETE /api/employees/{id}
GET    /api/departments
POST   /api/departments
```

## Python Service Responsibilities

Use FastAPI for:

- Department-wise employee counts
- Salary statistics
- Joining trend analysis
- CSV report generation

Suggested endpoints:

```text
GET /reports/headcount
GET /reports/salary-summary
GET /reports/joining-trends
GET /reports/export
```

## Database

PostgreSQL tables:

```text
employees
---------
id
name
email
salary
department_id
joining_date


departments
-----------
id
name
location
```

## Suggested Repository Structure

```text
employee-platform/
├── java-service/
│   ├── src/
│   ├── pom.xml
│   └── Dockerfile
├── python-reporting/
│   ├── app/
│   ├── tests/
│   ├── requirements.txt
│   └── Dockerfile
├── docker-compose.yml
├── .github/
│   └── workflows/
│       ├── java-ci.yml
│       ├── python-ci.yml
│       └── docker-release.yml
└── README.md
```

## Containerization

Create separate images:

```text
employee-java-api
employee-python-reporting
```

Use Docker Compose to start:

- PostgreSQL
- Java service
- Python service

## GitHub Actions Pipeline

Pipeline stages:

```text
Checkout
   |
   +--> Java Build -> JUnit -> JaCoCo
   |
   +--> Python Install -> pytest -> coverage
   |
   +--> CodeQL / dependency scan
   |
   +--> Docker Build
   |
   +--> Trivy Image Scan
   |
   +--> Push Docker Images
```

## Docker Image Tags

Use:

```text
latest
sha-<git-sha>
v1.0.0
```

## Deployment

Deploy initially using Docker Compose on a Linux VM.

Example flow:

```text
git push
   -> GitHub Actions
   -> Build
   -> Test
   -> Docker Build
   -> Push GHCR
   -> Linux VM pulls image
   -> docker compose up -d
```

## Enhancements

- JWT authentication
- Role-based access control
- Prometheus metrics
- Grafana dashboard
- Flyway database migrations

---

# Project 2 — E-Commerce Product and Recommendation Platform

## Difficulty

Intermediate to Advanced

## Objective

Build an e-commerce backend where Java handles product/order management and Python provides recommendation functionality.

## Java Services

### Product Service

Responsibilities:

- Products
- Categories
- Inventory
- Pricing

### Order Service

Responsibilities:

- Create order
- Validate inventory
- Calculate totals
- Maintain order status

## Python Recommendation Service

Use Python for:

- Similar-product recommendations
- Trending products
- Recommendation scores
- Basic machine-learning experimentation

## Architecture

```text
                       +------------------+
                       | Python Recommend |
                       +--------+---------+
                                |
Client -> Java API Gateway -----+
           |
           +---- Product Service ---- PostgreSQL
           |
           +---- Order Service ------ PostgreSQL
```

## APIs

```text
GET  /products
POST /products
POST /orders
GET  /orders/{id}
GET  /recommendations/{productId}
GET  /recommendations/trending
```

## Containers

Build images for:

```text
product-service
order-service
recommendation-service
```

## GitHub Actions

Create reusable workflows.

```text
.github/workflows/
├── java-service-template.yml
├── python-service-template.yml
└── release.yml
```

Release workflow should:

1. Trigger on Git tags.
2. Build all services.
3. Run unit tests.
4. Run integration tests.
5. Build Docker images.
6. Scan images.
7. Push images.
8. Deploy release.

## Deployment

First stage:

- Docker Compose

Advanced stage:

- Kubernetes
- Helm

## Production Improvements

- Redis caching
- Kafka-based order events
- OpenTelemetry
- Centralized logging

---

# Project 3 — Online Banking and Fraud Detection System

## Difficulty

Advanced

## Objective

Build a simplified banking platform with Java transaction services and a Python fraud-detection service.

## Java Components

### Account Service

- Create account
- Retrieve balance
- Credit account
- Debit account

### Transaction Service

- Money transfer
- Transaction history
- Idempotency
- Validation

## Python Fraud Service

Analyze transactions using rules such as:

- High-value transfers
- Too many transfers in a short interval
- New recipient
- Unusual geographic or temporal patterns

## Architecture

```text
User
 |
 v
Transaction Service
 |
 +--> Account Service
 |
 +--> Kafka
        |
        +--> Python Fraud Detector
        |
        +--> Audit Service
```

## Event Example

```json
{
  "transactionId": "tx123",
  "accountId": "acc100",
  "amount": 75000,
  "type": "TRANSFER",
  "timestamp": "2026-08-07T10:20:00Z"
}
```

## Container Stack

```text
Java Account Service
Java Transaction Service
Python Fraud Service
PostgreSQL
Kafka
Kafka UI
```

## CI/CD Requirements

Use GitHub Actions matrix strategy.

Example:

```text
Build Java services concurrently
Build Python service concurrently
Run service-specific tests
Run integration test environment
Build images
Run Trivy
Publish images
```

## Deployment

Deploy using Docker Compose initially.

Advanced deployment:

- Kubernetes
- Separate namespace
- ConfigMaps
- Secrets
- Horizontal Pod Autoscaler

## Important Learning Areas

- Distributed transactions
- Event-driven architecture
- Eventual consistency
- Idempotency
- Security
- Audit trails

---

# Project 4 — URL Shortener with Analytics

## Difficulty

Intermediate

## Objective

Build a high-performance URL shortening service.

## Java Responsibilities

Spring Boot API:

```text
POST /shorten
GET  /{shortCode}
DELETE /links/{id}
```

Java handles:

- URL creation
- Redirect resolution
- Expiration
- Link ownership

## Python Analytics Service

Python consumes access logs and produces:

- Click counts
- Browser statistics
- Country statistics
- Time-series statistics

## Architecture

```text
Browser
  |
  v
Java URL Service ---> PostgreSQL
  |
  +--> Redis
  |
  +--> Kafka ---> Python Analytics ---> Analytics DB
```

## Container Images

```text
url-service
analytics-service
```

## GitHub Actions Features

- Maven cache
- pip cache
- Unit tests
- Coverage
- Docker layer caching using BuildKit
- Multi-platform images using buildx

Example platforms:

```text
linux/amd64
linux/arm64
```

## Deployment

Deploy behind Nginx.

```text
Internet
  |
 Nginx
  |
  +--> URL Java Service
  +--> Python Analytics API
```

## Advanced Topics

- Base62 ID encoding
- Hash collision handling
- Redis caching
- Rate limiting

---

# Project 5 — Food Delivery Order Processing Platform

## Difficulty

Advanced

## Objective

Build a distributed food-delivery backend inspired by real-world delivery applications.

## Services

### Java

- Restaurant Service
- Order Service
- Delivery Service

### Python

- ETA Prediction Service
- Delivery Assignment Service

## Flow

```text
Customer creates order
        |
        v
Order Service
        |
        v
Kafka: order-created
        |
        +--> Restaurant Service
        +--> Python Assignment Service
        +--> Notification Service
```

## Python ETA Service

Inputs:

- Distance
- Restaurant preparation time
- Driver availability
- Traffic factor

Output:

```json
{
  "estimatedDeliveryMinutes": 32
}
```

## Docker Compose Stack

- PostgreSQL
- Redis
- Kafka
- Java services
- Python services
- Prometheus
- Grafana

## CI Pipeline

Each pull request should run:

```text
Formatting checks
Static analysis
Unit tests
Integration tests
Docker build check
Security scans
```

Only tagged releases should push production images.

## Deployment Strategy

Use blue-green deployment on Docker hosts or Kubernetes.

## Learning Outcomes

- Microservices
- Kafka
- Distributed workflows
- Async processing
- API composition
- Observability

---

# Project 6 — Document Processing and Search Platform

## Difficulty

Advanced

## Objective

Build a system where users upload documents and search extracted content.

## Java API

Handles:

- User authentication
- File metadata
- Upload requests
- Search API
- Document status

## Python Worker

Handles:

- PDF/text parsing
- Keyword extraction
- Metadata extraction
- Text indexing

## Architecture

```text
User
 |
 v
Java Document API
 |
 +--> MinIO
 |
 +--> PostgreSQL
 |
 +--> Kafka ---> Python Processor ---> OpenSearch
```

## APIs

```text
POST /documents
GET  /documents/{id}
GET  /documents/{id}/status
GET  /search?q=container
```

## Containers

Images:

```text
document-api
python-document-worker
```

Infrastructure containers:

```text
PostgreSQL
Kafka
MinIO
OpenSearch
```

## GitHub Actions

Use separate stages:

```text
lint
unit-test
integration-test
container-build
container-scan
container-publish
```

## Deployment

Deploy worker independently from API.

This demonstrates why microservices can scale based on workload.

Example:

```text
API replicas: 2
Worker replicas: 10
```

## Advanced Features

- OCR
- Virus scanning
- Signed download URLs
- Document retention
- Vector search

---

# Project 7 — IoT Device Monitoring Platform

## Difficulty

Advanced

## Objective

Build a backend for collecting telemetry from IoT devices.

## Java Device Management Service

Responsibilities:

- Device registration
- Device ownership
- Device configuration
- API keys
- Device status

## Python Telemetry Processor

Responsibilities:

- Process telemetry
- Detect anomalies
- Compute aggregates
- Trigger alerts

## Architecture

```text
IoT Devices
    |
    v
 MQTT Broker
    |
    v
 Python Telemetry Consumer
    |
    +--> Time-series DB
    +--> Kafka
            |
            v
      Java Device API
```

## Telemetry Example

```json
{
  "deviceId": "sensor-101",
  "temperature": 31.4,
  "humidity": 68,
  "timestamp": "2026-08-07T08:00:00Z"
}
```

## Containers

- Mosquitto
- Java API
- Python telemetry processor
- PostgreSQL
- TimescaleDB
- Grafana

## GitHub Actions

Add integration tests that start dependencies using Testcontainers or GitHub Actions service containers.

## Deployment

Use Docker Compose for lab setup.

Production:

- Kubernetes
- Persistent storage
- TLS
- Device certificate authentication

## Learning Areas

- IoT messaging
- MQTT
- Streaming
- Time-series data
- Alert generation

---

# Project 8 — CI/CD Build and Deployment Dashboard

## Difficulty

Intermediate to Advanced

## Objective

Build a DevOps dashboard that tracks application builds, deployments, and container images.

## Java Backend

Use Spring Boot for:

- Project registration
- Build metadata
- Deployment history
- User access

## Python Automation Service

Use Python for:

- GitHub API integration
- Docker registry queries
- Build statistics
- Deployment automation

## Architecture

```text
GitHub Actions
     |
     | Webhook / API
     v
Java DevOps API ---> PostgreSQL
     |
     +--> Python Automation Service
                    |
                    +--> GitHub
                    +--> Docker Registry
```

## Key APIs

```text
POST /projects
GET  /projects/{id}/builds
GET  /projects/{id}/deployments
POST /deployments
GET  /images
```

## GitHub Actions Integration

Application repositories report:

- Commit SHA
- Build number
- Branch
- Image tag
- Deployment status

## Docker Images

```text
devops-dashboard-api
devops-automation-service
```

## Deployment

Deploy the dashboard itself using its own GitHub Actions workflow.

This project therefore demonstrates self-hosting CI/CD concepts.

## Advanced Features

- GitHub App integration
- Webhooks
- Environment approvals
- Deployment audit trail
- Rollback buttons

---

# Project 9 — Real-Time Logistics and Shipment Tracking Platform

## Difficulty

Advanced

## Objective

Track shipments, update locations, and estimate arrival times.

## Java Services

### Shipment Service

- Shipment creation
- Shipment status
- Customer mapping

### Tracking API

- Retrieve current location
- Retrieve shipment history

## Python Service

ETA calculation based on:

- Previous movement speed
- Distance
- Route
- Delay history

## Architecture

```text
GPS / Mobile Client
      |
      v
Location Ingestion API
      |
      v
Kafka
  |
  +--> Java Tracking Consumer
  |
  +--> Python ETA Service
  |
  +--> Alert Service
```

## Container Environment

```text
shipment-service
tracking-service
eta-service
postgres
kafka
redis
```

## GitHub Actions

Use environment-based pipelines.

```text
main branch -> staging
v* tag      -> production
```

Stages:

```text
build
unit-test
integration-test
image-build
image-scan
push
staging-deploy
smoke-test
production-approval
production-deploy
```

## Deployment

Support rolling updates.

## Production Topics

- Geo indexes
- Event ordering
- Kafka partitions
- Exactly-once vs at-least-once processing

---

# Project 10 — Cloud-Native Online Learning Platform

## Difficulty

Advanced / Capstone

## Objective

Build a complete online-learning backend using Java and Python microservices with containerized CI/CD.

## Java Microservices

### User Service

- Registration
- Login
- Profile
- RBAC

### Course Service

- Courses
- Modules
- Lessons

### Enrollment Service

- Enrollment
- Completion
- Progress

## Python Microservices

### Recommendation Service

- Recommend courses
- Skill-based matching

### Reporting Service

- Completion reports
- Course popularity
- User activity

## Architecture

```text
                        +-------------------------+
                        |     API Gateway         |
                        +-----------+-------------+
                                    |
        +---------------------------+--------------------------+
        |                           |                          |
        v                           v                          v
 User Service                Course Service             Enrollment Service
   Java                         Java                        Java
        |                           |                          |
        +---------------------------+--------------------------+
                                    |
                                  Kafka
                                    |
                  +-----------------+------------------+
                  |                                    |
                  v                                    v
       Python Recommendation                Python Reporting
```

## Infrastructure

- PostgreSQL
- Redis
- Kafka
- MinIO
- Prometheus
- Grafana

## Authentication

Recommended options:

- Spring Security
- JWT

Advanced:

- Keycloak
- OAuth 2.0
- OpenID Connect

## Repository Strategy

### Monorepo Option

```text
learning-platform/
├── services/
│   ├── user-service/
│   ├── course-service/
│   ├── enrollment-service/
│   ├── recommendation-service/
│   └── reporting-service/
├── infra/
│   ├── docker-compose.yml
│   └── prometheus/
├── .github/workflows/
└── docs/
```

### Multi-Repo Option

Each microservice gets a separate Git repository.

Use reusable GitHub workflows from a central DevOps repository.

## Full CI/CD Workflow

```text
Developer
   |
   v
Git Push / Pull Request
   |
   v
GitHub Actions
   |
   +--> Checkout
   +--> Dependency Restore
   +--> Lint
   +--> Unit Tests
   +--> Integration Tests
   +--> Coverage
   +--> CodeQL
   +--> Sonar Analysis
   +--> Docker Build
   +--> Trivy Scan
   +--> Push to GHCR
   +--> Deploy Staging
   +--> Smoke Test
   +--> Manual Approval
   +--> Deploy Production
```

## Image Naming

Example:

```text
ghcr.io/company/user-service:v1.0.0
ghcr.io/company/course-service:v1.0.0
ghcr.io/company/enrollment-service:v1.0.0
ghcr.io/company/recommendation-service:v1.0.0
```

## Production Deployment

Recommended final stage:

- Kubernetes
- Helm
- Ingress
- ConfigMaps
- Secrets
- HPA
- Persistent Volumes

## Observability

Add:

- Spring Boot Actuator
- Prometheus
- Grafana
- OpenTelemetry
- Loki

## Capstone Learning Outcomes

This project combines:

- Java
- Python
- REST
- Microservices
- Kafka
- PostgreSQL
- Redis
- Docker
- GitHub Actions
- GHCR
- Security scanning
- Kubernetes
- Observability

---

# Standard Dockerfile — Java Spring Boot

```dockerfile
FROM eclipse-temurin:21-jdk AS builder
WORKDIR /app

COPY pom.xml .
COPY src ./src

RUN ./mvnw clean package -DskipTests

FROM eclipse-temurin:21-jre
WORKDIR /app

COPY --from=builder /app/target/*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
```

Prefer using the Maven wrapper (`mvnw`) checked into the repository.

---

# Standard Dockerfile — Python FastAPI

```dockerfile
FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app ./app

EXPOSE 8000

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

---

# Example Docker Compose

```yaml
services:
  postgres:
    image: postgres:17
    environment:
      POSTGRES_USER: appuser
      POSTGRES_PASSWORD: apppassword
      POSTGRES_DB: appdb
    volumes:
      - postgres_data:/var/lib/postgresql/data

  java-api:
    build: ./java-service
    environment:
      SPRING_DATASOURCE_URL: jdbc:postgresql://postgres:5432/appdb
      SPRING_DATASOURCE_USERNAME: appuser
      SPRING_DATASOURCE_PASSWORD: apppassword
    depends_on:
      - postgres
    ports:
      - "8080:8080"

  python-api:
    build: ./python-service
    environment:
      DATABASE_URL: postgresql://appuser:apppassword@postgres:5432/appdb
    depends_on:
      - postgres
    ports:
      - "8000:8000"

volumes:
  postgres_data:
```

---

# Example GitHub Actions — Java CI

```yaml
name: Java CI

on:
  pull_request:
  push:
    branches:
      - main

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Java
        uses: actions/setup-java@v4
        with:
          distribution: temurin
          java-version: "21"
          cache: maven

      - name: Run tests
        run: ./mvnw --batch-mode clean verify
```

---

# Example GitHub Actions — Python CI

```yaml
name: Python CI

on:
  pull_request:
  push:
    branches:
      - main

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.12"
          cache: pip

      - name: Install dependencies
        run: pip install -r requirements.txt

      - name: Run tests
        run: pytest --cov=app --cov-report=term-missing
```

---

# Example GitHub Actions — Build and Push Docker Image to GHCR

```yaml
name: Build and Publish Container

on:
  push:
    tags:
      - "v*"

permissions:
  contents: read
  packages: write

jobs:
  publish:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: |
            ghcr.io/${{ github.repository }}:${{ github.ref_name }}
            ghcr.io/${{ github.repository }}:latest
```

---

# Recommended Complete CI/CD Pipeline

```text
Developer commits code
        |
        v
Pull Request
        |
        +--> Lint
        +--> Unit Tests
        +--> Integration Tests
        +--> Code Quality
        +--> Security Scan
        |
        v
Merge to main
        |
        +--> Build application
        +--> Build container image
        +--> Trivy scan
        +--> Push SHA-tagged image
        +--> Deploy staging
        +--> Smoke tests
        |
        v
Create Git tag v1.0.0
        |
        +--> Build release image
        +--> Push version tag
        +--> Push latest tag
        +--> Production deployment
```

---

# Recommended Git Strategy

Use:

```text
feature/*
   |
   v
Pull Request
   |
   v
main
   |
   v
release tag
```

Avoid long-lived branches unless the organization specifically requires them.

---

# Docker Image Tagging Strategy

For development builds:

```text
service:sha-a1b2c3d
```

For releases:

```text
service:1.3.0
service:1.3
service:1
service:latest
```

For environments:

Prefer immutable SHA/version tags rather than relying on:

```text
staging
production
```

---

# Suggested GitHub Actions Security Controls

Use minimal permissions:

```yaml
permissions:
  contents: read
  packages: write
```

Avoid storing passwords directly inside YAML.

Store secrets under:

```text
Repository Settings
-> Secrets and variables
-> Actions
```

Examples:

```text
DOCKERHUB_USERNAME
DOCKERHUB_TOKEN
SONAR_TOKEN
DEPLOY_SSH_KEY
```

---

# Image Security Scanning with Trivy

Example stage:

```yaml
- name: Run Trivy
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ghcr.io/example/app:${{ github.sha }}
    format: table
    exit-code: "1"
    severity: CRITICAL,HIGH
```

In real production repositories, pin actions to trusted immutable revisions according to organizational supply-chain policy.

---

# Suggested Training Order

| Project | Primary Learning Goal | Difficulty |
|---|---|---|
| 1 | REST, Docker, basic CI/CD | Intermediate |
| 2 | Multi-service architecture | Intermediate-Advanced |
| 3 | Kafka and fraud events | Advanced |
| 4 | Caching and analytics | Intermediate |
| 5 | Event-driven microservices | Advanced |
| 6 | Async document processing | Advanced |
| 7 | IoT and streaming | Advanced |
| 8 | DevOps automation | Intermediate-Advanced |
| 9 | Real-time logistics | Advanced |
| 10 | Full cloud-native capstone | Advanced |

---

# Recommended Project Execution Template

Use the following implementation sequence for every project.

## Phase 1 — Requirements

Define:

- Functional requirements
- Non-functional requirements
- APIs
- Database schema
- Service boundaries

## Phase 2 — Local Development

Create:

- Java application
- Python application
- Unit tests
- Local database

## Phase 3 — Containerization

Create:

- Dockerfiles
- `.dockerignore`
- Docker Compose
- Health checks

## Phase 4 — CI

GitHub Actions should execute:

```text
checkout
build
lint
test
coverage
security checks
```

## Phase 5 — Container Release

Build images using Docker Buildx.

Run:

```text
security scan
image tagging
registry login
image push
```

## Phase 6 — Deployment

Start with:

```text
Docker Compose on Linux VM
```

Then progress to:

```text
Kubernetes
```

## Phase 7 — Production Readiness

Add:

- Monitoring
- Centralized logging
- Distributed tracing
- Secrets management
- Database migrations
- Backups
- Rollbacks
- Horizontal scaling
- Health checks
- Graceful shutdown

---

# Final Learning Outcome

After implementing these ten projects, a learner should be able to explain and demonstrate the complete software delivery lifecycle:

```text
Requirement
   -> Design
   -> Code
   -> Unit Test
   -> Integration Test
   -> Package
   -> Containerize
   -> Scan
   -> Publish Image
   -> Deploy
   -> Monitor
   -> Upgrade
   -> Rollback
```

The most important objective is not merely learning Python, Java, or Docker individually. The projects should teach how these technologies work together in a realistic engineering workflow from source code to a running production-style containerized application.
