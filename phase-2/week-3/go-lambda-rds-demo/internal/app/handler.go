package app

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "strings"

    "github.com/aws/aws-lambda-go/events"
    "go-lambda-rds-demo/internal/db"
)

type Service struct{ store db.UserStore }
func New(store db.UserStore) *Service { return &Service{store: store} }

type createUserRequest struct {
    Name string `json:"name"`
    Email string `json:"email"`
}

func (s *Service) Handle(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
    method := req.RequestContext.HTTP.Method
    path := req.RawPath
    if path == "" { path = "/" }

    switch {
    case method == http.MethodGet && path == "/health":
        return jsonResponse(http.StatusOK, map[string]string{"status":"ok"})
    case method == http.MethodPost && path == "/users":
        var in createUserRequest
        if err := json.Unmarshal([]byte(req.Body), &in); err != nil {
            return jsonResponse(http.StatusBadRequest, map[string]string{"error":"invalid JSON"})
        }
        in.Name, in.Email = strings.TrimSpace(in.Name), strings.TrimSpace(in.Email)
        if in.Name == "" || in.Email == "" {
            return jsonResponse(http.StatusBadRequest, map[string]string{"error":"name and email are required"})
        }
        u, err := s.store.Create(ctx, in.Name, in.Email)
        if err != nil { return dbError(err) }
        return jsonResponse(http.StatusCreated, u)
    case method == http.MethodGet && path == "/users":
        users, err := s.store.List(ctx)
        if err != nil { return dbError(err) }
        return jsonResponse(http.StatusOK, users)
    default:
        return jsonResponse(http.StatusNotFound, map[string]string{"error":"route not found"})
    }
}

func dbError(err error) (events.LambdaFunctionURLResponse, error) {
    if err == nil { err = errors.New("unknown database error") }
    // Do not leak database details to clients.
    return jsonResponse(http.StatusInternalServerError, map[string]string{"error":"database operation failed"})
}

func jsonResponse(status int, v any) (events.LambdaFunctionURLResponse, error) {
    b, err := json.Marshal(v)
    if err != nil { return events.LambdaFunctionURLResponse{}, err }
    return events.LambdaFunctionURLResponse{
        StatusCode: status,
        Headers: map[string]string{"content-type":"application/json"},
        Body: string(b),
    }, nil
}
