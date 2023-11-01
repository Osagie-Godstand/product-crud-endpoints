package main

import (
	"database/sql"
	"log"

	"github.com/Osagie-Godstand/crud-product-endpoints/types"
)

func CreateNewProducts(db *sql.DB) {
	newProducts := []types.Product{
		{
			Brand:       "Levis",
			Description: "Denim Jeans",
			Colour:      "Navy Blue",
			Size:        "31/32",
			Price:       79.99,
			SKU:         "799999",
		},
		{
			Brand:       "Nike",
			Description: "Air Max 97",
			Colour:      "Black",
			Size:        "10",
			Price:       129.99,
			SKU:         "129999",
		},
	}

	for _, product := range newProducts {
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

		newID := types.NewUUID()

		insertQuery := `
			INSERT INTO products (id, brand, description, colour, size, price, sku)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

		_, err = db.Exec(insertQuery, newID, product.Brand, product.Description, product.Colour, product.Size, product.Price, product.SKU)
		if err != nil {
			log.Println("Error creating product:", err)
		} else {
			log.Println("Product created successfully")
		}
	}
}
