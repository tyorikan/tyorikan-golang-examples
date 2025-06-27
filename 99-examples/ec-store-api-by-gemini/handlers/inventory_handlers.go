package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ec-store-api/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Mock data (replace with database)
var inventory = map[uuid.UUID]models.Inventory{}

// GetInventory ...
func GetInventory(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	inv, ok := inventory[productID]
	if !ok {
		http.Error(w, "Inventory not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}

// UpdateInventory ...
func UpdateInventory(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var inventoryUpdate models.InventoryUpdate
	err = json.NewDecoder(r.Body).Decode(&inventoryUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	inv, ok := inventory[productID]
	if !ok {
		http.Error(w, "Inventory not found", http.StatusNotFound)
		return
	}

	inv.StockQuantity = inventoryUpdate.StockQuantity
	inv.LastUpdated = time.Now()
	inventory[productID] = inv

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inv)
}
