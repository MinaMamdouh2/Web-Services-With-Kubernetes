package web

import (
	"context"
	"time"
)

// Anytime you gotta store something in the go context, what you need is a key
// and that key needs to be of a unique type.
// The reason we wanna unique type is because we don't want to overide values
// we are stroing in the context thorough out the call chain.
type ctxKey int

// We use arbitrary number to represent the key. it could be any number
const key ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// SetValues sets the specified Values in the context.
func SetValues(ctx context.Context, v *Values) context.Context {
	return context.WithValue(ctx, key, v)
}

// GetValues returns the values from the context.
func GetValues(ctx context.Context) *Values {
	// This line attempts to retrieve a value associated with a specific key from the provided context (ctx)
	// The retrieved value (v) is asserted to be of type *Values.
	// The ok variable is a boolean flag indicating whether the assertion was successful.
	v, ok := ctx.Value(key).(*Values)
	// Check if the retrieval was successful.
	if !ok {
		// If not successful, return a default *Values instance.
		return &Values{
			TraceID: "00000000-0000-0000-0000-000000000000",
			Now:     time.Now(),
		}
	}

	return v
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}
	return v.TraceID
}

// GetTime returns the time from the context.
func GetTime(ctx context.Context) time.Time {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return time.Now()
	}
	return v.Now
}

// SetStatusCode sets the status code back into the context.
func SetStatusCode(ctx context.Context, statusCode int) {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return
	}

	v.StatusCode = statusCode
}
