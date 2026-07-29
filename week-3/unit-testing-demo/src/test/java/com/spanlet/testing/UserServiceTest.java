package com.spanlet.testing;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.NoSuchElementException;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UserServiceTest {

    @Mock
    private UserRepository userRepository;

    @Mock
    private NotificationService notificationService;

    private UserService userService;

    @BeforeEach
    void setUp() {
        userService = new UserService(userRepository, notificationService);
    }

    @Test
    void getUserShouldReturnUserWhenFound() {
        User expected = new User(1L, "Jiten", "jiten@example.com");
        when(userRepository.findById(1L)).thenReturn(Optional.of(expected));

        User actual = userService.getUser(1L);

        assertEquals(expected, actual);
        verify(userRepository).findById(1L);
        verifyNoInteractions(notificationService);
    }

    @Test
    void getUserShouldThrowWhenUserDoesNotExist() {
        when(userRepository.findById(99L)).thenReturn(Optional.empty());

        NoSuchElementException exception = assertThrows(
                NoSuchElementException.class,
                () -> userService.getUser(99L)
        );

        assertEquals("User not found with id: 99", exception.getMessage());
        verify(userRepository).findById(99L);
    }

    @Test
    void registerShouldSaveUserAndSendWelcomeEmail() {
        User newUser = new User(0L, "Jiten", "jiten@example.com");
        User savedUser = new User(101L, "Jiten", "jiten@example.com");

        when(userRepository.save(newUser)).thenReturn(savedUser);

        User result = userService.register(newUser);

        assertEquals(savedUser, result);
        verify(userRepository).save(newUser);
        verify(notificationService).sendWelcomeEmail(savedUser);
    }

    @Test
    void registerShouldPassSavedUserToNotificationService() {
        User input = new User(0L, "Jiten", "jiten@example.com");
        User saved = new User(501L, "Jiten", "jiten@example.com");

        when(userRepository.save(input)).thenReturn(saved);

        userService.register(input);

        ArgumentCaptor<User> captor = ArgumentCaptor.forClass(User.class);
        verify(notificationService).sendWelcomeEmail(captor.capture());

        assertEquals(501L, captor.getValue().id());
    }

    @Test
    void registerShouldNotNotifyWhenRepositoryFails() {
        User user = new User(0L, "Jiten", "jiten@example.com");
        when(userRepository.save(user))
                .thenThrow(new RuntimeException("Database unavailable"));

        assertThrows(RuntimeException.class, () -> userService.register(user));

        verify(userRepository).save(user);
        verifyNoInteractions(notificationService);
    }
}
