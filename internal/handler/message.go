package handler

import "swagger_petstore/petstore"

type LoginData struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

type AuthResponse struct {
	Success   bool      `json:"success"`
	ErrorCode int       `json:"error_code,omitempty"`
	Data      LoginData `json:"data"`
}

type ErrorResponse struct {
	Success   bool `json:"success"`
	ErrorCode int  `json:"error_code,omitempty"`
	Data      Data `json:"data"`
}

type ResponseUser struct {
	Success bool     `json:"success"`
	Data    UserData `json:"data"`
}
type UserData struct {
	Message string        `json:"message"`
	User    petstore.User `json:"user"`
}

type ResponseUsers struct {
	Success bool      `json:"success"`
	Data    UsersData `json:"data"`
}
type UsersData struct {
	Message string          `json:"message"`
	Users   []petstore.User `json:"users"`
}

type ResponseData struct {
	Success bool `json:"success"`
	Data    Data `json:"data"`
}
type ResponsePets struct {
	Success bool     `json:"success"`
	Data    PetsData `json:"data"`
}

type PetsData struct {
	Message string         `json:"message"`
	Pets    []petstore.Pet `json:"pets"`
}

type ResponsePet struct {
	Success bool    `json:"success"`
	Data    PetData `json:"data"`
}

type ResponseInventory struct {
	Success bool          `json:"success"`
	Data    InventoryData `json:"data"`
}

type PetData struct {
	Message string       `json:"message"`
	Pet     petstore.Pet `json:"pet"`
}

type ResponseOrder struct {
	Success bool      `json:"success"`
	Data    OrderData `json:"data"`
}

type OrderData struct {
	Message string         `json:"message"`
	Order   petstore.Order `json:"order"`
}

type InventoryData struct {
	Message   string           `json:"message"`
	Inventory map[string]int32 `json:"inventory"`
}

type Data struct {
	Message string `json:"message"`
}
