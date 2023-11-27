// This package is used to unmarshal any json it is coming in
package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dimfeld/httptreemux/v5"
)

type validator interface {
	Validate() error
}

// Param returns the web call parameters from the request.
func Param(r *http.Request, key string) string {
	m := httptreemux.ContextParams(r.Context())
	return m[key]
}

// Decode reads the body of an HTTP request looking for a JSON document. The
// body is decoded into the provided value.
// If the provided value is a struct then it is checked for validation tags.
// If the value implements a validate function, it is executed.
// This function job is to take any json we get in a post call and use the
// json package decoder function to unmarshal it and then call validation against it
// The validation is happening after the decoding
func Decode(r *http.Request, val any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	// There is a bug in the decode function in the standard library :)
	// When it fails it doesn't tell you why it failed this is for native types like "UUID"
	// that's why we are using scalar types "string"
	// Solutions to fix it:
	// 1) Fix it in the standard library
	// 2) Choose another json package
	if err := decoder.Decode(val); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	if v, ok := val.(validator); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("unable to validate payload: %w", err)
		}
	}

	return nil
}
