package mid

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

			v := web.GetValues(ctx)

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Info("request started", " trace_id ", v.TraceID, " method ", r.Method, " path ", path,
				" remoteaddr ", r.RemoteAddr)

			err := handler(ctx, w, r)

			log.Info("request completed", " trace_id ", v.TraceID, " method ", r.Method, " path ", path,
				" remoteaddr ", r.RemoteAddr, " statuscode ", v.StatusCode, " since ", time.Since(v.Now))

			return err
		}

		return h
	}
	return m
}
