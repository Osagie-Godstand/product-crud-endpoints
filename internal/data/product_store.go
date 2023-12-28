package data

import (
	"database/sql"

	"github.com/google/uuid"
)

type ProductStore struct {
	DB *sql.DB
}

func (d *ProductStore) InsertProduct(product Product) error {
	_, err := d.DB.Exec("INSERT INTO products (id, brand, description, colour, size, price, sku) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		uuid.New(), product.Brand, product.Description, product.Colour, product.Size, product.Price, product.SKU)
	return err
}

func (d *ProductStore) GetProducts() ([]Products, error) {
	rows, err := d.DB.Query("SELECT * FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var productTypes []Products

	for rows.Next() {
		var product Products
		if err := rows.Scan(&product.ID, &product.Brand, &product.Description, &product.Colour, &product.Size, &product.Price, &product.SKU); err != nil {
			return nil, err
		}
		productTypes = append(productTypes, product)
	}

	return productTypes, nil
}

func (d *ProductStore) GetProductByID(id uuid.UUID) (Products, error) {
	var product Products
	err := d.DB.QueryRow("SELECT * FROM products WHERE id = $1", id).Scan(&product.ID, &product.Brand, &product.Description, &product.Colour, &product.Size, &product.Price, &product.SKU)
	return product, err
}

func (d *ProductStore) UpdateProductByID(id uuid.UUID, product Products) error {
	_, err := d.DB.Exec("UPDATE products SET brand = $1, description = $2, colour = $3, size = $4, price = $5, sku = $6 WHERE id = $7",
		product.Brand, product.Description, product.Colour, product.Size, product.Price, product.SKU, id)
	return err
}

func (d *ProductStore) DeleteProductByID(id uuid.UUID) error {
	_, err := d.DB.Exec("DELETE FROM products WHERE id = $1", id)
	return err
}
