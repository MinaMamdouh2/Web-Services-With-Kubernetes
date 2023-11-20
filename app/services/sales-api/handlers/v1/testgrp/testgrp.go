package testgrp

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
)

// Test is our example route
func Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		return errors.New("untrusted error")
	}
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
