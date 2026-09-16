package db

import (
    "context"
    "fmt"
    "net/url"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UserStore interface {
    Create(ctx context.Context, name, email string) (User, error)
    List(ctx context.Context) ([]User, error)
}

type PostgresUserStore struct{ pool *pgxpool.Pool }

func NewPostgresUserStore(pool *pgxpool.Pool) *PostgresUserStore { return &PostgresUserStore{pool: pool} }

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
    _, err := pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS users (
            id BIGSERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        )
    `)
    return err
}

func (s *PostgresUserStore) Create(ctx context.Context, name, email string) (User, error) {
    var u User
    err := s.pool.QueryRow(ctx,
        `INSERT INTO users(name,email) VALUES($1,$2) RETURNING id,name,email`,
        name, email,
    ).Scan(&u.ID, &u.Name, &u.Email)
    return u, err
}

func (s *PostgresUserStore) List(ctx context.Context) ([]User, error) {
    rows, err := s.pool.Query(ctx, `SELECT id,name,email FROM users ORDER BY id`)
    if err != nil { return nil, err }
    defer rows.Close()

    users := make([]User, 0)
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil { return nil, err }
        users = append(users, u)
    }
    return users, rows.Err()
}

func ConnectionStringFromEnv() string {
    host := mustEnv("DB_HOST")
    port := envOr("DB_PORT", "5432")
    name := mustEnv("DB_NAME")
    user := mustEnv("DB_USER")
    pass := mustEnv("DB_PASSWORD")
    sslmode := envOr("DB_SSLMODE", "require")

    return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
        url.QueryEscape(user), url.QueryEscape(pass), host, port, name, url.QueryEscape(sslmode))
}

func mustEnv(k string) string {
    v := os.Getenv(k)
    if v == "" { panic("missing environment variable: " + k) }
    return v
}
func envOr(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }
