package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"register/model"
)

type mockUserRepo struct {
	users map[string]*model.User
	seq   int
}

func newMockRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*model.User)}
}

func (m *mockUserRepo) nextID() string {
	m.seq++
	return strings.TrimSpace(time.Now().Format("150405")) + "-" + string(rune('a'+m.seq-1))
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	user.ID = m.nextID()
	cp := *user
	m.users[user.ID] = &cp
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	if u, ok := m.users[id]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) List(ctx context.Context) ([]*model.User, error) {
	res := make([]*model.User, 0, len(m.users))
	for _, u := range m.users {
		cp := *u
		res = append(res, &cp)
	}
	return res, nil
}

func (m *mockUserRepo) Update(ctx context.Context, id, name, email string) (*model.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	u.Name = name
	u.Email = email
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return errors.New("not found")
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func TestRegisterAndLogin(t *testing.T) {
	repo := newMockRepo()
	svc := NewUserService(repo, "secret")

	user, err := svc.Register(context.Background(), "Alice", "alice@example.com", "password")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if user.ID == "" {
		t.Fatal("expected ID to be set")
	}
	if user.Password == "password" {
		t.Fatal("expected password to be hashed")
	}
	if user.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}

	token, err := svc.Login(context.Background(), "alice@example.com", "password")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}

	if _, err := svc.Login(context.Background(), "alice@example.com", "wrong"); err == nil {
		t.Fatal("expected login with wrong password to fail")
	}
}

func TestCRUD(t *testing.T) {
	repo := newMockRepo()
	svc := NewUserService(repo, "secret")

	user, err := svc.Register(context.Background(), "Bob", "bob@example.com", "p4ss")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	got, err := svc.GetUser(context.Background(), user.ID)
	if err != nil || got.Email != user.Email {
		t.Fatalf("get user failed: %v", err)
	}

	updated, err := svc.UpdateUser(context.Background(), user.ID, "Bobby", "bobby@example.com")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "Bobby" || updated.Email != "bobby@example.com" {
		t.Fatalf("update returned wrong data: %+v", updated)
	}

	list, err := svc.ListUsers(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatalf("list failed: %v len=%d", err, len(list))
	}

	if err := svc.DeleteUser(context.Background(), user.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if _, err := svc.GetUser(context.Background(), user.ID); err == nil {
		t.Fatal("expected missing user after delete")
	}
}
