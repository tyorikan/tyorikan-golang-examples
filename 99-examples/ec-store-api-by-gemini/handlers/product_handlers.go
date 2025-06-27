package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ec-store-api/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Mock data (replace with database)
var products = []models.Product{}

// ListProducts ...
func ListProducts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	category := r.URL.Query().Get("category")

	limit := 10
	offset := 0

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
	}

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
	}

	filteredProducts := []models.Product{}
	for _, p := range products {
		if category == "" || p.Category == category {
			filteredProducts = append(filteredProducts, p)
		}
	}

	// Pagination (basic example)
	start := offset
	end := offset + limit
	if end > len(filteredProducts) {
		end = len(filteredProducts)
	}
	if start > len(filteredProducts) {
		start = len(filteredProducts)
	}
	paginatedProducts := filteredProducts[start:end]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedProducts)
}

// CreateProduct ...
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var productCreate models.ProductCreate
	err := json.NewDecoder(r.Body).Decode(&productCreate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newProduct := models.Product{
		ProductID:   uuid.New(),
		Name:        productCreate.Name,
		Description: productCreate.Description,
		Price:       productCreate.Price,
		Currency:    productCreate.Currency,
		ImageURL:    productCreate.ImageURL,
		Category:    productCreate.Category,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	products = append(products, newProduct)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProduct)
}

// GetProduct ...
func GetProduct(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	for _, p := range products {
		if p.ProductID == productID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(p)
			return
		}
	}

	http.Error(w, "Product not found", http.StatusNotFound)
}

// UpdateProduct ...
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var productUpdate models.ProductUpdate
	err = json.NewDecoder(r.Body).Decode(&productUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i, p := range products {
		if p.ProductID == productID {
			if productUpdate.Name != "" {
				products[i].Name = productUpdate.Name
			}
			if productUpdate.Description != "" {
				products[i].Description = productUpdate.Description
			}
			if productUpdate.Price != 0 {
				products[i].Price = productUpdate.Price
			}
			if productUpdate.Currency != "" {
				products[i].Currency = productUpdate.Currency
			}
			if productUpdate.ImageURL != "" {
				products[i].ImageURL = productUpdate.ImageURL
			}
			if productUpdate.Category != "" {
				products[i].Category = productUpdate.Category
			}
			products[i].UpdatedAt = time.Now()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(products[i])
			return
		}
	}

	http.Error(w, "Product not found", http.StatusNotFound)
}

// DeleteProduct ...
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	for i, p := range products {
		if p.ProductID == productID {
			products = append(products[:i], products[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Product not found", http.StatusNotFound)
}
