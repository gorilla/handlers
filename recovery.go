package handlers

import (
	"log"
	"net/http"
	"runtime/debug"
)

type recoveryHandler struct {
	handler    http.Handler
	printTrace bool
}

// RecoveryHandler is HTTP middleware that recovers from a panic,
// logs the panic, writes http.StatusInternalServerError, and
// continues to the next handler.
//
// Example:
//
//  r := mux.NewRouter()
//  r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//  	panic("Unexpected error!")
//  })
//
//  recoverRouter := handlers.RecoveryHandler(r)
//  http.ListenAndServe(":1123", recoverRouter)
func RecoveryHandler(h http.Handler) http.Handler {
	return recoveryHandler{h, true}
}

func (h recoveryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)

			if h.printTrace {
				debug.PrintStack()
			}
		}
	}()

	h.handler.ServeHTTP(w, req)
}
