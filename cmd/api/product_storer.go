package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Osagie-Godstand/product-crud-endpoints/internal/data"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type ProductStore struct {
	DB *sql.DB
}

func (r *ProductStore) CreateProduct(w http.ResponseWriter, req *http.Request) {
	products := []data.Product{}

	err := json.NewDecoder(req.Body).Decode(&products)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	maxConcurrent := 11
	concurrencyLimiter := make(chan struct{}, maxConcurrent)

	errorChannel := make(chan error, len(products))

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, product := range products {
		if product.Brand == "" || product.Price <= 0 {
			http.Error(w, "Create Product Request Failed: Invalid input data", http.StatusBadRequest)
			return
		}

		concurrencyLimiter <- struct{}{}
		wg.Add(1)
		go func(product data.Product) {
			defer func() { <-concurrencyLimiter }()
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				newID := uuid.New()

				query := `
                    INSERT INTO products (id, brand, description, colour, size, price, sku)
                    VALUES ($1, $2, $3, $4, $5, $6, $7)`

				_, err := tx.ExecContext(ctx, query, newID, product.Brand, product.Description, product.Colour, product.Size, product.Price, product.SKU)
				if err != nil {
					errorChannel <- fmt.Errorf("failed to insert product: %s (%s)", product.Brand, err.Error())
				}
			}
		}(product)
	}

	wg.Wait()

	close(errorChannel)

	if len(errorChannel) == 0 {
		if err := tx.Commit(); err != nil {
			http.Error(w, "Transaction Commit Failed", http.StatusInternalServerError)
			return
		}
	}

	for err := range errorChannel {
		http.Error(w, fmt.Sprintf("Failed to insert product: %s", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Products have been created"}
	json.NewEncoder(w).Encode(response)
}

func (r *ProductStore) GetProducts(w http.ResponseWriter, req *http.Request) {
	productTypes := []data.Products{}

	rows, err := r.DB.Query("SELECT * FROM products")
	if err != nil {
		http.Error(w, "Get Product Request Failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var product data.Products
		if err := rows.Scan(&product.ID, &product.Brand, &product.Description, &product.Colour, &product.Size, &product.Price, &product.SKU); err != nil {
			http.Error(w, "Get Product Request Failed", http.StatusInternalServerError)
			return
		}
		productTypes = append(productTypes, product)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": productTypes})
}

func (r *ProductStore) GetProductByID(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	if id == "" {
		http.Error(w, "Cannot Request Product Without ID", http.StatusBadRequest)
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	var product data.Products
	err = r.DB.QueryRow("SELECT * FROM products WHERE id = $1", parsedID).Scan(&product.ID, &product.Brand, &product.Description, &product.Colour, &product.Size, &product.Price, &product.SKU)
	if err != nil {
		http.Error(w, "Get Product Request Failed", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": product})
}

func (r *ProductStore) UpdateProductByID(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	var product data.Products
	err = json.NewDecoder(req.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	_, err = r.DB.Exec("UPDATE products SET brand = $1, description = $2, colour = $3, size = $4, price = $5, sku = $6 WHERE id = $7", product.Brand, product.Description, product.Colour, product.Size, product.Price, product.SKU, parsedID)
	if err != nil {
		http.Error(w, "Product Not Updated", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Product has been updated"})
}

func (r *ProductStore) DeleteProductByID(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	_, err = r.DB.Exec("DELETE FROM products WHERE id = $1", parsedID)
	if err != nil {
		http.Error(w, "Product Not Deleted", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted successfully"})
}
