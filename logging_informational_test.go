// Copyright 2026 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingInformationalResponses(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{"implicit status with body", 0, "response body"},
		{"implicit status without body", 0, ""},
		{"explicit status", http.StatusCreated, "response body"},
	} {
		t.Run(test.name, func(t *testing.T) {
			logged := make(chan LogFormatterParams, 1)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusEarlyHints)
				w.WriteHeader(http.StatusProcessing)
				if test.status != 0 {
					w.WriteHeader(test.status)
				}
				if test.body != "" {
					if _, err := io.WriteString(w, test.body); err != nil {
						t.Errorf("write body: %v", err)
					}
				}
			})
			server := httptest.NewServer(CustomLoggingHandler(io.Discard, handler, func(_ io.Writer, params LogFormatterParams) {
				logged <- params
			}))
			defer server.Close()

			response, err := server.Client().Get(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(body) != test.body {
				t.Errorf("response body = %q, want %q", body, test.body)
			}
			wantStatus := test.status
			if wantStatus == 0 {
				wantStatus = http.StatusOK
			}
			if response.StatusCode != wantStatus {
				t.Fatalf("response status = %d, want %d", response.StatusCode, wantStatus)
			}
			params := <-logged
			if params.StatusCode != response.StatusCode {
				t.Errorf("logged status = %d, response status = %d", params.StatusCode, response.StatusCode)
			}
			if params.Size != len(test.body) {
				t.Errorf("logged size = %d, want %d", params.Size, len(test.body))
			}
		})
	}
}

func TestLoggingSwitchingProtocols(t *testing.T) {
	recorder := httptest.NewRecorder()
	logger, writer := makeLogger(recorder)
	writer.WriteHeader(http.StatusSwitchingProtocols)
	if logger.Status() != recorder.Code {
		t.Errorf("logged status = %d, response status = %d", logger.Status(), recorder.Code)
	}
}
