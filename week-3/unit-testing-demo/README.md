# Unit Testing Demo

- JUnit 5
- Assertions
- Test lifecycle annotations
- Parameterized tests
- Exception testing
- Mockito mocks
- Stubbing with `when(...).thenReturn(...)`
- Interaction verification with `verify(...)`
- `ArgumentCaptor`
- JaCoCo test coverage
- Docker
- Docker Compose

## Project Structure

```text
unit-testing-demo/
├── pom.xml
├── Dockerfile
├── compose.yaml
├── src/
│   ├── main/java/com/spanlet/testing/
│   │   ├── Calculator.java
│   │   ├── NotificationService.java
│   │   ├── User.java
│   │   ├── UserRepository.java
│   │   └── UserService.java
│   └── test/java/com/spanlet/testing/
│       ├── CalculatorTest.java
│       ├── UserTest.java
│       └── UserServiceTest.java
```

## Requirements for Local Execution

- Java 21
- Maven 3.9 or later

Verify:

```bash
java -version
mvn -version
```

## Run Tests Locally

```bash
mvn clean test
```

## Run One Test Class

```bash
mvn -Dtest=CalculatorTest test
```

## Run One Test Method

```bash
mvn -Dtest=CalculatorTest#addShouldReturnSum test
```

## Generate Test Coverage

JaCoCo runs automatically during the Maven test phase.

```bash
mvn clean test
```

Open the report:

```text
target/site/jacoco/index.html
```

On macOS:

```bash
open target/site/jacoco/index.html
```

On Linux:

```bash
xdg-open target/site/jacoco/index.html
```

## Run Tests with Docker

Build the image:

```bash
docker build -t java21-unit-testing-demo .
```

Run the tests:

```bash
docker run --rm java21-unit-testing-demo
```

To keep the generated `target` directory on the host:

```bash
docker run --rm \
  -v "$(pwd)/target:/workspace/target" \
  java21-unit-testing-demo
```

## Run Tests with Docker Compose

```bash
docker compose up --build
```

Remove the container:

```bash
docker compose down
```

## Important JUnit 5 Concepts

### `@Test`

Marks a method as a test method.

```java
@Test
void addShouldReturnSum() {
    assertEquals(5, calculator.add(2, 3));
}
```

### `@BeforeEach`

Runs before every test method. It is commonly used to create fresh test objects.

### `@AfterEach`

Runs after every test method. It can be used for cleanup.

### `@BeforeAll` and `@AfterAll`

Run once before and after all tests in a class.

### Common Assertions

```java
assertEquals(expected, actual);
assertTrue(condition);
assertFalse(condition);
assertNull(value);
assertNotNull(value);
assertThrows(ExceptionType.class, executable);
assertAll(...);
```

## Mockito Concepts

### Create Mock Dependencies

```java
@Mock
private UserRepository userRepository;
```

### Stub Behaviour

```java
when(userRepository.findById(1L))
        .thenReturn(Optional.of(user));
```

### Verify Interaction

```java
verify(userRepository).findById(1L);
```

### Verify No Interaction

```java
verifyNoInteractions(notificationService);
```

### Capture an Argument

```java
ArgumentCaptor<User> captor = ArgumentCaptor.forClass(User.class);
verify(notificationService).sendWelcomeEmail(captor.capture());
assertEquals(501L, captor.getValue().id());
```

## Unit-Testing Principles Demonstrated

1. Test one behaviour at a time.
2. Use meaningful test names.
3. Keep every test independent.
4. Mock external dependencies.
5. Test successful and failure paths.
6. Verify outputs and dependency interactions.
7. Avoid real databases and network calls in unit tests.
8. Run tests automatically in CI/CD.
