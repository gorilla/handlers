package handlers

import (
	"log"
	"net/http"
	"runtime/debug"
)

type recoveryHandler struct {
	handler    http.Handler
	logger     *log.Logger
	printStack bool
}

// Option provides a functional approach to define
// configuration for a handler; such as setting the logging
// whether or not to print strack traces on panic.
type Option func(http.Handler)

func parseOptions(h http.Handler, opts ...Option) http.Handler {
	for _, option := range opts {
		option(h)
	}

	return h
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
func RecoveryHandler(opts ...Option) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		r := &recoveryHandler{handler: h}
		return parseOptions(r, opts...)
	}
}

func RecoveryLogger(logger *log.Logger) Option {
	return func(h http.Handler) {
		r := h.(*recoveryHandler)
		r.logger = logger
	}
}

func PrintRecoveryStack(print bool) Option {
	return func(h http.Handler) {
		r := h.(*recoveryHandler)
		r.printStack = print
	}
}

func (h recoveryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log(err)
		}
	}()

	h.handler.ServeHTTP(w, req)
}

func (h recoveryHandler) log(message interface{}) {
	if h.logger != nil {
		h.logger.Println(message)
	} else {
		log.Println(message)
	}

	if h.printStack {
		debug.PrintStack()
	}
}
