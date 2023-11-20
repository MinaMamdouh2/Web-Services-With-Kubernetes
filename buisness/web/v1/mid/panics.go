package mid

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics() web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			// Defer us used as injection between the handler and the error middleware
			// we use a named return variable so that defer can modify it
			defer func() {
				if rec := recover(); rec != nil {
					// This gives us the stack trace of the goroutine which panicked.
					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace))
				}
			}()

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
