package service

import (
	"context"
	"fmt"
	"swagger_petstore/internal/pet/repository"
	"swagger_petstore/petstore"
)

type Servicer interface {
	AddPet(ctx context.Context, pet petstore.Pet) error
	UpdatePet(ctx context.Context, pet petstore.Pet) error
	FindPetsByStatus(ctx context.Context, status petstore.FindPetsByStatusParams) ([]petstore.Pet, error)
	FindPetsByTags(ctx context.Context, status petstore.FindPetsByTagsParams) ([]petstore.Pet, error)
	DeletePet(ctx context.Context, petId int64, params petstore.DeletePetParams) error
	GetPetById(ctx context.Context, petId int64) (petstore.Pet, error)
	UpdatePetWithForm(ctx context.Context, petId int64, params petstore.UpdatePetWithFormParams) error
}
type Service struct {
	repository repository.PetsRepository
}

func PetService(repository repository.PetsRepository) *Service {
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
