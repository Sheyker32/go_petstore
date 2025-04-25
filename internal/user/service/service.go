package service

import (
	"context"
	"fmt"
	"swagger_petstore/internal/user/repository"
	"swagger_petstore/middleware"
	"swagger_petstore/petstore"
	"time"

	"github.com/google/uuid"
)

type Servicer interface {
	CreateUser(ctx context.Context, user petstore.User) error
	CreateUsersWithListInput(ctx context.Context, users []petstore.User) error
	LoginUser(ctx context.Context, params petstore.LoginUserParams) (string, error)
	LogoutUser(ctx context.Context, tokenID string, token string, exp time.Time) error
	DeleteUser(ctx context.Context, username string) error
	GetUserByName(ctx context.Context, username string) (petstore.User, error)
	UpdateUser(ctx context.Context, user petstore.User) error
}
type UserService struct {
	repository repository.UsersRepository
}

func NewUserService(repository repository.UsersRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) CreateUser(ctx context.Context, user petstore.User) error {
	if user.Username != nil {
		name := *user.Username
		_, err := s.repository.GetUserByName(ctx, name)
		if err == nil {
			return fmt.Errorf("user with that name already exists")
		}
	}

	return s.repository.CreateUser(ctx, user)
}

func (s *UserService) CreateUsersWithListInput(ctx context.Context, users []petstore.User) error {
	if users == nil {
		return fmt.Errorf("need list of users")
	}
	return s.repository.CreateUsersWithListInput(ctx, users)
}

func (s *UserService) LoginUser(ctx context.Context, params petstore.LoginUserParams) (string, error) {
	if params.Username == nil || params.Password == nil {
		return "", fmt.Errorf("need login information")
	}
	tokenAuth := middleware.TokenAuth

	user, err := s.repository.LoginUser(ctx, params)
	if err != nil {
		return "", fmt.Errorf("Unauthorized")
	}
	claims := map[string]interface{}{
		"user_id": user.Username,
		"jti":     uuid.New().String(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	_, token, err := tokenAuth.Encode(claims)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) LogoutUser(ctx context.Context, tokenID string, token string, exp time.Time) error {
	if token == "" {
		return fmt.Errorf("token required")
	}
	return s.repository.LogoutUser(ctx, tokenID, token, exp)
}

func (s *UserService) DeleteUser(ctx context.Context, username string) error {
	if username == "" {
		return fmt.Errorf("username required")
	}
	return s.repository.DeleteUser(ctx, username)
}

func (s *UserService) GetUserByName(ctx context.Context, username string) (petstore.User, error) {
	if username == "" {
		return petstore.User{}, fmt.Errorf("username required")
	}
	return s.repository.GetUserByName(ctx, username)
}

func (s *UserService) UpdateUser(ctx context.Context, user petstore.User) error {
	return s.repository.UpdateUser(ctx, user)
}
