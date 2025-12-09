package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"stockpilot/internal/domain"
)

type userRepoMock struct {
	existing  *domain.User
	created   *domain.User
	createErr error
	getErr    error
}

func (m *userRepoMock) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	m.created = user
	if m.createErr != nil {
		return nil, m.createErr
	}
	return user, nil
}

func (m *userRepoMock) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.existing, m.getErr
}

func (m *userRepoMock) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}

func TestRegisterRejectsUnderage(t *testing.T) {
	svc := NewUserService(&userRepoMock{})
	_, err := svc.Register(context.Background(), RegisterInput{
		Email:     "a@b.c",
		Password:  "password",
		Age:       17,
		FirstName: "John",
	})
	require.Error(t, err)
}

func TestRegisterRejectsWeakPassword(t *testing.T) {
	svc := NewUserService(&userRepoMock{})
	_, err := svc.Register(context.Background(), RegisterInput{
		Email:    "a@b.c",
		Password: "short",
		Age:      20,
	})
	require.Error(t, err)
}

func TestRegisterRejectsDuplicate(t *testing.T) {
	repo := &userRepoMock{existing: &domain.User{ID: "u1", Email: "a@b.c"}}
	svc := NewUserService(repo)
	_, err := svc.Register(context.Background(), RegisterInput{
		Email:    "a@b.c",
		Password: "password",
		Age:      25,
	})
	require.Error(t, err)
}

func TestRegisterSuccess(t *testing.T) {
	repo := &userRepoMock{}
	svc := NewUserService(repo)
	user, err := svc.Register(context.Background(), RegisterInput{
		Email:     "a@b.c",
		Password:  "password123",
		Age:       30,
		FirstName: "Jane",
		LastName:  "Doe",
	})
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, repo.created, user)
	require.NotEqual(t, "password123", user.PasswordHash)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")))
	require.Equal(t, "Jane Doe", user.FullName())
}
