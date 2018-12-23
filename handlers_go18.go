// +build go1.8

package handlers

import (
	"fmt"
	"net/http"
)

type loggingResponseWriter interface {
	commonLoggingResponseWriter
	http.Pusher
}

func (l *responseLogger) Push(target string, opts *http.PushOptions) error {
	p, ok := l.w.(http.Pusher)
	if !ok {
		return fmt.Errorf("responseLogger does not implement http.Pusher")
	}
	return p.Push(target, opts)
}

func (c *compressResponseWriter) Push(target string, opts *http.PushOptions) error {
	p, ok := c.ResponseWriter.(http.Pusher)
	if !ok {
		return fmt.Errorf("compressResponseWriter does not implement http.Pusher")
	}

	opts.Header.Add(xGorillaHeaderPush, "1") // make CompressHandler aware of Push request

	return p.Push(target, opts)
}
