//go:build go1.8
// +build go1.8

package handlers

import (
	"io/ioutil" //nolint:staticcheck //this test is for go1.8 hence deprecated api usage is allowed
	"net/http"
	"net/http/httptest"
	"testing"
)

// *httptest.ResponseRecorder doesn't implement Pusher, so wrap it.
type pushRecorder struct {
	*httptest.ResponseRecorder
}

func (pr pushRecorder) Push(_ string, _ *http.PushOptions) error {
	return nil
}

func TestLoggingHandlerWithPush(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := w.(http.Pusher); !ok {
			t.Fatalf("%T from LoggingHandler does not satisfy http.Pusher interface when built with Go >=1.8", w)
		}
		w.WriteHeader(http.StatusOK)
	})

	logger := LoggingHandler(ioutil.Discard, handler)
	logger.ServeHTTP(pushRecorder{httptest.NewRecorder()}, newRequest(http.MethodGet, "/"))
}

func TestCombinedLoggingHandlerWithPush(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := w.(http.Pusher); !ok {
			t.Fatalf("%T from CombinedLoggingHandler does not satisfy http.Pusher interface when built with Go >=1.8", w)
		}
		w.WriteHeader(http.StatusOK)
	})

	logger := CombinedLoggingHandler(ioutil.Discard, handler)
	logger.ServeHTTP(pushRecorder{httptest.NewRecorder()}, newRequest(http.MethodGet, "/"))
}
