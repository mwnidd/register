package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"register/model"
	"register/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

const testSecret = "test-secret"

type mockUserService struct {
	users map[string]*model.User
}

func newMockService() *mockUserService {
	return &mockUserService{users: make(map[string]*model.User)}
}

func (m *mockUserService) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	id := email // deterministic for tests
	user := &model.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
	}
	m.users[id] = user
	return user, nil
}

func (m *mockUserService) Login(ctx context.Context, email, password string) (string, error) {
	if u, ok := m.users[email]; ok && u.Password == password {
		return signToken(u.ID), nil
	}
	return "", fiber.ErrUnauthorized
}

func (m *mockUserService) GetUser(ctx context.Context, id string) (*model.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, fiber.ErrNotFound
}

func (m *mockUserService) ListUsers(ctx context.Context) ([]*model.User, error) {
	var res []*model.User
	for _, u := range m.users {
		res = append(res, u)
	}
	return res, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, id, name, email string) (*model.User, error) {
	if u, ok := m.users[id]; ok {
		u.Name, u.Email = name, email
		return u, nil
	}
	return nil, fiber.ErrNotFound
}

func (m *mockUserService) DeleteUser(ctx context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return fiber.ErrNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserService) CountUsers(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func signToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	s, _ := token.SignedString([]byte(testSecret))
	return s
}

func setupApp() *fiber.App {
	svc := newMockService()
	h := NewUserHandler(svc)
	app := fiber.New()
	app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("ok") })
	app.Post("/register", h.Register)
	app.Post("/login", h.Login)

	api := app.Group("/api", middleware.Auth(testSecret))
	api.Get("/users", h.List)
	api.Get("/users/:id", h.Get)
	api.Put("/users/:id", h.Update)
	api.Delete("/users/:id", h.Delete)

	// seed one user for protected routes
	svc.Register(context.Background(), "Seed", "seed@example.com", "pass")
	return app
}

func authedReq(method, path string, body []byte) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+signToken("seed@example.com"))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHealth(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("health failed: %v status=%d", err, resp.StatusCode)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	app := setupApp()
	body := []byte(`{"name":"Alice","email":"alice@example.com","password":"secret"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 201 {
		t.Fatalf("register failed: %v status=%d", err, resp.StatusCode)
	}

	loginBody := []byte(`{"email":"alice@example.com","password":"secret"}`)
	loginReq := httptest.NewRequest("POST", "/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	if err != nil || loginResp.StatusCode != 200 {
		t.Fatalf("login failed: %v status=%d", err, loginResp.StatusCode)
	}
}

func TestListAndGet(t *testing.T) {
	app := setupApp()
	req := authedReq("GET", "/api/users", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("list failed: %v status=%d", err, resp.StatusCode)
	}

	getReq := authedReq("GET", "/api/users/seed@example.com", nil)
	getResp, err := app.Test(getReq)
	if err != nil || getResp.StatusCode != 200 {
		t.Fatalf("get failed: %v status=%d", err, getResp.StatusCode)
	}
}

func TestUpdate(t *testing.T) {
	app := setupApp()
	body, _ := json.Marshal(map[string]string{"name": "Updated", "email": "updated@example.com"})
	req := authedReq("PUT", "/api/users/seed@example.com", body)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("update failed: %v status=%d", err, resp.StatusCode)
	}
}

func TestDelete(t *testing.T) {
	app := setupApp()
	req := authedReq("DELETE", "/api/users/seed@example.com", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 204 {
		t.Fatalf("delete failed: %v status=%d", err, resp.StatusCode)
	}
}
