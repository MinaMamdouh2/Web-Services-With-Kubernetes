// Package metrics constructs the metrics the application will track.
package metrics

import (
	"context"
	"expvar"
	"runtime"
)

// This holds the single instance of the metrics value needed for
// collecting metrics. The expvar package is already based on singleton
// for the different metrics that are registered with the package so there
// isn't much choice here
// Also we are not exporting the singleton instance directly thorugh the package
var m *metrics

// =========================================================================

// metrics represents the set of metrics we gather. These fields are
// safe to be accessed concurrently thanks to expvar. No extra abstraction is required.
type metrics struct {
	goroutines *expvar.Int
	requests   *expvar.Int
	errors     *expvar.Int
	panics     *expvar.Int
}

// init constructs the metrics value that will be used to capture metrics.
// The metrics value is stored in a package level variable since everything
// inside the expvar is registered as a singleton. The use of once will make
// sure this initialization is only done once.
func init() {
	m = &metrics{
		goroutines: expvar.NewInt("goroutines"),
		requests:   expvar.NewInt("requests"),
		errors:     expvar.NewInt("errors"),
		panics:     expvar.NewInt("panics"),
	}
}

// =========================================================================
// I am gonna use context since this is a web application
type ctxkey int

const key ctxkey = 1

// Set sets the context with the metrics value.
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

// AddGoroutines refreshes the goroutine metric every 100 requests.
func AddGoroutines(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		if v.requests.Value()%100 == 0 {
			v.goroutines.Set(int64(runtime.NumGoroutine()))
		}
	}
}

// AddRequests increments the request metrics by 1.
func AddRequests(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.requests.Add(1)
	}
}

// AddErrors increments the errors metrics by 1
func AddErrors(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.errors.Add(1)
	}
}

// AddPanics increments the panics metrics by 1
func AddPanics(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.panics.Add(1)
	}
}
