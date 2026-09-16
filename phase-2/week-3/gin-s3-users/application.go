package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type Application struct {
	S3Client *s3.Client
	Bucket   string
}

func main() {

	ctx := context.Background()

	// ---------------------------------------------------------
	// Read configuration
	// ---------------------------------------------------------

	bucket := os.Getenv("S3_BUCKET_NAME")

	if bucket == "" {
		log.Fatal("S3_BUCKET_NAME environment variable is required")
	}

	// ---------------------------------------------------------
	// Load AWS configuration
	// ---------------------------------------------------------

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load AWS configuration: %v", err)
	}

	// ---------------------------------------------------------
	// Create S3 client
	// ---------------------------------------------------------

	s3Client := s3.NewFromConfig(cfg)

	app := &Application{
		S3Client: s3Client,
		Bucket:   bucket,
	}

	// ---------------------------------------------------------
	// Gin router
	// ---------------------------------------------------------

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Gin + AWS S3 User API",
		})
	})

	router.GET("/health", app.health)

	api := router.Group("/api")
	{
		api.POST("/users", app.createUser)

		api.GET("/users", app.getUsers)

		api.GET("/users/:id", app.getUser)

		api.PUT("/users/:id", app.updateUser)

		api.DELETE("/users/:id", app.deleteUser)
	}

	// ---------------------------------------------------------
	// Elastic Beanstalk provides PORT.
	// For local testing we'll use 5000.
	// ---------------------------------------------------------

	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}

	log.Printf("Application starting on port %s", port)
	log.Printf("S3 bucket: %s", bucket)

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// ============================================================
// Health Check
// ============================================================

func (app *Application) health(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}

// ============================================================
// CREATE USER
//
// POST /api/users
// ============================================================

func (app *Application) createUser(c *gin.Context) {

	var request CreateUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	id, err := generateID()
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate user id",
		})

		return
	}

	now := time.Now().UTC()

	user := User{
		ID:        id,
		Name:      request.Name,
		Email:     request.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := app.saveUser(c.Request.Context(), user); err != nil {

		log.Printf("failed to create user: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save user to S3",
		})

		return
	}

	c.JSON(http.StatusCreated, user)
}

// ============================================================
// GET USER
//
// GET /api/users/:id
// ============================================================

func (app *Application) getUser(c *gin.Context) {

	id := c.Param("id")

	user, err := app.readUser(
		c.Request.Context(),
		id,
	)

	if err != nil {

		var noSuchKey *types.NoSuchKey

		if errors.As(err, &noSuchKey) {

			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})

			return
		}

		log.Printf("failed to get user: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read user from S3",
		})

		return
	}

	c.JSON(http.StatusOK, user)
}

// ============================================================
// GET ALL USERS
//
// GET /api/users
// ============================================================

func (app *Application) getUsers(c *gin.Context) {

	ctx := c.Request.Context()

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(app.Bucket),
		Prefix: aws.String("users/"),
	}

	paginator := s3.NewListObjectsV2Paginator(
		app.S3Client,
		input,
	)

	users := make([]User, 0)

	for paginator.HasMorePages() {

		page, err := paginator.NextPage(ctx)
		if err != nil {

			log.Printf("failed listing users: %v", err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to list S3 objects",
			})

			return
		}

		for _, object := range page.Contents {

			if object.Key == nil {
				continue
			}

			key := *object.Key

			if !strings.HasSuffix(key, ".json") {
				continue
			}

			id := strings.TrimPrefix(key, "users/")
			id = strings.TrimSuffix(id, ".json")

			user, err := app.readUser(ctx, id)

			if err != nil {

				log.Printf(
					"unable to read object %s: %v",
					key,
					err,
				)

				continue
			}

			users = append(users, user)
		}
	}

	c.JSON(http.StatusOK, users)
}

// ============================================================
// UPDATE USER
//
// PUT /api/users/:id
// ============================================================

func (app *Application) updateUser(c *gin.Context) {

	id := c.Param("id")

	var request UpdateUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	// First verify that the user exists.

	existingUser, err := app.readUser(
		c.Request.Context(),
		id,
	)

	if err != nil {

		var noSuchKey *types.NoSuchKey

		if errors.As(err, &noSuchKey) {

			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read user",
		})

		return
	}

	existingUser.Name = request.Name
	existingUser.Email = request.Email
	existingUser.UpdatedAt = time.Now().UTC()

	if err := app.saveUser(
		c.Request.Context(),
		existingUser,
	); err != nil {

		log.Printf("failed updating user: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update user",
		})

		return
	}

	c.JSON(http.StatusOK, existingUser)
}

// ============================================================
// DELETE USER
//
// DELETE /api/users/:id
// ============================================================

func (app *Application) deleteUser(c *gin.Context) {

	id := c.Param("id")

	// Verify user exists before deleting.

	_, err := app.readUser(
		c.Request.Context(),
		id,
	)

	if err != nil {

		var noSuchKey *types.NoSuchKey

		if errors.As(err, &noSuchKey) {

			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to check user",
		})

		return
	}

	key := userKey(id)

	_, err = app.S3Client.DeleteObject(
		c.Request.Context(),
		&s3.DeleteObjectInput{
			Bucket: aws.String(app.Bucket),
			Key:    aws.String(key),
		},
	)

	if err != nil {

		log.Printf("failed deleting user: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete user",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user deleted successfully",
		"id":      id,
	})
}

// ============================================================
// Save user to S3
// ============================================================

func (app *Application) saveUser(
	ctx context.Context,
	user User,
) error {

	data, err := json.MarshalIndent(
		user,
		"",
		"  ",
	)

	if err != nil {
		return err
	}

	key := userKey(user.ID)

	_, err = app.S3Client.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(app.Bucket),
			Key:         aws.String(key),
			Body:        bytes.NewReader(data),
			ContentType: aws.String("application/json"),
		},
	)

	return err
}

// ============================================================
// Read user from S3
// ============================================================

func (app *Application) readUser(
	ctx context.Context,
	id string,
) (User, error) {

	var user User

	key := userKey(id)

	result, err := app.S3Client.GetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(app.Bucket),
			Key:    aws.String(key),
		},
	)

	if err != nil {
		return user, err
	}

	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)

	if err != nil {
		return user, err
	}

	if err := json.Unmarshal(data, &user); err != nil {
		return user, err
	}

	return user, nil
}

// ============================================================
// Build S3 key
// ============================================================

func userKey(id string) string {

	return fmt.Sprintf(
		"users/%s.json",
		id,
	)
}

// ============================================================
// Generate random ID
// ============================================================

func generateID() (string, error) {

	bytes := make([]byte, 8)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// demo-users-2026
// GinUsersS3Polic
