package repository

import (
	"context"
	"fmt"
	"swagger_petstore/petstore"
	"time"

	"github.com/jmoiron/sqlx"
)

type UsersRepository interface {
	CreateUsersWithListInput(ctx context.Context, users []petstore.User) error
	CreateUser(ctx context.Context, user petstore.User) error
	LoginUser(ctx context.Context, params petstore.LoginUserParams) (petstore.User, error)
	LogoutUser(ctx context.Context, tokenID string, token string, exp time.Time) error
	DeleteUser(ctx context.Context, username string) error
	GetUserByName(ctx context.Context, username string) (petstore.User, error)
	UpdateUser(ctx context.Context, user petstore.User) error
}
type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UsersRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user petstore.User) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO users (id, username, firstName, lastName, password, email, phone, userStatus) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		user.Id, user.Username, user.FirstName, user.LastName, user.Password, user.Email, user.Phone, user.UserStatus)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) CreateUsersWithListInput(ctx context.Context, users []petstore.User) error {
	for _, user := range users {
		err := r.CreateUser(ctx, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRepository) LoginUser(ctx context.Context, params petstore.LoginUserParams) (petstore.User, error) {
	var user petstore.User
	query := `SELECT *
			 FROM users 
			 WHERE username = $1 and password = $2`
	err := r.db.GetContext(ctx, &user, query, params.Username, params.Password)
	if err != nil {
		return petstore.User{}, fmt.Errorf("user was not found: %w", err)
	}
	return user, nil
}

func (r *UserRepository) LogoutUser(ctx context.Context, tokenID string, token string, exp time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO token_blacklist (token_id, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token_id) DO UPDATE
		SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at
	`, tokenID, token, exp)
	return err
}

func (r *UserRepository) DeleteUser(ctx context.Context, username string) error {
	_, err := r.db.Exec("DELETE FROM users WHERE username=$1", username)
	return err
}

func (r *UserRepository) GetUserByName(ctx context.Context, username string) (petstore.User, error) {
	var user petstore.User
	query := `SELECT *
			 FROM users 
			 WHERE username = $1`
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		return petstore.User{}, fmt.Errorf("failed to get order by name: %w", err)
	}
	return user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user petstore.User) error {
	query := `UPDATE users 
			 SET username = $1, firstName = $2, lastName = $3, password = $4, email = $5, phone = $6, userStatus = $7
			 WHERE id = $8`
	_, err := r.db.ExecContext(ctx, query, user.Username, user.FirstName, user.LastName, user.Password, user.Email, user.Phone, user.UserStatus, user.Id)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
