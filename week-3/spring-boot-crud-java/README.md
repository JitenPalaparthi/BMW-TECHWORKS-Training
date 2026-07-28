# Spring Boot CRUD Demo — Java 21

A production-structured CRUD REST API for products using Spring Boot 4.1.0, Java 21, Spring Web, Spring Data JPA, Bean Validation, and H2.

## Requirements

- OpenJDK 21 (tested target: OpenJDK 21.0.11)
- Apache Maven 3.6.3 or newer

Verify:

```bash
java -version
mvn -version
```

## Run

```bash
unzip spring-boot-crud-java21.zip
cd spring-boot-crud-java21
mvn clean spring-boot:run
```

Application: `http://localhost:8080`

H2 console: `http://localhost:8080/h2-console`

H2 connection values:

- JDBC URL: `jdbc:h2:file:./data/cruddb`
- User: `sa`
- Password: leave empty

## API endpoints

| Method | URL | Operation |
|---|---|---|
| POST | `/api/products` | Create |
| GET | `/api/products` | Read all |
| GET | `/api/products/{id}` | Read one |
| PUT | `/api/products/{id}` | Update |
| DELETE | `/api/products/{id}` | Delete |

## Test with curl

Create:

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

Get all:

```bash
curl http://localhost:8080/api/products
```

Get one:

```bash
curl http://localhost:8080/api/products/1
```

Update:

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

Delete:

```bash
curl -i -X DELETE http://localhost:8080/api/products/1
```

## Build executable JAR

```bash
mvn clean package
java -jar target/spring-boot-crud-java21-0.0.1-SNAPSHOT.jar
```

## Run tests

```bash
mvn test
```

## Project structure

```text
src/main/java/com/example/crud
├── controller/ProductController.java
├── dto/ProductRequest.java
├── dto/ProductResponse.java
├── entity/Product.java
├── exception/ApiError.java
├── exception/GlobalExceptionHandler.java
├── exception/ResourceNotFoundException.java
├── repository/ProductRepository.java
├── service/ProductService.java
└── CrudApplication.java
```
