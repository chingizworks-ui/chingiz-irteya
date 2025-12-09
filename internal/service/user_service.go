package service

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"stockpilot/internal/domain"
	"stockpilot/pkg/gonerve/errors"
)

type RegisterInput struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
	Age       int
	IsMarried bool
}

type UserService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	if input.Age < 18 {
		return nil, errors.New("user must be at least 18")
	}
	if len(input.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	if input.Email == "" {
		return nil, errors.New("email is required")
	}
	existing, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "hash password")
	}
	user := domain.User{
		Email:        input.Email,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Age:          input.Age,
		IsMarried:    input.IsMarried,
		PasswordHash: string(hash),
	}
	created, err := s.users.CreateUser(ctx, &user)
	if err != nil {
		return nil, err
	}
	return created, nil
}
