package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// ---------------------------------------------------------
// Database initialization
// ---------------------------------------------------------

func initDatabase() error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	sslMode := os.Getenv("DB_SSLMODE")

	if port == "" {
		port = "5432"
	}

	if sslMode == "" {
		sslMode = "require"
	}

	if host == "" {
		return fmt.Errorf("DB_HOST environment variable is missing")
	}

	if dbName == "" {
		return fmt.Errorf("DB_NAME environment variable is missing")
	}

	if user == "" {
		return fmt.Errorf("DB_USER environment variable is missing")
	}

	if password == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is missing")
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user,
		password,
		host,
		port,
		dbName,
		sslMode,
	)

	var err error

	db, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("unable to create database connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("unable to connect to PostgreSQL RDS: %w", err)
	}

	log.Println("Connected successfully to PostgreSQL RDS")

	return createTable()
}

// ---------------------------------------------------------
// Create users table automatically
// ---------------------------------------------------------

func createTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to create users table: %w", err)
	}

	log.Println("users table is ready")

	return nil
}

// ---------------------------------------------------------
// Response helper
// ---------------------------------------------------------

func jsonResponse(
	statusCode int,
	data interface{},
) events.LambdaFunctionURLResponse {

	body, err := json.Marshal(data)

	if err != nil {
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message":"failed to create response"}`,
		}
	}

	return events.LambdaFunctionURLResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

// ---------------------------------------------------------
// GET /health
// ---------------------------------------------------------

func healthHandler(
	ctx context.Context,
) events.LambdaFunctionURLResponse {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		log.Printf("database health check failed: %v", err)

		return jsonResponse(
			500,
			MessageResponse{
				Message: "Lambda is running, but database connection failed",
			},
		)
	}

	return jsonResponse(
		200,
		MessageResponse{
			Message: "Lambda and RDS are healthy",
		},
	)
}

// ---------------------------------------------------------
// POST /users
// ---------------------------------------------------------

func createUserHandler(
	ctx context.Context,
	req events.LambdaFunctionURLRequest,
) events.LambdaFunctionURLResponse {

	var input CreateUserRequest

	if err := json.Unmarshal([]byte(req.Body), &input); err != nil {
		return jsonResponse(
			400,
			MessageResponse{
				Message: "Invalid JSON request",
			},
		)
	}

	if input.Name == "" {
		return jsonResponse(
			400,
			MessageResponse{
				Message: "name is required",
			},
		)
	}

	if input.Email == "" {
		return jsonResponse(
			400,
			MessageResponse{
				Message: "email is required",
			},
		)
	}

	query := `
		INSERT INTO users(name, email)
		VALUES($1, $2)
		RETURNING id, name, email, created_at
	`

	var user User

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := db.QueryRow(
		ctx,
		query,
		input.Name,
		input.Email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		log.Printf("error inserting user: %v", err)

		return jsonResponse(
			500,
			MessageResponse{
				Message: "Unable to create user",
			},
		)
	}

	return jsonResponse(
		201,
		user,
	)
}

// ---------------------------------------------------------
// GET /users
// ---------------------------------------------------------

func getUsersHandler(
	ctx context.Context,
) events.LambdaFunctionURLResponse {

	query := `
		SELECT
			id,
			name,
			email,
			created_at
		FROM users
		ORDER BY id
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, query)

	if err != nil {
		log.Printf("error querying users: %v", err)

		return jsonResponse(
			500,
			MessageResponse{
				Message: "Unable to retrieve users",
			},
		)
	}

	defer rows.Close()

	users := make([]User, 0)

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
		)

		if err != nil {
			log.Printf("error scanning user row: %v", err)

			return jsonResponse(
				500,
				MessageResponse{
					Message: "Unable to read user",
				},
			)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows iteration error: %v", err)

		return jsonResponse(
			500,
			MessageResponse{
				Message: "Unable to retrieve users",
			},
		)
	}

	return jsonResponse(
		200,
		users,
	)
}

// ---------------------------------------------------------
// Main Lambda Function URL handler
// ---------------------------------------------------------

func handler(
	ctx context.Context,
	req events.LambdaFunctionURLRequest,
) (events.LambdaFunctionURLResponse, error) {

	method := req.RequestContext.HTTP.Method
	path := req.RawPath

	log.Printf(
		"FUNCTION URL REQUEST: method=%q path=%q body=%q",
		method,
		path,
		req.Body,
	)

	switch {

	case method == "GET" && path == "/health":

		return jsonResponse(
			200,
			MessageResponse{
				Message: "Lambda is running",
			},
		), nil

	case method == "GET" && path == "/users":

		return getUsersHandler(ctx), nil

	case method == "POST" && path == "/users":

		return createUserHandler(ctx, req), nil

	default:

		return jsonResponse(
			404,
			MessageResponse{
				Message: fmt.Sprintf(
					"Route not found: %s %s",
					method,
					path,
				),
			},
		), nil
	}
}

// ---------------------------------------------------------
// Application entry point
// ---------------------------------------------------------

func main() {
	log.Println("Starting Go AWS Lambda")

	if err := initDatabase(); err != nil {
		log.Fatalf(
			"Database initialization failed: %v",
			err,
		)
	}

	lambda.Start(handler)
}

/*
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jiten",
    "email": "jiten@example.com"
  }' \
  https://3qtxtathn4zzhz65qe7l3a32yi0neoca.lambda-url.eu-north-1.on.aws/users

*/
