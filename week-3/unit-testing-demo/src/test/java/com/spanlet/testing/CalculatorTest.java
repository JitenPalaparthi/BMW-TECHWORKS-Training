package com.spanlet.testing;

import org.junit.jupiter.api.*;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;
import org.junit.jupiter.params.provider.ValueSource;

@DisplayName("Calculator unit tests")
class CalculatorTest {

    private Calculator calculator;

    @BeforeAll
    static void beforeAllTests() {
        System.out.println("Starting Calculator tests");
    }

    @BeforeEach
    void setUp() {
        calculator = new Calculator();
    }

    @Test
    @DisplayName("add should return the sum of two integers")
    void addShouldReturnSum() {
        assertEquals(5, calculator.add(2, 3));
    }

    @Test
    void subtractShouldReturnDifference() {
        assertEquals(7, calculator.subtract(10, 3));
    }

    @Test
    void multiplyShouldReturnProduct() {
        assertEquals(20, calculator.multiply(4, 5));
    }

    @Test
    void divideShouldReturnQuotient() {
        assertEquals(2.5, calculator.divide(5, 2), 0.001);
    }

    @Test
    void divideByZeroShouldThrowException() {
        IllegalArgumentException exception = assertThrows(
                IllegalArgumentException.class,
                () -> calculator.divide(10, 0)
        );

        assertEquals("Divisor must not be zero", exception.getMessage());
    }

    @ParameterizedTest
    @ValueSource(ints = {2, 4, 6, 8, 10})
    void isEvenShouldReturnTrueForEvenNumbers(int number) {
        assertTrue(calculator.isEven(number));
    }

    @ParameterizedTest
    @CsvSource({
            "1, 2, 3",
            "10, 20, 30",
            "-5, 5, 0"
    })
    void addShouldWorkForMultipleInputs(int first, int second, int expected) {
        assertEquals(expected, calculator.add(first, second));
    }

    @AfterEach
    void tearDown() {
        calculator = null;
    }

    @AfterAll
    static void afterAllTests() {
        System.out.println("Calculator tests completed");
    }
}
