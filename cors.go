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
// no additioanl headers are set on the response but the wrapped Handler is still called.
type CORSHandler struct {
	// Handler is called to handle all requests.
	Handler http.Handler

	// AllowOrigin is an optional function that returns true if the origin is one for which CORS requests are allowed.
	// If AllowOrigin is nil, all origins are allowed.
	AllowOrigin func(origin string) bool

	// AllowMethod is an optional function that returns true if the method is one for which CORS requests are allowed.
	// If AllowMethod is false, all methods are allowed.
	AllowMethod func(method string) bool

	// AllowedHeaders is an optional function will be used to check the Access-Control-Request-Headers
	// header from any preflight requests. It returns true if the headers are allowed.
	// If AllowHeaders is nil, no headers are allowed.
	AllowHeaders func(headers []string) bool

	// ExposeHeaders is an optional list of headers that is
	ExposeHeaders []string

	// If SupportCredentials is true, the Access-Control-Allow-Credentials header is set to the string 'true'.
	SupportsCredentials bool

	// If MaxAge is not 0 the Access-Control-Max-Age header is set to this value.
	MaxAge int64
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

func (h CORSHandler) allowOrigin(origin string) bool {
	return h.AllowOrigin == nil || h.AllowOrigin(origin)
}

func (h CORSHandler) allowMethod(method string) bool {
	return h.AllowMethod == nil || h.AllowMethod(method)
}

func (h CORSHandler) allowHeaders(headers []string) bool {
	return len(headers) == 0 || h.AllowHeaders != nil && h.AllowHeaders(headers)
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
		if req.Method == "OPTIONS" {
			handler = http.HandlerFunc(nilHandler)
			method := req.Header.Get("Access-Control-Request-Method")
			if method == "" || !h.allowMethod(method) {
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

			if h.MaxAge != 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.FormatInt(h.MaxAge, 10))
			}
			w.Header().Set("Access-Control-Allow-Methods", method)
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
		} else {
			if len(h.ExposeHeaders) != 0 {
				w.Header()["Access-Control-Expose-Headers"] = h.ExposeHeaders
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if h.SupportsCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}
}

// nilHandler is a handler that does nothing.
func nilHandler(w http.ResponseWriter, req *http.Request) {
	return
}
