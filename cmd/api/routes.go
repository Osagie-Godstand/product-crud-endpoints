package main

import "github.com/go-chi/chi"

func (r *ProductHandler) setupRoutes(router chi.Router) {
	router.Route("/api", func(api chi.Router) {
		api.Post("/create_products", r.createProduct)
		api.Get("/get_products", r.getProducts)
		api.Get("/get_product/{id}", r.getProductByID)
		api.Put("/update_product/{id}", r.updateProductByID)
		api.Delete("/delete_product/{id}", r.deleteProductByID)
	})
}
