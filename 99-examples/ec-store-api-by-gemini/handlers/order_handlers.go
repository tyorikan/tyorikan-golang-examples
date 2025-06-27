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
var orders = []models.Order{}

// ListOrders ...
func ListOrders(w http.ResponseWriter, r *http.Request) {
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
	if end > len(orders) {
		end = len(orders)
	}
	if start > len(orders) {
		start = len(orders)
	}
	paginatedOrders := orders[start:end]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedOrders)
}

// CreateOrder ...
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var orderCreate models.OrderCreate
	err := json.NewDecoder(r.Body).Decode(&orderCreate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newOrder := models.Order{
		OrderID:         uuid.New(),
		CustomerID:      orderCreate.CustomerID,
		OrderDate:       time.Now(),
		Status:          "pending", // Default status
		TotalAmount:     0,         // Calculate total amount later
		Currency:        "",        // Default currency
		Items:           []models.OrderItem{},
		ShippingAddress: orderCreate.ShippingAddress,
		BillingAddress:  orderCreate.BillingAddress,
	}

	// Calculate total amount
	for _, itemCreate := range orderCreate.Items {
		// Find product price (replace with database lookup)
		for _, product := range products {
			if product.ProductID == itemCreate.ProductID {
				newOrder.TotalAmount += product.Price * float64(itemCreate.Quantity)
				newOrder.Items = append(newOrder.Items, models.OrderItem{
					ProductID: itemCreate.ProductID,
					Quantity:  itemCreate.Quantity,
					Price:     product.Price,
				})
				if newOrder.Currency == "" {
					newOrder.Currency = product.Currency
				}
				break
			}
		}
	}

	orders = append(orders, newOrder)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newOrder)
}

// GetOrder ...
func GetOrder(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	for _, o := range orders {
		if o.OrderID == orderID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(o)
			return
		}
	}

	http.Error(w, "Order not found", http.StatusNotFound)
}

// UpdateOrder ...
func UpdateOrder(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var orderUpdate models.OrderUpdate
	err = json.NewDecoder(r.Body).Decode(&orderUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i, o := range orders {
		if o.OrderID == orderID {
			if orderUpdate.Status != "" {
				orders[i].Status = orderUpdate.Status
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(orders[i])
			return
		}
	}

	http.Error(w, "Order not found", http.StatusNotFound)
}

// DeleteOrder ...
func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	for i, o := range orders {
		if o.OrderID == orderID {
			orders = append(orders[:i], orders[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Order not found", http.StatusNotFound)
}
