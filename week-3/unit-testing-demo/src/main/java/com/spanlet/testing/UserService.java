package com.spanlet.testing;

import java.util.NoSuchElementException;
import java.util.Objects;

public class UserService {

    private final UserRepository userRepository;
    private final NotificationService notificationService;

    public UserService(UserRepository userRepository,
                       NotificationService notificationService) {
        this.userRepository = Objects.requireNonNull(userRepository);
        this.notificationService = Objects.requireNonNull(notificationService);
    }

    public User getUser(long id) {
        return userRepository.findById(id)
                .orElseThrow(() -> new NoSuchElementException(
                        "User not found with id: " + id));
    }

    public User register(User user) {
        Objects.requireNonNull(user, "User must not be null");

        User savedUser = userRepository.save(user);
        notificationService.sendWelcomeEmail(savedUser);
        return savedUser;
    }
}
