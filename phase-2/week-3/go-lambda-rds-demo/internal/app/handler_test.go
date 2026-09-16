package app

import (
    "context"
    "encoding/json"
    "testing"

    "github.com/aws/aws-lambda-go/events"
    "go-lambda-rds-demo/internal/db"
)

type fakeStore struct{ users []db.User }
func (f *fakeStore) Create(ctx context.Context, name, email string) (db.User, error) {
    u := db.User{ID:int64(len(f.users)+1), Name:name, Email:email}
    f.users = append(f.users, u); return u, nil
}
func (f *fakeStore) List(ctx context.Context) ([]db.User, error) { return f.users, nil }

func request(method, path, body string) events.LambdaFunctionURLRequest {
    return events.LambdaFunctionURLRequest{RawPath:path, Body:body,
        RequestContext: events.LambdaFunctionURLRequestContext{HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{Method:method}}}
}

func TestCreateAndListUsers(t *testing.T) {
    svc := New(&fakeStore{})
    body, _ := json.Marshal(map[string]string{"name":"Jiten","email":"jiten@example.com"})
    created, err := svc.Handle(context.Background(), request("POST","/users",string(body)))
    if err != nil || created.StatusCode != 201 { t.Fatalf("create: status=%d err=%v body=%s", created.StatusCode, err, created.Body) }

    listed, err := svc.Handle(context.Background(), request("GET","/users",""))
    if err != nil || listed.StatusCode != 200 { t.Fatalf("list: status=%d err=%v body=%s", listed.StatusCode, err, listed.Body) }
    if !contains(listed.Body, "jiten@example.com") { t.Fatalf("unexpected list body: %s", listed.Body) }
}
func TestValidation(t *testing.T) {
    svc := New(&fakeStore{})
    r, _ := svc.Handle(context.Background(), request("POST","/users",`{"name":""}`))
    if r.StatusCode != 400 { t.Fatalf("want 400 got %d", r.StatusCode) }
}
func contains(s, sub string) bool { return len(sub) == 0 || (len(s) >= len(sub) && func() bool { for i:=0;i+len(sub)<=len(s);i++ { if s[i:i+len(sub)]==sub { return true } }; return false }()) }
