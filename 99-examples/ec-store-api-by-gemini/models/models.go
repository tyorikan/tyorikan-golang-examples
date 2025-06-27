package models

import (
	"time"

	"github.com/google/uuid"
)

// Product ...
type Product struct {
	ProductID   uuid.UUID `json:"product_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	ImageURL    string    `json:"image_url"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductCreate ...
type ProductCreate struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
}

// ProductUpdate ...
type ProductUpdate struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
}

// Order ...
type Order struct {
	OrderID         uuid.UUID   `json:"order_id"`
	CustomerID      uuid.UUID   `json:"customer_id"`
	OrderDate       time.Time   `json:"order_date"`
	Status          string      `json:"status"`
	TotalAmount     float64     `json:"total_amount"`
	Currency        string      `json:"currency"`
	Items           []OrderItem `json:"items"`
	ShippingAddress Address     `json:"shipping_address"`
	BillingAddress  Address     `json:"billing_address"`
}

// OrderCreate ...
type OrderCreate struct {
	CustomerID      uuid.UUID         `json:"customer_id"`
	Items           []OrderItemCreate `json:"items"`
	ShippingAddress Address           `json:"shipping_address"`
	BillingAddress  Address           `json:"billing_address"`
}

// OrderUpdate ...
type OrderUpdate struct {
	Status string `json:"status"`
}

// OrderItem ...
type OrderItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
}

// OrderItemCreate ...
type OrderItemCreate struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// Customer ...
type Customer struct {
	CustomerID uuid.UUID `json:"customer_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CustomerCreate ...
type CustomerCreate struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// CustomerUpdate ...
type CustomerUpdate struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// Address ...
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

// Inventory ...
type Inventory struct {
	ProductID     uuid.UUID `json:"product_id"`
	StockQuantity int       `json:"stock_quantity"`
	LastUpdated   time.Time `json:"last_updated"`
}

// InventoryUpdate ...
type InventoryUpdate struct {
	StockQuantity int `json:"stock_quantity"`
}
