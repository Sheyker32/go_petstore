package service

import (
	"context"
	"fmt"
	"swagger_petstore/internal/order/repository"
	petRepository "swagger_petstore/internal/pet/repository"

	"swagger_petstore/petstore"
)

type Servicer interface {
	GetInventory(ctx context.Context) (map[string]int32, error)
	PlaceOrder(ctx context.Context, order petstore.Order) error
	DeleteOrder(ctx context.Context, orderId int64) error
	GetOrderById(ctx context.Context, orderId int64) (petstore.Order, error)
}
type OrderService struct {
	pRepository petRepository.PetsRepository
	repository  repository.OrderRepository
}

func NewService(pRepository petRepository.PetsRepository, repository repository.OrderRepository) *OrderService {
	return &OrderService{pRepository: pRepository, repository: repository}
}

func (s *OrderService) GetInventory(ctx context.Context) (map[string]int32, error) {
	return s.repository.GetInventory(ctx)
}

func (s *OrderService) PlaceOrder(ctx context.Context, order petstore.Order) error {
	pet, err := s.pRepository.GetPetById(ctx, *order.PetId)
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
	s.pRepository.UpdatePetWithForm(ctx, *pet.Id, petStatus)
	return nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, orderId int64) error {
	if orderId <= 0 {
		return fmt.Errorf("invalid orderId: must be a positive number")
	}
	return s.repository.DeleteOrder(ctx, orderId)
}

func (s *OrderService) GetOrderById(ctx context.Context, orderId int64) (petstore.Order, error) {
	if orderId <= 0 {
		return petstore.Order{}, fmt.Errorf("invalid orderId: must be a positive number")
	}
	return s.repository.GetOrderById(ctx, orderId)
}
