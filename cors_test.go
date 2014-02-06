package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var handlerFunc = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello\n"))
})

func TestMatchHeaders(t *testing.T) {
	matcher := MatchHeaders("Origin", "Accept", "X-Requested-With")

	tests := []struct {
		Headers []string
		Allowed bool
	}{
		{[]string{"accept", "origin", "x-requested-with"}, true},
		{[]string{"Accept", "Origin", "X-Requested-With"}, true},
		{[]string{"Accept", "Origin"}, true},
		{[]string{"Accept", "Origin", "Pizza"}, false},
	}

	for i, test := range tests {
		if allowed := matcher(test.Headers); allowed != test.Allowed {
			t.Logf("%d: %v got %v, want %v", i, test.Headers, allowed, test.Allowed)
			t.Fail()
		}
	}
}

func TestCORS(t *testing.T) {
	basicHandler := CORSHandler{Handler: handlerFunc}
	maxAgeHandler := CORSHandler{Handler: handlerFunc, MaxAge: 3600}
	headersHandler := CORSHandler{Handler: handlerFunc, AllowHeaders: MatchHeaders("Origin", "Allow", "X-Requested-With")}
	credentialsHandler := CORSHandler{Handler: handlerFunc, SupportsCredentials: true}
	originsHandler := CORSHandler{Handler: handlerFunc, AllowOrigin: func(o string) bool {
		return o == "http://foo" || o == "http://bar"
	}}
	methodsHandler := CORSHandler{Handler: handlerFunc, AllowMethod: func(m string) bool {
		return m == "PUT" || m == "DELETE"
	}}

	tests := []struct {
		Handler http.Handler

		// Request Headers
		Method                      string
		Origin                      string
		AccessControlRequestMethod  string
		AccessControlRequestHeaders string

		// Response Headers
		AccessControlMaxAge           string
		AccessControlAllowMethods     string
		AccessControlAllowHeaders     string
		AccessControlExposeHeaders    string
		AccessControlAllowOrigin      string
		AccessControlAllowCredentials string
	}{
		{basicHandler, "GET", "", "", "", "", "", "", "", "", ""},
		{basicHandler, "POST", "", "", "", "", "", "", "", "", ""},
		{basicHandler, "OPTIONS", "", "", "", "", "", "", "", "", ""},
		{basicHandler, "GET", "http://www.example.com", "", "", "", "", "", "", "http://www.example.com", ""},
		{basicHandler, "POST", "http://www.example.com", "", "", "", "", "", "", "http://www.example.com", ""},
		{basicHandler, "OPTIONS", "http://www.example.com", "POST", "", "", "POST", "", "", "http://www.example.com", ""},
		{basicHandler, "OPTIONS", "http://www.example.com", "POST", "Some-Header", "", "", "", "", "", ""},
		{maxAgeHandler, "OPTIONS", "http://www.example.com", "POST", "", "3600", "POST", "", "", "http://www.example.com", ""},
		{headersHandler, "OPTIONS", "http://www.example.com", "POST", "allow, origin, x-requested-with", "", "POST", "allow, origin, x-requested-with", "", "http://www.example.com", ""},
		{headersHandler, "OPTIONS", "http://www.example.com", "POST", "Origin", "", "POST", "Origin", "", "http://www.example.com", ""},
		{headersHandler, "OPTIONS", "http://www.example.com", "POST", "Bar", "", "", "", "", "", ""},
		{credentialsHandler, "OPTIONS", "http://www.example.com", "POST", "", "", "POST", "", "", "http://www.example.com", "true"},
		{originsHandler, "OPTIONS", "http://www.example.com", "POST", "", "", "", "", "", "", ""},
		{originsHandler, "OPTIONS", "http://foo", "POST", "", "", "POST", "", "", "http://foo", ""},
		{originsHandler, "OPTIONS", "http://bar", "POST", "", "", "POST", "", "", "http://bar", ""},
		{methodsHandler, "OPTIONS", "http://www.example.com", "GET", "", "", "", "", "", "", ""},
		{methodsHandler, "OPTIONS", "http://www.example.com", "POST", "", "", "", "", "", "", ""},
		{methodsHandler, "OPTIONS", "http://www.example.com", "PUT", "", "", "PUT", "", "", "http://www.example.com", ""},
		{methodsHandler, "OPTIONS", "http://www.example.com", "DELETE", "", "", "DELETE", "", "", "http://www.example.com", ""},
	}

	for i, test := range tests {
		rec := httptest.NewRecorder()
		req := newRequest(test.Method, "http://example.com")
		if test.Origin != "" {
			req.Header.Set("Origin", test.Origin)
		}
		if test.AccessControlRequestMethod != "" {
			req.Header.Set("Access-Control-Request-Method", test.AccessControlRequestMethod)
		}
		if test.AccessControlRequestHeaders != "" {
			req.Header.Set("Access-Control-Request-Headers", test.AccessControlRequestHeaders)
		}

		test.Handler.ServeHTTP(rec, req)

		expectBody := "hello\n"
		if body := rec.Body.String(); test.Method != "OPTIONS" && body != expectBody {
			t.Logf("%d: wrong body: got %q want %q", i, body, expectBody)
			t.Fail()
		}
		if val := rec.HeaderMap.Get("Access-Control-Max-Age"); val != test.AccessControlMaxAge {
			t.Logf("%d: wrong value for Access-Control-Max-Age: got %q want %q", i, val, test.AccessControlMaxAge)
			t.Fail()
		}
		if val := rec.HeaderMap.Get("Access-Control-Allow-Methods"); val != test.AccessControlAllowMethods {
			t.Logf("%d: wrong value for Access-Control-Allow-Methods: got %q want %q", i, val, test.AccessControlAllowMethods)
			t.Fail()
		}
		if val := rec.HeaderMap.Get("Access-Control-Allow-Headers"); val != test.AccessControlAllowHeaders {
			t.Logf("%d: wrong value for Access-Control-Allow-Headers: got %q want %q", i, val, test.AccessControlAllowHeaders)
			t.Fail()
		}
		if val := rec.HeaderMap.Get("Access-Control-Expose-Headers"); val != test.AccessControlExposeHeaders {
			t.Logf("%d: wrong value for Access-Control-Expose-Headers: got %q want %q", i, val, test.AccessControlExposeHeaders)
			t.Fail()
		}
		if val := rec.HeaderMap.Get("Access-Control-Allow-Origin"); val != test.AccessControlAllowOrigin {
			t.Logf("%d: wrong value for Access-Control-Allow-Origin: got %q want %q", i, val, test.AccessControlAllowOrigin)
			t.Fail()
		}
		if val := rec.HeaderMap.Get("Access-Control-Allow-Credentials"); val != test.AccessControlAllowCredentials {
			t.Logf("%d: wrong value for Access-Control-Allow-Credentials: got %q want %q", i, val, test.AccessControlAllowCredentials)
			t.Fail()
		}
	}
}
