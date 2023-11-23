package mid

import (
	"context"
	"net/http"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/auth"
	v1 "github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/v1"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
	"go.uber.org/zap"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// Call the handler to see if an error occurred.
			// So we can process it
			if err := handler(ctx, w, r); err != nil {
				// First step in error handeling is to log the error.
				log.Errorw("ERROR", "trace_id", web.GetTraceID(ctx), "message", err)

				// We want to figure out what the response looks like
				// what the status looks like
				var er v1.ErrorResponse
				var status int
				// Inspect the error to reply accordingly
				switch {
				// Is this a trusted error?
				case v1.IsRequestError(err):
					reqErr := v1.GetRequestError(err)
					er = v1.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status
				// Is it an auth error
				case auth.IsAuthError(err):
					er = v1.ErrorResponse{
						Error: http.StatusText(http.StatusUnauthorized),
					}
					status = http.StatusUnauthorized
				// If it is not a trusted error
				default:
					er = v1.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				// If we get an error in the Respond, we will mark it as untrusted error
				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				if web.IsShutdown(err) {
					return err
				}
				// This should be the error handler middleware
				// Why are we returning the error?
				// Because we want to return the error to the base handler
				// to be able to shut down the service or make further processing

			}
			return nil
		}
		return h
	}

	return m
}
