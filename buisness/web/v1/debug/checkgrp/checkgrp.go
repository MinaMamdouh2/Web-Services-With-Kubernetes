// Package checkgrp maintains the group of handlers for health checking.
package checkgrp

import (
	"context"
	"net/http"
	"os"
	"time"

	database "github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/sys/database/pgx"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Handlers manages the set of check endpoints.
type Handlers struct {
	Build string
	Log   *zap.SugaredLogger
	DB    *sqlx.DB
}

// Readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
// Readiness basicly means that I am ready to recieve traffic, I am alive
// If you send me a request I should be able to process it no problem
// Part of us saying that we are ready is that we can talk to the DB
func (h Handlers) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	statusCode := http.StatusOK
	status := "ok"
	if err := database.StatusCheck(ctx, h.DB); err != nil {
		status = "db not ready"
		statusCode = http.StatusInternalServerError
	}
	data := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	if err := web.Respond(r.Context(), w, data, statusCode); err != nil {
		h.Log.Errorw("readiness", "ERROR", err)
	}

	h.Log.Infow("readiness", " statusCode ", statusCode, " method ", r.Method, " path ", r.URL.Path, " remoteaddr ", r.RemoteAddr)
}

// Liveness returns simple status info if the service is alive. If the
// app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
// Liveness is simply just a ping, are you alive? are you breathing?
// We also use the liveness handler as a way of getting also information about the service
func (h Handlers) Liveness(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}
	// Here we are defining this data structure that gives us status
	// build version, GOMAXPROCS
	data := struct {
		Status     string `json:"status,omitempty"`
		Build      string `json:"build,omitempty"`
		Host       string `json:"host,omitempty"`
		Name       string `json:"name,omitempty"`
		PodIP      string `json:"podIP,omitempty"`
		Node       string `json:"node,omitempty"`
		Namespace  string `json:"namespace,omitempty"`
		GOMAXPROCS string `json:"GOMAXPROCS,omitempty"`
	}{
		Status:     "up",
		Build:      h.Build,
		Host:       host,
		Name:       os.Getenv("KUBERNETES_NAME"),
		PodIP:      os.Getenv("KUBERNETES_POD_IP"),
		Node:       os.Getenv("KUBERNETES_NODE_NAME"),
		Namespace:  os.Getenv("KUBERNETES_NAMESPACE"),
		GOMAXPROCS: os.Getenv("GOMAXPROCS"),
	}

	if err := web.Respond(r.Context(), w, data, statusCode); err != nil {
		h.Log.Errorw("readiness", "ERROR", err)
	}

	// THIS IS A FREE TIMER. WE COULD UPDATE THE METRIC GOROUTINE COUNT HERE.
	h.Log.Infow("Liveness", " statusCode ", statusCode, " method ", r.Method, " path ", r.URL.Path, " remoteaddr ", r.RemoteAddr)

}
