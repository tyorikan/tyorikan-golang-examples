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
var customers = []models.Customer{}

// ListCustomers ...
func ListCustomers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

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

	// Pagination (basic example)
	start := offset
	end := offset + limit
	if end > len(customers) {
		end = len(customers)
	}
	if start > len(customers) {
		start = len(customers)
	}
	paginatedCustomers := customers[start:end]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedCustomers)
}

// CreateCustomer ...
func CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customerCreate models.CustomerCreate
	err := json.NewDecoder(r.Body).Decode(&customerCreate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newCustomer := models.Customer{
		CustomerID: uuid.New(),
		FirstName:  customerCreate.FirstName,
		LastName:   customerCreate.LastName,
		Email:      customerCreate.Email,
		Phone:      customerCreate.Phone,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	customers = append(customers, newCustomer)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCustomer)
}

// GetCustomer ...
func GetCustomer(w http.ResponseWriter, r *http.Request) {
	customerIDStr := chi.URLParam(r, "customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	for _, c := range customers {
		if c.CustomerID == customerID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(c)
			return
		}
	}

	http.Error(w, "Customer not found", http.StatusNotFound)
}

// UpdateCustomer ...
func UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	customerIDStr := chi.URLParam(r, "customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	var customerUpdate models.CustomerUpdate
	err = json.NewDecoder(r.Body).Decode(&customerUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i, c := range customers {
		if c.CustomerID == customerID {
			if customerUpdate.FirstName != "" {
				customers[i].FirstName = customerUpdate.FirstName
			}
			if customerUpdate.LastName != "" {
				customers[i].LastName = customerUpdate.LastName
			}
			if customerUpdate.Email != "" {
				customers[i].Email = customerUpdate.Email
			}
			if customerUpdate.Phone != "" {
				customers[i].Phone = customerUpdate.Phone
			}
			customers[i].UpdatedAt = time.Now()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(customers[i])
			return
		}
	}

	http.Error(w, "Customer not found", http.StatusNotFound)
}

// DeleteCustomer ...
func DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	customerIDStr := chi.URLParam(r, "customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	for i, c := range customers {
		if c.CustomerID == customerID {
			customers = append(customers[:i], customers[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Customer not found", http.StatusNotFound)
}
