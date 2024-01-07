package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Osagie-Godstand/product-crud-endpoints/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductStore is a mock implementation of the ProductStore interface.
type MockProductStore struct {
	mock.Mock
}

func (m *MockProductStore) InsertProduct(product model.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

type TestProductHandler struct {
	ProductStore *MockProductStore
}

func (p *TestProductHandler) createProduct(w http.ResponseWriter, req *http.Request) {
	products := []model.Product{}

	err := json.NewDecoder(req.Body).Decode(&products)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	maxConcurrent := 11
	concurrencyLimiter := make(chan struct{}, maxConcurrent)

	ctx, cancel := req.Context(), func() {}
	defer cancel()

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
				if err := p.ProductStore.InsertProduct(product); err != nil {
					errorChannel <- err
				}
			}
		}(product)
	}

	wg.Wait()
	close(errorChannel)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Products have been created"}
	json.NewEncoder(w).Encode(response)
}

func TestCreateProductConcurrent(t *testing.T) {
	// Mocking the product store
	mockProductStore := &MockProductStore{}

	// Creating the ProductHandler with the mock product store
	productHandler := &TestProductHandler{
		ProductStore: mockProductStore,
	}

	// Creating sample products
	products := []model.Product{
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
		{
			Brand:       "Levis",
			Description: "Denim Jeans",
			Colour:      "Navy Blue",
			Size:        "31/32",
			Price:       79.99,
			SKU:         "789999",
		},
		{
			Brand:       "Nike",
			Description: "Air Max 97",
			Colour:      "Black",
			Size:        "10",
			Price:       129.99,
			SKU:         "139999",
		},
	}

	// Set expectations for the mock product store
	mockProductStore.On("InsertProduct", products[0]).Return(nil)
	mockProductStore.On("InsertProduct", products[1]).Return(nil)
	mockProductStore.On("InsertProduct", products[2]).Return(nil)
	mockProductStore.On("InsertProduct", products[3]).Return(nil)

	// Converting products to a JSON string
	requestBody, err := json.Marshal(products)
	if err != nil {
		t.Fatal(err)
	}

	// Creating a request with the JSON body
	req, err := http.NewRequest("POST", "/createProduct", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Creating a response recorder to record the response
	rr := httptest.NewRecorder()

	// Calling the createProduct function to handle the request
	productHandler.createProduct(rr, req)

	// Asserting the HTTP status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Asserting the response body
	expectedResponse := `{"message":"Products have been created"}`
	assert.Equal(t, strings.TrimSpace(expectedResponse), strings.TrimSpace(rr.Body.String()))

	// Verify that the expected methods were called on the mock
	mockProductStore.AssertExpectations(t)
}
