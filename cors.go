package handlers

import "net/http"

// CORSOption represents a functional option for configuring the CORS middleware.
type CORSOption func(*cors) error

type cors struct {
	h              http.Handler
	allowedHeaders []string
	allowedMethods []string
	allowedOrigins []string
	maxAge         int
	ignoreOptions  bool
}

func (ch *cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ch.h.ServeHTTP(w, r)
}

// CORS provides Cross-Origin Resource Sharing middleware.
// Example:
//
//  import (
//      "net/http"
//
//      "github.com/gorilla/handlers"
//      "github.com/gorilla/mux"
//  )
//
//  func main() {
//      r := mux.NewRouter()
//      r.HandleFunc("/users", UserEndpoint)
//      r.HandleFunc("/projects", ProjectEndpoint)
//
//      // Apply the CORS middleware to our top-level router, with the defaults.
//      http.ListenAndServe(":8000", handlers.CORS()(r))
//  }
//
func CORS(opts ...CORSOption) func(http.Handler) http.Handler {
	ch := parseCORSOptions(opts...)

	// TODO(all): Set defaults
	// Note: append(allowedHeaders, defaultHeaders...) - the default headers here
	// should always be allowed:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS#Simple_requests

	return func(h http.Handler) http.Handler {
		ch.h = h
		return ch
	}
}

func parseCORSOptions(opts ...CORSOption) *cors {
	ch := &cors{}

	for _, option := range opts {
		option(ch)
	}

	return ch
}

//
// Functional options for configuring CORS.
//

// AllowedHeaders adds the provided headers to the list of allowed headers in a
// CORS request.
// The headers Content-Type, Expires, Cache-Control, ... are always allowed.
func AllowedHeaders(headers []string) CORSOption {
	return func(ch *cors) error {
		ch.allowedHeaders = headers
		return nil
	}
}

// AllowedMethods ...
func AllowedMethods(methods []string) CORSOption {
	return func(ch *cors) error {
		ch.allowedMethods = methods
		return nil
	}
}

// AllowedOrigins sets the allowed origins for CORS requests, as used in the
// 'Allow-Access-Control-Origin' HTTP header.
// Note: Passing in a []string{"*"} will allow any domain.
func AllowedOrigins(origins []string) CORSOption {
	return func(ch *cors) error {
		ch.allowedOrigins = origins
		return nil
	}
}

// MaxAge determines the maximum age (in seconds) between preflight requests. A
// maximum of 10 minutes is allowed. An age above this value will default to 10
// minutes.
func MaxAge(age int) CORSOption {
	return func(ch *cors) error {
		// Maximum of 10 minutes.
		if age > 600 {
			age = 600
		}

		ch.maxAge = age
		return nil
	}
}

// IgnoreOptions causes the CORS middleware to ignore OPTIONS requests, instead
// passing them through to the next handler. This is useful when your application
// or framework has a pre-existing mechanism for responding to OPTIONS requests.
func IgnoreOptions() CORSOption {
	return func(ch *cors) error {
		ch.ignoreOptions = true
		return nil
	}
}
