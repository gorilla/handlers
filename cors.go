package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// CORSHandler adds the headers required for Cross-Origin Resource Sharing.
//
// If the requests do not meet the criteria at http://www.w3.org/TR/cors,
// no additional headers are set on the response but the wrapped Handler is still called.
type CORSHandler struct {
	// Handler is called to handle all requests.
	Handler http.Handler

	// AllowOrigin is an optional function that returns true if the origin is one for which CORS requests are allowed.
	// If AllowOrigin is nil, all origins are allowed.
	AllowOrigin func(origin string) bool

	// AllowMethod is an optional function that returns true if the method is one for which CORS requests are allowed.
	// If AllowMethod is nil, all methods are allowed.
	AllowMethod func(method string) bool

	// AllowedHeaders is an optional function will be used to check the Access-Control-Request-Headers
	// header from any preflight requests. It returns true if the headers are allowed.
	// If AllowHeaders is nil, no headers are allowed.
	AllowHeaders func(headers []string) bool

	// ExposeHeaders is an optional list of headers that is accessible to the browser.
	ExposeHeaders func(r *http.Request) []string

	// SupportCredentials sets the value of the Access-Control-Allow-Credentials header.
	// It is set to the string 'true' if the returned value is true.
	// If SupportCredentials is nil, the header is not set.
	//
	// Note this is invalid to have both this value to true and a wilcard (*) Origin.
	SupportsCredentials func(r *http.Request) bool

	// MaxAge defines the time a preflight request can be cached.
	// If MaxAge is nil, not header is set.
	// If MaxAge returns a value >0, then the Access-Control-Max-Age header is set to this value.
	MaxAge func(r *http.Request) int64
}

func (h CORSHandler) allowOrigin(origin string) bool {
	return h.AllowOrigin == nil || h.AllowOrigin(origin)
}

func (h CORSHandler) allowMethod(method string) bool {
	if method == "" {
		return false
	}
	return h.AllowMethod == nil || h.AllowMethod(method)
}

func (h CORSHandler) allowHeaders(headers []string) bool {
	return len(headers) == 0 || h.AllowHeaders != nil && h.AllowHeaders(headers)
}

func (h CORSHandler) exposeHeaders(r *http.Request) []string {
	if h.ExposeHeaders == nil {
		return nil
	}
	return h.ExposeHeaders(r)
}

func (h CORSHandler) supportsCredentials(r *http.Request) bool {
	if h.SupportsCredentials == nil {
		return false
	}
	return h.SupportsCredentials(r)
}

func (h CORSHandler) maxAge(r *http.Request) int64 {
	if h.MaxAge == nil {
		return 0
	}
	return h.MaxAge(r)
}

func (h CORSHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := h.Handler
	defer func() {
		handler.ServeHTTP(w, req)
	}()

	if origin := req.Header.Get("Origin"); origin != "" {
		if !h.allowOrigin(origin) {
			return
		}
		// Delete existing Access-Control-Allow-Origin
		w.Header().Del("Access-Control-Allow-Origin")
		if req.Method == "OPTIONS" {
			handler = http.HandlerFunc(nilHandler)
			method := req.Header.Get("Access-Control-Request-Method")
			if !h.allowMethod(method) {
				return
			}

			var allowedHeaders []string
			for _, headerFieldNames := range req.Header["Access-Control-Request-Headers"] {
				headers := strings.Split(headerFieldNames, ",")
				for i := range headers {
					headers[i] = strings.TrimSpace(headers[i])
				}
				if !h.allowHeaders(headers) {
					return
				}
				allowedHeaders = append(allowedHeaders, headers...)
			}

			if maxAge := h.maxAge(req); maxAge != 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.FormatInt(maxAge, 10))
			}
			w.Header().Set("Access-Control-Allow-Methods", method)
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
		} else {
			if headers := h.exposeHeaders(req); len(headers) != 0 {
				w.Header()["Access-Control-Expose-Headers"] = headers
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if h.supportsCredentials(req) {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}
}

// MatchHeaders returns a function that can be used to match a list of header keys against those
// provided by the headers arguments.
func MatchHeaders(headers ...string) func([]string) bool {
	allowed := make([]string, len(headers))
	for i, h := range headers {
		allowed[i] = http.CanonicalHeaderKey(h)
	}
	sort.Strings(allowed)

	return func(reqHeaders []string) bool {
		for _, h := range reqHeaders {
			ch := http.CanonicalHeaderKey(h)
			i := sort.SearchStrings(allowed, ch)
			if i >= len(allowed) || allowed[i] != ch {
				return false
			}
		}
		return true
	}
}

// nilHandler is a handler that does nothing.
func nilHandler(w http.ResponseWriter, req *http.Request) {
	return
}
