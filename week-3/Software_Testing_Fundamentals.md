# Software Testing Fundamentals

## Table of Contents

1.  Introduction
2.  Why Testing Matters
3.  Testing Pyramid
4.  Unit Testing
5.  Integration Testing
6.  System Testing
7.  Regression Testing
8.  Test Case Design
9.  JUnit Introduction
10. Assertions
11. Test Lifecycle
12. Mockito Basics
13. Mocking Dependencies
14. Test Coverage
15. TDD Introduction
16. CI/CD Testing Integration
17. Best Practices
18. Summary

# 1. Introduction

Software testing is the process of verifying and validating that
software behaves as expected, satisfies requirements, and is free from
critical defects.

**Objectives** - Find defects early - Improve software quality - Reduce
maintenance cost - Increase confidence before release

------------------------------------------------------------------------

# 2. Why Testing Matters

-   Prevent production failures
-   Improve maintainability
-   Enable safe refactoring
-   Increase developer confidence
-   Reduce cost of fixing bugs

------------------------------------------------------------------------

# 3. Testing Pyramid

               System / UI Tests
            ------------------------
            Integration Tests
        ------------------------------
             Unit Tests
    ------------------------------------

## Unit Tests

-   Fast
-   Isolated
-   Thousands can be executed

## Integration Tests

-   Verify interaction between components
-   Database
-   REST APIs
-   Messaging

## System Tests

-   Entire application
-   End-to-end business workflows
-   Highest confidence but slowest

------------------------------------------------------------------------

# 4. Unit Testing

A unit test validates one small unit (usually a method/class).

Characteristics: - Fast - Deterministic - Independent - Repeatable

Example:

``` java
assertEquals(5, calculator.add(2,3));
```

Benefits: - Early bug detection - Easier refactoring - Better
documentation

------------------------------------------------------------------------

# 5. Integration Testing

Purpose: Verify multiple modules work together correctly.

Examples: - Spring Boot + PostgreSQL - REST API + Service - Kafka
Producer + Consumer

Checks: - Data flow - Transactions - API communication

------------------------------------------------------------------------

# 6. System Testing

Tests the complete application from the user's perspective.

Examples: - Login - Checkout - Order Placement - Payment

Usually executed in an environment close to production.

------------------------------------------------------------------------

# 7. Regression Testing

Regression testing ensures existing functionality still works after
changes.

When? - Bug fixes - New features - Refactoring - Dependency upgrades

Automation is strongly recommended.

------------------------------------------------------------------------

# 8. Test Case Design

A good test case contains: - Test ID - Title - Preconditions - Test
Steps - Test Data - Expected Result - Actual Result - Status

Example:

  Field      Value
  ---------- ---------------------------
  ID         TC_LOGIN_001
  Scenario   Valid Login
  Input      Correct username/password
  Expected   Dashboard displayed

Design techniques: - Equivalence Partitioning - Boundary Value
Analysis - Decision Tables - State Transition Testing - Error Guessing -
Pairwise Testing

------------------------------------------------------------------------

# 9. JUnit Introduction

JUnit is the most popular Java unit testing framework.

Common annotations: - @Test - @BeforeEach - @AfterEach - @BeforeAll -
@AfterAll - @Disabled

Advantages: - Easy execution - Assertions - IDE integration - Build tool
integration

------------------------------------------------------------------------

# 10. Assertions

Common assertions:

``` java
assertEquals()
assertNotEquals()
assertTrue()
assertFalse()
assertNull()
assertNotNull()
assertThrows()
assertSame()
assertArrayEquals()
```

Purpose: Verify expected behavior automatically.

------------------------------------------------------------------------

# 11. Test Lifecycle

    @BeforeAll

    @BeforeEach

    @Test

    @AfterEach

    @AfterAll

Meaning: - Setup once - Setup before each test - Execute test -
Cleanup - Final cleanup

------------------------------------------------------------------------

# 12. Mockito Basics

Mockito creates mock objects.

Benefits: - Isolate dependencies - Faster testing - No real database -
No real network

Common annotations: - @Mock - @InjectMocks - @Spy

------------------------------------------------------------------------

# 13. Mocking Dependencies

Example dependencies: - Database - Email Service - Payment Gateway -
REST Client

Typical methods:

``` java
when(service.findById(1)).thenReturn(user);

verify(service).findById(1);
```

------------------------------------------------------------------------

# 14. Test Coverage

Measures how much code is tested.

Types: - Line Coverage - Branch Coverage - Method Coverage - Condition
Coverage

Remember: 100% coverage ≠ Bug-free software.

------------------------------------------------------------------------

# 15. Test Driven Development (TDD)

Cycle:

    Red
     ↓
    Green
     ↓
    Refactor

Steps: 1. Write failing test 2. Write minimal code 3. Pass test 4.
Refactor

Benefits: - Cleaner design - Better maintainability - Reduced defects

------------------------------------------------------------------------

# 16. CI/CD Testing Integration

Typical pipeline:

    Developer
       ↓
    Git Push
       ↓
    CI Server
       ↓
    Build
       ↓
    Unit Tests
       ↓
    Integration Tests
       ↓
    Static Analysis
       ↓
    Package
       ↓
    Deploy
       ↓
    System Tests
       ↓
    Production

Common tools: - GitHub Actions - Jenkins - GitLab CI - Azure DevOps

Benefits: - Early feedback - Automated quality checks - Faster releases

------------------------------------------------------------------------

# 17. Best Practices

-   Write independent tests
-   Keep tests readable
-   Test one behavior per test
-   Use meaningful names
-   Mock external dependencies
-   Avoid sleeping in tests
-   Keep unit tests fast
-   Automate regression tests
-   Run tests in CI/CD

------------------------------------------------------------------------

# 18. Summary

Testing is a layered quality strategy.

-   Unit Testing → Individual components
-   Integration Testing → Component interaction
-   System Testing → Entire application
-   Regression Testing → Protect existing functionality

A strong testing strategy combines all of these with automation, code
coverage analysis, mocking, and continuous integration.
