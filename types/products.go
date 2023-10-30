package types

import "github.com/google/uuid"

type Product struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SKU         string  `json:"sku"`
}

type Products struct {
	ID          uuid.UUID `json:"id"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Price       *float64  `json:"price"`
	SKU         *string   `json:"sku"`
}

func NewUUID() uuid.UUID {
	return uuid.New()
}

type ProductID struct {
	ID uuid.UUID `json:"id"`
}
