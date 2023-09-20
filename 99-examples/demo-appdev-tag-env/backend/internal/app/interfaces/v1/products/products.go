package v1

import (
	"backend/internal/app/interfaces/helper"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ProductResources provide products interfaces
type ProductResources struct {
}

// NewProductResources returns productResources struct
func NewProductResources() *ProductResources {
	return &ProductResources{}
}

// Routes creates a REST router for the product resource
func (rs ProductResources) Routes() chi.Router {
	r := chi.NewRouter()
	r.With().Get("/", rs.list)
	return r
}

type Product struct {
	ID    int
	Name  string
	Price float64
}

func (rs ProductResources) list(w http.ResponseWriter, r *http.Request) {
	products := []Product{
		{ID: 1, Name: "Product 1", Price: 100.0},
		{ID: 2, Name: "Product 2", Price: 200.0},
	}
	helper.Succeed(w, products)
}
