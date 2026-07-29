package com.spanlet.testing;

import java.util.Optional;

public interface UserRepository {
    Optional<User> findById(long id);
    User save(User user);
}
