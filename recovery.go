package handlers

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"
)

type recoveryHandler struct {
	handler http.Handler
	options *RecoveryOptions
}

// RecoveryOptions provides configuration options for the
// reovery handler; such as setting the logging type and
// whether or not to print strack traces on panic.
type RecoveryOptions struct {
	Logger     *log.Logger
	PrintTrace bool
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
func RecoveryHandler(h http.Handler, options *RecoveryOptions) http.Handler {
	if options == nil {
		options = makeDefaultRecoveryOptions()
	}

	return recoveryHandler{h, options}
}

func makeDefaultRecoveryOptions() *RecoveryOptions {
	return &RecoveryOptions{
		Logger:     log.New(os.Stderr, "", log.LstdFlags),
		PrintTrace: true,
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
	if h.options.Logger != nil {
		h.options.Logger.Println(message)
	} else {
		log.Println(message)
	}

	if h.options.PrintTrace {
		debug.PrintStack()
	}
}
