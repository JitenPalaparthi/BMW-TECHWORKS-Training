# Spring Boot CRUD with PostgreSQL and Adminer

A layered Product CRUD REST API built for Java 21 with:

- Spring Boot 4.1.0
- Spring Web
- Spring Data JPA / Hibernate
- Bean Validation
- PostgreSQL
- Adminer database UI
- Docker and Docker Compose
- Maven

## Services

| Service | Address | Purpose |
|---|---|---|
| Spring Boot API | http://localhost:8080 | CRUD REST API |
| Adminer | http://localhost:8081 | PostgreSQL browser UI |
| PostgreSQL | localhost:5432 | Database server |

## Recommended: Run everything with Docker Compose

Requirements:

- Docker Desktop, or Docker Engine with Docker Compose

From the project directory:

```bash
docker compose up --build
```

The first build downloads the Java, Maven, PostgreSQL, and Adminer images.

To run in the background:

```bash
docker compose up --build -d
```

View logs:

```bash
docker compose logs -f app
```

Stop the containers:

```bash
docker compose down
```

Stop containers and delete PostgreSQL data:

```bash
docker compose down -v
```

## Adminer login

Open:

```text
http://localhost:8081
```

Use these values:

| Field | Value |
|---|---|
| System | PostgreSQL |
| Server | postgres |
| Username | cruduser |
| Password | crudpassword |
| Database | cruddb |

Inside Docker Compose, the PostgreSQL host is `postgres`, not `localhost`.

## CRUD endpoints

| Operation | Method | URL |
|---|---|---|
| Create product | POST | `/api/products` |
| List products | GET | `/api/products` |
| Get product | GET | `/api/products/{id}` |
| Update product | PUT | `/api/products/{id}` |
| Delete product | DELETE | `/api/products/{id}` |

### Create

```bash
curl -i -X POST http://localhost:8080/api/products \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Mechanical Keyboard",
    "description": "Hot-swappable keyboard",
    "price": 4999.00,
    "quantity": 10
  }'
```

### Get all

```bash
curl -i http://localhost:8080/api/products
```

### Get by ID

```bash
curl -i http://localhost:8080/api/products/1
```

### Update

```bash
curl -i -X PUT http://localhost:8080/api/products/1 \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Mechanical Keyboard Pro",
    "description": "Wireless hot-swappable keyboard",
    "price": 5999.00,
    "quantity": 8
  }'
```

### Delete

```bash
curl -i -X DELETE http://localhost:8080/api/products/1
```

## Run PostgreSQL in Docker and Spring Boot locally

Start only PostgreSQL and Adminer:

```bash
docker compose up -d postgres adminer
```

Verify Java:

```bash
java -version
```

Run the application:

```bash
mvn clean spring-boot:run
```

The local defaults already match the Compose database credentials:

```text
URL:      jdbc:postgresql://localhost:5432/cruddb
Username: cruduser
Password: crudpassword
```

## Run with a locally installed PostgreSQL server

Create the user and database:

```sql
CREATE USER cruduser WITH PASSWORD 'crudpassword';
CREATE DATABASE cruddb OWNER cruduser;
```

Then run:

```bash
mvn clean spring-boot:run
```

You can override connection settings without editing the source:

```bash
export SPRING_DATASOURCE_URL='jdbc:postgresql://localhost:5432/cruddb'
export SPRING_DATASOURCE_USERNAME='cruduser'
export SPRING_DATASOURCE_PASSWORD='crudpassword'
mvn spring-boot:run
```

## Change Docker database credentials

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env`, then recreate the stack. When changing credentials for an already initialized PostgreSQL volume, remove the old volume first:

```bash
docker compose down -v
docker compose up --build
```

## Build and run the JAR

```bash
mvn clean package
java -jar target/spring-boot-crud-postgres-adminer-0.0.1-SNAPSHOT.jar
```

PostgreSQL must already be running.

## Database persistence

Compose stores PostgreSQL files in the named volume:

```text
postgres_data
```

`docker compose down` preserves it. `docker compose down -v` deletes it.

## Project structure

```text
src/main/java/com/example/crud/
├── controller/
├── dto/
├── entity/
├── exception/
├── repository/
├── service/
└── CrudApplication.java
```

## Production notes

For a production system:

- Do not keep default passwords.
- Store credentials in a secret manager.
- Replace `ddl-auto=update` with Flyway or Liquibase migrations.
- Do not publish PostgreSQL or Adminer ports publicly.
- Put the API behind HTTPS and authentication.
- Pin and regularly update container images.


docker build . -t jpalaparthi/bmw-k8s-java-demo:v0.0.1