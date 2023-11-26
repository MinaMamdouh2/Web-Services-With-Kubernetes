package summary

import "github.com/google/uuid"

// Summary represents information about an individual user and their products.
type Summary struct {
	UserID   uuid.UUID
	UserName string
	// Total number of products
	TotalCount int
	// Total cost of these products
	TotalCost float64
	// These 2 fields are aggregation, the only way to get
	// these aggergations is to query the products domain
}
