package data

import (
	"github.com/google/uuid"
)

type Product struct {
	Brand       string  `json:"brand"`
	Description string  `json:"description"`
	Colour      string  `json:"colour"`
	Size        string  `json:"size"`
	Price       float64 `json:"price"`
	SKU         string  `json:"sku"`
}

type Products struct {
	ID          uuid.UUID `json:"id"`
	Brand       string    `json:"brand"`
	Description string    `json:"description"`
	Colour      string    `json:"colour"`
	Size        string    `json:"size"`
	Price       float64   `json:"price"`
	SKU         string    `json:"sku"`
}

func NewUUID() uuid.UUID {
	return uuid.New()
}

type ProductID struct {
	ID uuid.UUID `json:"id"`
}
