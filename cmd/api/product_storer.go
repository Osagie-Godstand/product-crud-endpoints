package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Osagie-Godstand/crud-product-endpoints/types"
	"github.com/go-chi/chi"
)

type ProductStore struct {
	DB *sql.DB
}

func (r *ProductStore) CreateProduct(w http.ResponseWriter, req *http.Request) {
	products := []types.Product{}

	err := json.NewDecoder(req.Body).Decode(&products)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	// Creating a buffered channel to limit concurrency
	maxConcurrent := 11
	concurrencyLimiter := make(chan struct{}, maxConcurrent)

	// Creating a buffered channel to communicate errors
	errorChannel := make(chan error, len(products))

	// Using waitgroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Using a context for graceful shutdown
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	// Assuming r.DB is a connection pool instance
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Iterating over each product and inserting into the database
	for _, product := range products {
		// Input validation
		if product.Name == "" || product.Price <= 0 {
			http.Error(w, "Create Product Request Failed: Invalid input data", http.StatusBadRequest)
			return
		}

		concurrencyLimiter <- struct{}{} // Acquiring a concurrency limiter slot
		wg.Add(1)
		go func(p types.Product) {
			defer func() { <-concurrencyLimiter }() // Releasing a concurrency limiter slot
			defer wg.Done()                         // Marking the end of the goroutine

			select {
			case <-ctx.Done():
				return // Aborting if the parent context is canceled
			default:
				query := `
					INSERT INTO products (name, description, price, sku)
					VALUES ($1, $2, $3, $4)`

				_, err := tx.ExecContext(ctx, query, p.Name, p.Description, p.Price, p.SKU)
				if err != nil {
					errorChannel <- fmt.Errorf("failed to insert product: %s (%s)", p.Name, err.Error())
				}
			}
		}(product)
	}

	// Waiting for all goroutines to finish
	wg.Wait()

	// Closing the errorChannel to signal that all errors are received
	close(errorChannel)

	// Committing the transaction if no errors occurred
	if len(errorChannel) == 0 {
		if err := tx.Commit(); err != nil {
			http.Error(w, "Transaction Commit Failed", http.StatusInternalServerError)
			return
		}
	}

	// Collecting errors and respond accordingly
	for err := range errorChannel {
		http.Error(w, fmt.Sprintf("Failed to insert product: %s", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Products have been added"}
	json.NewEncoder(w).Encode(response)
}

func (r *ProductStore) GetProducts(w http.ResponseWriter, req *http.Request) {
	productTypes := []types.Products{}

	rows, err := r.DB.Query("SELECT * FROM products")
	if err != nil {
		http.Error(w, "Get Product Request Failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var product types.Products
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.SKU); err != nil {
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

	var product types.Products
	err := r.DB.QueryRow("SELECT * FROM products WHERE id = $1", id).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.SKU)
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

	var product types.Products
	err := json.NewDecoder(req.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	if id == "" {
		http.Error(w, "Cannot Request Product Update Without ID", http.StatusBadRequest)
		return
	}

	_, err = r.DB.Exec("UPDATE products SET name = $1, description = $2, price = $3, sku = $4 WHERE id = $5", product.Name, product.Description, product.Price, product.SKU, id)
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

	if id == "" {
		http.Error(w, "Product Deletion Needs ID Input", http.StatusBadRequest)
		return
	}

	_, err := r.DB.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Product Not Deleted", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted successfully"})
}
