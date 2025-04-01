package main

import (
	"log"
	"net/http"

	"ec-store-api/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Products routes
	r.Route("/products", func(r chi.Router) {
		r.Get("/", handlers.ListProducts)
		r.Post("/", handlers.CreateProduct)
		r.Route("/{product_id}", func(r chi.Router) {
			r.Get("/", handlers.GetProduct)
			r.Put("/", handlers.UpdateProduct)
			r.Delete("/", handlers.DeleteProduct)
		})
	})

	// Orders routes
	r.Route("/orders", func(r chi.Router) {
		r.Get("/", handlers.ListOrders)
		r.Post("/", handlers.CreateOrder)
		r.Route("/{order_id}", func(r chi.Router) {
			r.Get("/", handlers.GetOrder)
			r.Put("/", handlers.UpdateOrder)
			r.Delete("/", handlers.DeleteOrder)
		})
	})

	// Customers routes
	r.Route("/customers", func(r chi.Router) {
		r.Get("/", handlers.ListCustomers)
		r.Post("/", handlers.CreateCustomer)
		r.Route("/{customer_id}", func(r chi.Router) {
			r.Get("/", handlers.GetCustomer)
			r.Put("/", handlers.UpdateCustomer)
			r.Delete("/", handlers.DeleteCustomer)
		})
	})

	// Inventory routes
	r.Route("/inventory", func(r chi.Router) {
		r.Route("/{product_id}", func(r chi.Router) {
			r.Get("/", handlers.GetInventory)
			r.Put("/", handlers.UpdateInventory)
		})
	})

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
