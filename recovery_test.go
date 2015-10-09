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

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		panic("Unexpected error!")
	})

	logger := log.New(&buf, "", log.LstdFlags)

	recovery := RecoveryHandler(logger, handler)
	recovery.ServeHTTP(httptest.NewRecorder(), newRequest("GET", "/subdir/asdf"))

	if !strings.Contains(buf.String(), "Unexpected error!") {
		t.Fatalf("Got log %#v, wanted substring %#v", buf.String(), "Unexpected error!")
	}
}
