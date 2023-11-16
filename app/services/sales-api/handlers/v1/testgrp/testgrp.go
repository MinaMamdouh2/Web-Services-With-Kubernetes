package testgrp

import (
	"context"
	"net/http"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
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
	// This should be used for the OK responses only.
	return web.Respond(ctx, w, status, http.StatusOK)
}
