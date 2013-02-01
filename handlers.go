// Copyright 2013 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package handlers is a collection of handlers for use with Go's net/http package.
*/
package handlers

import (
	"net/http"
	"sort"
	"strings"
)

// MethodHandler is a Handler that dispatches to a handler whose key in the MethodHandler's
// map matches the name of the HTTP request's method, eg: GET
//
// If the request's method is OPTIONS and OPTIONS is not a key in the map then the handler
// responds with a status of 200 and sets the Allow header to a comma-separated list of
// available methods.
//
// If the request's method doesn't match any of its keys the handler responds with
// a status of 406, Method not allowed and sets the Allow header to a comma-separated list
// of available methods.
type MethodHandler map[string]http.Handler

func (h MethodHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if handler, ok := h[req.Method]; ok {
		handler.ServeHTTP(w, req)
	} else {
		allow := []string{}
		for k := range h {
			allow = append(allow, k)
		}
		sort.Strings(allow)
		w.Header().Set("Allow", strings.Join(allow, ", "))
		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
