package service

import (
	"context"
	"fmt"
	"swagger_petstore/internal/repository"
	"swagger_petstore/middleware"
	"swagger_petstore/petstore"
	"time"

	"github.com/google/uuid"
)

type Servicer interface {
	AddPet(ctx context.Context, pet petstore.Pet) error
	UpdatePet(ctx context.Context, pet petstore.Pet) error
	FindPetsByStatus(ctx context.Context, status petstore.FindPetsByStatusParams) ([]petstore.Pet, error)
	FindPetsByTags(ctx context.Context, status petstore.FindPetsByTagsParams) ([]petstore.Pet, error)
	DeletePet(ctx context.Context, petId int64, params petstore.DeletePetParams) error
	GetPetById(ctx context.Context, petId int64) (petstore.Pet, error)
	UpdatePetWithForm(ctx context.Context, petId int64, params petstore.UpdatePetWithFormParams) error
	GetInventory(ctx context.Context) (map[string]int32, error)
	PlaceOrder(ctx context.Context, order petstore.Order) error
	DeleteOrder(ctx context.Context, orderId int64) error
	GetOrderById(ctx context.Context, orderId int64) (petstore.Order, error)
	CreateUser(ctx context.Context, user petstore.User) error
	CreateUsersWithListInput(ctx context.Context, users []petstore.User) error
	LoginUser(ctx context.Context, params petstore.LoginUserParams) (string, error)
	LogoutUser(ctx context.Context, tokenID string, token string, exp time.Time) error
	DeleteUser(ctx context.Context, username string) error
	GetUserByName(ctx context.Context, username string) (petstore.User, error)
	UpdateUser(ctx context.Context, user petstore.User) error
}
type Service struct {
	repository repository.PetUserRepository
}

func NewService(repository repository.PetUserRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) AddPet(ctx context.Context, pet petstore.Pet) error {
	return s.repository.AddPet(ctx, pet)
}
func (s *Service) UpdatePet(ctx context.Context, pet petstore.Pet) error {
	return s.repository.UpdatePet(ctx, pet)
}

func (s *Service) FindPetsByStatus(ctx context.Context, status petstore.FindPetsByStatusParams) ([]petstore.Pet, error) {
	if status.Status == nil {
		return nil, fmt.Errorf("status required")
	}
	return s.repository.FindPetsByStatus(ctx, status)
}

func (s *Service) FindPetsByTags(ctx context.Context, status petstore.FindPetsByTagsParams) ([]petstore.Pet, error) {
	if status.Tags == nil {
		return nil, fmt.Errorf("tags required")
	}
	return s.repository.FindPetsByTags(ctx, status)
}

func (s *Service) DeletePet(ctx context.Context, petId int64, params petstore.DeletePetParams) error {
	if petId <= 0 {
		return fmt.Errorf("invalid petId: must be a positive number")
	}
	return s.repository.DeletePet(ctx, petId, params)
}
func (s *Service) GetPetById(ctx context.Context, petId int64) (petstore.Pet, error) {
	if petId <= 0 {
		return petstore.Pet{}, fmt.Errorf("invalid petId: must be a positive number")
	}
	return s.repository.GetPetById(ctx, petId)
}

func (s *Service) UpdatePetWithForm(ctx context.Context, petId int64, params petstore.UpdatePetWithFormParams) error {
	if petId <= 0 {
		return fmt.Errorf("invalid petId: must be a positive number")
	}
	if params.Name == nil && params.Status == nil {
		return fmt.Errorf("name or status required")
	}
	return s.repository.UpdatePetWithForm(ctx, petId, params)
}

func (s *Service) GetInventory(ctx context.Context) (map[string]int32, error) {
	return s.repository.GetInventory(ctx)
}

func (s *Service) PlaceOrder(ctx context.Context, order petstore.Order) error {
	pet, err := s.repository.GetPetById(ctx, *order.PetId)

	if err != nil {
		return fmt.Errorf("pet not found")
	}
	if *pet.Status == petstore.PetStatusSold {
		return fmt.Errorf("pet sold out")
	}

	err = s.repository.PlaceOrder(ctx, order)
	if err != nil {
		return err
	}
	sold := string(petstore.PetStatusSold)
	petStatus := petstore.UpdatePetWithFormParams{
		Name:   &pet.Name,
		Status: &sold,
	}
	s.repository.UpdatePetWithForm(ctx, *pet.Id, petStatus)
	return nil
}

func (s *Service) DeleteOrder(ctx context.Context, orderId int64) error {
	if orderId <= 0 {
		return fmt.Errorf("invalid orderId: must be a positive number")
	}
	return s.repository.DeleteOrder(ctx, orderId)
}

func (s *Service) GetOrderById(ctx context.Context, orderId int64) (petstore.Order, error) {
	if orderId <= 0 {
		return petstore.Order{}, fmt.Errorf("invalid orderId: must be a positive number")
	}
	return s.repository.GetOrderById(ctx, orderId)
}

func (s *Service) CreateUser(ctx context.Context, user petstore.User) error {
	if user.Username != nil {
		name := *user.Username
		_, err := s.repository.GetUserByName(ctx, name)
		if err == nil {
			return fmt.Errorf("user with that name already exists")
		}
	}

	return s.repository.CreateUser(ctx, user)
}

func (s *Service) CreateUsersWithListInput(ctx context.Context, users []petstore.User) error {
	if users == nil {
		return fmt.Errorf("need list of users")
	}
	return s.repository.CreateUsersWithListInput(ctx, users)
}

func (s *Service) LoginUser(ctx context.Context, params petstore.LoginUserParams) (string, error) {
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

func (s *Service) LogoutUser(ctx context.Context, tokenID string, token string, exp time.Time) error {
	if token == "" {
		return fmt.Errorf("token required")
	}
	return s.repository.LogoutUser(ctx, tokenID, token, exp)
}

func (s *Service) DeleteUser(ctx context.Context, username string) error {
	if username == "" {
		return fmt.Errorf("username required")
	}
	return s.repository.DeleteUser(ctx, username)
}

func (s *Service) GetUserByName(ctx context.Context, username string) (petstore.User, error) {
	if username == "" {
		return petstore.User{}, fmt.Errorf("username required")
	}
	return s.repository.GetUserByName(ctx, username)
}

func (s *Service) UpdateUser(ctx context.Context, user petstore.User) error {
	return s.repository.UpdateUser(ctx, user)
}
