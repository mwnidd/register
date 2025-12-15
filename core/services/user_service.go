package services

import (
	"context"
	"errors"
	"register/core/ports"
	"register/model"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      ports.UserRepository
	jwtSecret []byte
}

func NewUserService(repo ports.UserRepository, secret string) ports.UserService {
	return &userService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

func (s *userService) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:      name,
		Email:     email,
		Password:  string(hashed),
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}

func (s *userService) GetUser(ctx context.Context, id string) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.repo.List(ctx)
}

func (s *userService) UpdateUser(ctx context.Context, id, name, email string) (*model.User, error) {
	return s.repo.Update(ctx, id, name, email)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *userService) CountUsers(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}
