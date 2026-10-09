package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alex-zaleskih/interview/internal/model"
)

// ErrInvalidCredentials is returned when login fails.
var ErrInvalidCredentials = errors.New("invalid credentials")

type UserRepository interface {
	FindByID(ctx context.Context, id string) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
}

type UserService struct {
	users UserRepository
}

func NewUserService(users UserRepository) *UserService {
	return &UserService{users: users}
}

// Login authenticates a user by email.
// This is a demo application, so there are no passwords: knowing the email is enough.
func (s *UserService) Login(ctx context.Context, email string) (model.User, error) {
	user, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, model.ErrNotFound) {
		return model.User{}, ErrInvalidCredentials
	}
	if err != nil {
		return model.User{}, fmt.Errorf("login: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (model.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return model.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}
