package com.spanlet.testing;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class UserTest {

    @Test
    void shouldCreateValidUserRecord() {
        User user = new User(1L, "Jiten", "jiten@example.com");

        assertAll(
                () -> assertEquals(1L, user.id()),
                () -> assertEquals("Jiten", user.name()),
                () -> assertEquals("jiten@example.com", user.email())
        );
    }

    @Test
    void shouldRejectBlankName() {
        assertThrows(
                IllegalArgumentException.class,
                () -> new User(1L, " ", "jiten@example.com")
        );
    }

    @Test
    void shouldRejectBlankEmail() {
        assertThrows(
                IllegalArgumentException.class,
                () -> new User(1L, "Jiten", "")
        );
    }
}
