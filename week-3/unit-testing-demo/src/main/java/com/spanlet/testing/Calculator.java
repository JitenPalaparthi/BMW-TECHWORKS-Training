package com.spanlet.testing;

public class Calculator {

    public int add(int first, int second) {
        return first + second;
    }

    public int subtract(int first, int second) {
        return first - second;
    }

    public int multiply(int first, int second) {
        return first * second;
    }

    public double divide(int dividend, int divisor) {
        if (divisor == 0) {
            throw new IllegalArgumentException("Divisor must not be zero");
        }
        return (double) dividend / divisor;
    }

    public boolean isEven(int number) {
        return number % 2 == 0;
    }
}
