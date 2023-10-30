package main

import (
	"database/sql"
	"log"

	"github.com/Osagie-Godstand/crud-product-endpoints/types"
)

func CreateNewProducts(db *sql.DB) {
	newProducts := []types.Product{
		{
			Name:        "Levis Jeans",
			Description: "Navy Blue Denim",
			Price:       79.99,
			SKU:         "799999",
		},
		{
			Name:        "Nike Sneakers",
			Description: "Black Running Shoes",
			Price:       129.99,
			SKU:         "129999",
		},
	}

	for _, product := range newProducts {
		// Check if the product with the same SKU already exists
		existsQuery := "SELECT COUNT(*) FROM products WHERE sku = $1"
		var count int
		err := db.QueryRow(existsQuery, product.SKU).Scan(&count)
		if err != nil {
			log.Println("Error checking product existence:", err)
			continue
		}

		if count > 0 {
			log.Printf("Product with SKU %s already exists, skipping", product.SKU)
			continue
		}

		// Generate a new UUID for the 'id' field
		newID := types.NewUUID()

		// Insert the product with the generated 'id'
		insertQuery := `
			INSERT INTO products (id, name, description, price, sku)
			VALUES ($1, $2, $3, $4, $5)`

		_, err = db.Exec(insertQuery, newID, product.Name, product.Description, product.Price, product.SKU)
		if err != nil {
			log.Println("Error creating product:", err)
		} else {
			log.Println("Product created successfully")
		}
	}
}
