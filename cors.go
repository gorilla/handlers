package handlers

import (
	// "log"
	"net/http"
	"strconv"
	"strings"
)

// CORSOption represents a functional option for configuring the CORS middleware.
type CORSOption func(*cors) error

type cors struct {
	h                http.Handler
	allowedHeaders   []string
	allowedMethods   []string
	allowedOrigins   []string
	exposedHeaders   []string
	maxAge           int
	ignoreOptions    bool
	allowCredentials bool
}

var (
	defaultCorsMethods = []string{"GET", "HEAD", "POST"}
	defaultCorsHeaders = []string{"Accept", "Accept-Language", "Content-Language"}
)

const (
	corsOptionMethod           string = "OPTIONS"
	corsAllowOriginHeader      string = "Access-Control-Allow-Origin"
	corsExposeHeadersHeader    string = "Access-Control-Expose-Headers"
	corsMaxAgeHeader           string = "Access-Control-Max-Age"
	corsAllowMethodsHeader     string = "Access-Control-Allow-Methods"
	corsAllowHeadersHeader     string = "Access-Control-Allow-Headers"
	corsAllowCredentialsHeader string = "Access-Control-Allow-Credentials"
	corsRequestMethodHeader    string = "Access-Control-Request-Method"
	corsRequestHeadersHeader   string = "Access-Control-Request-Headers"
	corsOriginHeader           string = "Origin"
	corsVaryHeader             string = "Vary"
	corsOriginMatchAll         string = "*"
)

func (ch *cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get(corsOriginHeader)

	if !ch.isOriginAllowed(origin) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler := ch.h
	defer func() {
		handler.ServeHTTP(w, r)
	}()

	if r.Method == corsOptionMethod {
		if ch.ignoreOptions {
			return
		}

		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { return })
		if _, ok := r.Header[corsRequestMethodHeader]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		method := r.Header.Get(corsRequestMethodHeader)
		if !ch.isMatch(method, ch.allowedMethods) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		requestHeaders := strings.Split(r.Header.Get(corsRequestHeadersHeader), ",")
		allowedHeaders := []string{}
		for _, v := range requestHeaders {
			canonicalHeader := http.CanonicalHeaderKey(strings.TrimSpace(v))
			if canonicalHeader == "" || ch.isMatch(canonicalHeader, defaultCorsHeaders) {
				continue
			}

			if !ch.isMatch(canonicalHeader, ch.allowedHeaders) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			allowedHeaders = append(allowedHeaders, canonicalHeader)
		}

		if len(allowedHeaders) > 0 {
			w.Header().Set(corsAllowHeadersHeader, strings.Join(allowedHeaders, ","))
		}

		if ch.maxAge > 0 {
			w.Header().Set(corsMaxAgeHeader, strconv.Itoa(ch.maxAge))
		}

		if !ch.isMatch(method, defaultCorsMethods) {
			w.Header().Set(corsAllowMethodsHeader, method)
		}
	} else {
		if len(ch.exposedHeaders) > 0 {
			w.Header().Set(corsExposeHeadersHeader, strings.Join(ch.exposedHeaders, ","))
		}
	}

	if ch.allowCredentials {
		w.Header().Set(corsAllowCredentialsHeader, "true")
	}

	if len(ch.allowedOrigins) > 1 {
		w.Header().Set(corsVaryHeader, corsOriginHeader)
	}

	w.Header().Set(corsAllowOriginHeader, origin)
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
	ch := &cors{
		allowedMethods: defaultCorsMethods,
		allowedHeaders: defaultCorsHeaders,
		allowedOrigins: []string{corsOriginMatchAll},
	}

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
		for _, v := range headers {
			normalizedHeader := http.CanonicalHeaderKey(strings.TrimSpace(v))
			if normalizedHeader == "" {
				continue
			}

			if !ch.isMatch(normalizedHeader, ch.allowedHeaders) {
				ch.allowedHeaders = append(ch.allowedHeaders, normalizedHeader)
			}
		}

		return nil
	}
}

// AllowedMethods ...
func AllowedMethods(methods []string) CORSOption {
	return func(ch *cors) error {
		for _, v := range methods {
			normalizedMethod := strings.ToUpper(strings.TrimSpace(v))
			if normalizedMethod == "" {
				continue
			}

			if !ch.isMatch(normalizedMethod, ch.allowedMethods) {
				ch.allowedHeaders = append(ch.allowedHeaders, normalizedMethod)
			}
		}

		return nil
	}
}

// AllowedOrigins sets the allowed origins for CORS requests, as used in the
// 'Allow-Access-Control-Origin' HTTP header.
// Note: Passing in a []string{"*"} will allow any domain.
func AllowedOrigins(origins []string) CORSOption {
	return func(ch *cors) error {
		for _, v := range origins {
			if v == corsOriginMatchAll {
				ch.allowedOrigins = []string{corsOriginMatchAll}
				return nil
			}
		}

		ch.allowedOrigins = origins
		return nil
	}
}

// ExposeHeaders are additional headers outside of those which are apart
// of the simple response headers (http://www.w3.org/TR/cors/#simple-response-header)
func ExposedHeaders(headers []string) CORSOption {
	return func(ch *cors) error {
		ch.exposedHeaders = headers
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

// AllowCredentials ...
func AllowCredentials() CORSOption {
	return func(ch *cors) error {
		ch.allowCredentials = true
		return nil
	}
}

func (ch *cors) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range ch.allowedOrigins {
		if allowedOrigin == origin || allowedOrigin == corsOriginMatchAll {
			return true
		}
	}

	return false
}

func (ch *cors) isMatch(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}
