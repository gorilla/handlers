package handlers

import (
	"log"
	"net/http"
	"runtime"
)

const defaultRecoveryStackSize = 8192

type recoveryHandler struct {
	logger    *log.Logger
	handler   http.Handler
	stackSize int
}

// RecoveryHandler is HTTP middleware that recovers from a panic,
// logs the panic, writes http.StatusInternalServerError, and
// continues to the next handler.
//
// Example:
//
//  logger := log.New(os.Stdout, "prefix: ", log.LstdFlags)
//
//  r := mux.NewRouter()
//  r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//  	panic("Unexpected error!")
//  })
//
//  recoverRouter := handlers.RecoveryHandler(logger, r)
//  http.ListenAndServe(":1123", recoverRouter)
func RecoveryHandler(logger *log.Logger, h http.Handler) http.Handler {
	return recoveryHandler{logger, h, defaultRecoveryStackSize}
}

func (h recoveryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, h.stackSize)
			stack = stack[:runtime.Stack(stack, true)]

			h.logger.Printf("%s\n%s\n", err, stack)
		}
	}()

	h.handler.ServeHTTP(w, req)
}
