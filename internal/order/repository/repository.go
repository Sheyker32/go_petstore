package repository

import (
	"context"
	"fmt"
	"swagger_petstore/petstore"

	"github.com/jmoiron/sqlx"
)

type OrderRepository interface {
	GetInventory(ctx context.Context) (map[string]int32, error)
	PlaceOrder(ctx context.Context, order petstore.Order) error
	DeleteOrder(ctx context.Context, orderId int64) error
	GetOrderById(ctx context.Context, orderId int64) (petstore.Order, error)
}
type Repository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &Repository{db: db}
}

func (r *Repository) GetInventory(ctx context.Context) (map[string]int32, error) {
	var pets []petstore.Pet
	res := make(map[string]int32)

	query := `SELECT status
			 FROM pets`

	err := r.db.Select(&pets, query)
	if err != nil {
		return res, fmt.Errorf("failed to get pets: %w", err)
	}

	for _, pet := range pets {
		res[string(*pet.Status)]++
	}

	return res, nil
}

func (r *Repository) PlaceOrder(ctx context.Context, order petstore.Order) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO orders (id, complete, petId, quantity, shipDate, status) VALUES ($1, $2, $3, $4, $5, $6)",
		order.Id, order.Complete, order.PetId, order.Quantity, order.ShipDate, order.Status)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (r *Repository) DeleteOrder(ctx context.Context, orderId int64) error {
	_, err := r.db.Exec("DELETE FROM orders WHERE id=$1", orderId)
	return err
}

func (r *Repository) GetOrderById(ctx context.Context, orderId int64) (petstore.Order, error) {
	var order petstore.Order
	query := `SELECT *
			 FROM orders 
			 WHERE id = $1`
	err := r.db.GetContext(ctx, &order, query, orderId)
	if err != nil {
		return petstore.Order{}, fmt.Errorf("failed to get order by ID: %w", err)
	}
	return order, nil
}
