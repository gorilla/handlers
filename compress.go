// Copyright 2013 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const acceptEncoding string = "Accept-Encoding"

type compressResponseWriter struct {
	io.Writer
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier
}

func (w *compressResponseWriter) WriteHeader(c int) {
	w.ResponseWriter.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(c)
}

func (w *compressResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *compressResponseWriter) Write(b []byte) (int, error) {
	h := w.ResponseWriter.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}
	h.Del("Content-Length")

	return w.Writer.Write(b)
}

type flusher interface {
	Flush() error
}

func (w *compressResponseWriter) Flush() {
	// Flush compressed data if compressor supports it.
	if f, ok := w.Writer.(flusher); ok {
		f.Flush()
	}
	// Flush HTTP response.
	if w.Flusher != nil {
		w.Flusher.Flush()
	}
}

// CompressHandler gzip compresses HTTP responses for clients that support it
// via the 'Accept-Encoding' header.
//
// Compressing TLS traffic may leak the page contents to an attacker if the
// page contains user input: http://security.stackexchange.com/a/102015/12208
func CompressHandler(h http.Handler) http.Handler {
	return CompressHandlerLevel(h, gzip.DefaultCompression)
}

// CompressHandlerLevel gzip compresses HTTP responses with specified compression level
// for clients that support it via the 'Accept-Encoding' header.
//
// The compression level should be gzip.DefaultCompression, gzip.NoCompression,
// or any integer value between gzip.BestSpeed and gzip.BestCompression inclusive.
// gzip.DefaultCompression is used in case of invalid compression level.
func CompressHandlerLevel(h http.Handler, level int) http.Handler {
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	const (
		gzipEncoding  = "gzip"
		flateEncoding = "deflate"
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// detect what encoding to use
		var encoding string
		for _, curEnc := range strings.Split(r.Header.Get(acceptEncoding), ",") {
			curEnc = strings.TrimSpace(curEnc)
			if curEnc == gzipEncoding || curEnc == flateEncoding {
				encoding = curEnc
				break
			}
		}

		// always add Accept-Encoding to Vary to prevent intermediate caches corruption
		w.Header().Add("Vary", acceptEncoding)

		// if we weren't able to identify an encoding we're familiar with, pass on the
		// request to the handler and return
		if encoding == "" {
			h.ServeHTTP(w, r)
			return
		}

		// wrap the ResponseWriter with the writer for the chosen encoding
		var encWriter io.WriteCloser
		if encoding == gzipEncoding {
			encWriter, _ = gzip.NewWriterLevel(w, level)
		} else if encoding == flateEncoding {
			encWriter, _ = flate.NewWriter(w, level)
		}
		defer encWriter.Close()

		w.Header().Set("Content-Encoding", encoding)
		r.Header.Del(acceptEncoding)

		hijacker, ok := w.(http.Hijacker)
		if !ok { /* w is not Hijacker... oh well... */
			hijacker = nil
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			flusher = nil
		}

		closeNotifier, ok := w.(http.CloseNotifier)
		if !ok {
			closeNotifier = nil
		}

		w = &compressResponseWriter{
			Writer:         encWriter,
			ResponseWriter: w,
			Hijacker:       hijacker,
			Flusher:        flusher,
			CloseNotifier:  closeNotifier,
		}

		h.ServeHTTP(w, r)
	})
}
