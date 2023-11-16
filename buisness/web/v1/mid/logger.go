package mid

import (
	"context"
	"net/http"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
	"go.uber.org/zap"
)

// Used clausers (creates a function that returns a function) to make the logger more flexible.
func Logger(log *zap.SugaredLogger) web.Middleware {

	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// We log the start and end of the request. so we can make sure that
			// the goroutine we created is completed, so when we find out that
			// a request has started but not completed there maybe a data race, leak or a block.

			log.Info(ctx, "request started", "method", r.Method, "path", r.URL.Path,
				"remoteaddr", r.RemoteAddr)

			err := handler(ctx, w, r)

			log.Info(ctx, "request completed", "method", r.Method, "path", r.URL.Path,
				"remoteaddr", r.RemoteAddr)

			return err
		}

		return h
	}
	return m
}
