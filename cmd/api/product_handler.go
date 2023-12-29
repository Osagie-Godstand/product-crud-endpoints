package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Osagie-Godstand/product-crud-endpoints/internal/data"
	"github.com/Osagie-Godstand/product-crud-endpoints/internal/model"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type ProductHandler struct {
	DB *data.ProductStore
}

func (p *ProductHandler) createProduct(w http.ResponseWriter, req *http.Request) {
	products := []model.Product{}

	err := json.NewDecoder(req.Body).Decode(&products)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	maxConcurrent := 11
	concurrencyLimiter := make(chan struct{}, maxConcurrent)

	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	tx, err := p.DB.DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	errorChannel := make(chan error, len(products))
	var wg sync.WaitGroup

	for _, product := range products {
		if product.Brand == "" || product.Price <= 0 {
			http.Error(w, "Create Product Request Failed: Invalid input data", http.StatusBadRequest)
			return
		}

		concurrencyLimiter <- struct{}{}
		wg.Add(1)
		go func(product model.Product) {
			defer func() { <-concurrencyLimiter }()
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				if err := p.DB.InsertProduct(product); err != nil {
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

func (p *ProductHandler) getProducts(w http.ResponseWriter, req *http.Request) {
	productTypes, err := p.DB.GetProducts()
	if err != nil {
		http.Error(w, "Get Product Request Failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": productTypes})
}

func (p *ProductHandler) getProductByID(w http.ResponseWriter, req *http.Request) {
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

	product, err := p.DB.GetProductByID(parsedID)
	if err != nil {
		http.Error(w, "Get Product Request Failed", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": product})
}

func (p *ProductHandler) updateProductByID(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	var product model.Products
	err = json.NewDecoder(req.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	err = p.DB.UpdateProductByID(parsedID, product)
	if err != nil {
		http.Error(w, "Product Not Updated", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Product has been updated"})
}

func (p *ProductHandler) deleteProductByID(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	err = p.DB.DeleteProductByID(parsedID)
	if err != nil {
		http.Error(w, "Product Not Deleted", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted successfully"})
}
