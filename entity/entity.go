package entity

import "swagger_petstore/petstore"

type PetReq struct {
	Id       *int64              `json:"id"`
	Name     string              `json:"name"`
	Category *petstore.Category  `json:"category"`
	Tags     *[]petstore.Tag     `json:"tags"`
	Status   *petstore.PetStatus `json:"status"`
}
