package main

import "github.com/go-chi/chi"

func (r *ProductStore) SetupRoutes(router chi.Router) {
	router.Route("/api", func(api chi.Router) {
		api.Post("/create_products", r.CreateProduct)
		api.Get("/get_products", r.GetProducts)
		api.Get("/get_product/{id}", r.GetProductByID)
		api.Put("/update_product/{id}", r.UpdateProductByID)
		api.Delete("/delete_product/{id}", r.DeleteProductByID)
	})
}
