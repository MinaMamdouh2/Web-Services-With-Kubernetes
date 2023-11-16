package testgrp

import (
	"context"
	"encoding/json"
	"net/http"
)

// Test is our example route
func Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Vlidate the data
	// Call into theBusiness layer
	// Return errors
	//Handle OK responses
	status := struct {
		Status string
	}{
		Status: "OK",
	}
	return json.NewEncoder(w).Encode(status)
}
