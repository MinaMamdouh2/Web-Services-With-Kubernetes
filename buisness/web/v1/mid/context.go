package mid

import (
	"context"
	"net/http"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/metrics"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
)

// Metrics updates program counters.
func Metrics() web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// Anytime this gets called, we put the metrics data into the context
			ctx = metrics.Set(ctx)

			// We call the handler
			err := handler(ctx, w, r)

			metrics.AddRequests(ctx)
			metrics.AddGoroutines(ctx)

			if err != nil {
				metrics.AddErrors(ctx)
			}

			return err
		}

		return h
	}

	return m
}
