package com.spanlet.testing;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import java.util.NoSuchElementException;
import java.util.Optional;
import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UserServiceTest {
    @Mock UserRepository repository;
    @InjectMocks UserService service;

    @Test
    void createNormalizesAndSavesUser() {
        when(repository.existsByEmail("jiten@example.com")).thenReturn(false);
        when(repository.save(any(User.class))).thenAnswer(invocation -> invocation.getArgument(0));

        User result = service.create(new UserRequest(" Jiten ", " JITEN@EXAMPLE.COM "));

        assertEquals("Jiten", result.getName());
        assertEquals("jiten@example.com", result.getEmail());
        verify(repository).existsByEmail("jiten@example.com");
        verify(repository).save(any(User.class));
    }

    @Test
    void duplicateEmailIsRejected() {
        when(repository.existsByEmail("jiten@example.com")).thenReturn(true);
        assertThrows(IllegalArgumentException.class,
                () -> service.create(new UserRequest("Jiten", "jiten@example.com")));
        verify(repository, never()).save(any());
    }

    @Test
    void findExistingUser() {
        User user = new User("Jiten", "jiten@example.com");
        when(repository.findById(1L)).thenReturn(Optional.of(user));
        assertSame(user, service.findById(1L));
    }

    @Test
    void missingUserThrowsException() {
        when(repository.findById(99L)).thenReturn(Optional.empty());
        assertThrows(NoSuchElementException.class, () -> service.findById(99L));
    }
}
