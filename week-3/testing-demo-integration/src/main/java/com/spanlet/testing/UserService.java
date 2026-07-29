package com.spanlet.testing;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.util.List;
import java.util.NoSuchElementException;

@Service
@Transactional
public class UserService {
    private final UserRepository repository;

    public UserService(UserRepository repository) {
        this.repository = repository;
    }

    public User create(UserRequest request) {
        validate(request);
        String name = request.name().trim();
        String email = request.email().trim().toLowerCase();
        if (repository.existsByEmail(email)) {
            throw new IllegalArgumentException("Email already exists");
        }
        return repository.save(new User(name, email));
    }

    @Transactional(readOnly = true)
    public User findById(long id) {
        return repository.findById(id)
                .orElseThrow(() -> new NoSuchElementException("User not found: " + id));
    }

    @Transactional(readOnly = true)
    public List<User> findAll() { return repository.findAll(); }

    public void delete(long id) {
        if (!repository.existsById(id)) {
            throw new NoSuchElementException("User not found: " + id);
        }
        repository.deleteById(id);
    }

    private static void validate(UserRequest request) {
        if (request == null || request.name() == null || request.name().isBlank()) {
            throw new IllegalArgumentException("Name is required");
        }
        if (request.email() == null || request.email().isBlank() || !request.email().contains("@")) {
            throw new IllegalArgumentException("Valid email is required");
        }
    }
}
