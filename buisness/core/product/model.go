package product

import (
	"time"

	"github.com/google/uuid"
)

// Product represents an individual product.
// A piece of wisdom :)
// Sold & Revenue field in the products will cause us pain
// because of the aggregation that can get very easily get out of sync in this table
// Would be better if we introduced a sales domain
type Product struct {
	ID       uuid.UUID
	Name     string
	Cost     float64
	Quantity int
	Sold     int
	Revenue  int
	// We are recording here what user in the system has entered this product
	// This a relationship a product has a user
	UserID      uuid.UUID
	DateCreated time.Time
	DateUpdated time.Time
}

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	Name     string
	Cost     float64
	Quantity int
	UserID   uuid.UUID
}
