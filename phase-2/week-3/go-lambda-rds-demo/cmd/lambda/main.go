package main

import (
    "context"
    "log"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/jackc/pgx/v5/pgxpool"

    "go-lambda-rds-demo/internal/app"
    "go-lambda-rds-demo/internal/db"
)

func main() {
    ctx := context.Background()
    pool, err := pgxpool.New(ctx, db.ConnectionStringFromEnv())
    if err != nil {
        log.Fatalf("create db pool: %v", err)
    }
    if err := pool.Ping(ctx); err != nil {
        log.Fatalf("ping db: %v", err)
    }
    if err := db.EnsureSchema(ctx, pool); err != nil {
        log.Fatalf("ensure schema: %v", err)
    }

    svc := app.New(db.NewPostgresUserStore(pool))
    lambda.Start(func(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
        return svc.Handle(ctx, req)
    })

}
