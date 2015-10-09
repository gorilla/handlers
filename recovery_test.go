package handlers

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoveryLogger(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		panic("Unexpected error!")
	})

	recovery := recoveryHandler{handler: handler, printTrace: false}
	recovery.ServeHTTP(httptest.NewRecorder(), newRequest("GET", "/subdir/asdf"))

	if !strings.Contains(buf.String(), "Unexpected error!") {
		t.Fatalf("Got log %#v, wanted substring %#v", buf.String(), "Unexpected error!")
	}
}
