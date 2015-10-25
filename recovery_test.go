package handlers

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoveryLoggerWithDefaultOptionsUsingType(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		panic("Unexpected error!")
	})

	recovery := recoveryHandler{handler: handler, options: &RecoveryOptions{}}
	recovery.ServeHTTP(httptest.NewRecorder(), newRequest("GET", "/subdir/asdf"))

	if !strings.Contains(buf.String(), "Unexpected error!") {
		t.Fatalf("Got log %#v, wanted substring %#v", buf.String(), "Unexpected error!")
	}
}

func TestRecoveryLoggerWithDefaultOptionsUsingApi(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		panic("Unexpected error!")
	})

	recovery := RecoveryHandler(handler, &RecoveryOptions{Logger: nil, PrintTrace: false})
	recovery.ServeHTTP(httptest.NewRecorder(), newRequest("GET", "/subdir/asdf"))

	if !strings.Contains(buf.String(), "Unexpected error!") {
		t.Fatalf("Got log %#v, wanted substring %#v", buf.String(), "Unexpected error!")
	}
}

func TestRecoveryLoggerWithCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	var logger = log.New(&buf, "", log.LstdFlags)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		panic("Unexpected error!")
	})

	recovery := RecoveryHandler(handler, &RecoveryOptions{Logger: logger, PrintTrace: false})
	recovery.ServeHTTP(httptest.NewRecorder(), newRequest("GET", "/subdir/asdf"))

	if !strings.Contains(buf.String(), "Unexpected error!") {
		t.Fatalf("Got log %#v, wanted substring %#v", buf.String(), "Unexpected error!")
	}
}
