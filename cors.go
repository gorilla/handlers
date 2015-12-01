package handlers

import "net/http"

type CORSHandler struct {
	h    http.Handler
	opts corsOptions
}

type corsOptions struct {
	allowedHeaders []string
	allowedMethods []string
	allowedOrigins []string
	maxAge         int
	ignoreOptions  bool
}

func (ch *CORSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
