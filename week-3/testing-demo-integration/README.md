# Java 21 Unit and PostgreSQL Integration Testing Demo

## Versions
- Java 21
- Spring Boot 3.5.4
- Testcontainers 1.21.3
- PostgreSQL 17

This project deliberately uses Spring Boot 3.x to avoid the Spring Boot 4 modular test-package and Jackson 3 migration issues.

## Run unit tests
```bash
mvn clean test
```
This runs `UserServiceTest`. Integration-test source files are compiled but not executed.

## Run all tests
Docker Desktop must be running.
```bash
docker info
mvn clean verify
```
This also runs `UserRepositoryIT` and `UserApiIT` against disposable PostgreSQL containers.

## Run application
```bash
docker compose up --build
```

## API
```bash
curl -X POST http://localhost:8080/api/users   -H 'Content-Type: application/json'   -d '{"name":"Jiten","email":"jiten@example.com"}'

curl http://localhost:8080/api/users
```

## Coverage
After `mvn clean verify`, open `target/site/jacoco/index.html`.
