package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSHandler(t *testing.T) {
	// Test default configuration.
	r := newRequest("GET", "http://www.example.com/")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	CORS()(testHandler).ServeHTTP(rr, r)

	// TODO(all): Test this more heavily once the defaults are baked in.
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("bad status: got %v want %v", status, http.StatusFound)
	}
}
