// Copyright 2013 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"sync"
)

// compressableTypes are the HTTP Content-Types that will be compressed by
// compressHandler in a response.
var compressableTypes = []string{
	"text/html",
	"text/richtext",
	"text/plain",
	"text/css",
	"text/x-script",
	"text/x-component",
	"text/x-java-source",
	"application/javascript",
	"application/x-javascript",
	"text/javascript",
	"text/js",
	"image/x-icon",
	"application/x-perl",
	"application/x-httpd-cgi",
	"text/xml",
	"application/xml",
	"application/xml+rss",
	"application/json",
	"multipart/bag",
	"multipart/mixed",
	"application/xhtml+xml",
	"font/ttf",
	"font/otf",
	"font/x-woff",
	"image/svg+xml",
	"application/vnd.ms-fontobject",
	"application/ttf",
	"application/x-ttf",
	"application/otf",
	"application/x-otf",
	"application/truetype",
	"application/opentype",
	"application/x-opentype",
	"application/woff",
	"application/eot",
	"application/font",
	"application/font-woff",
	"application/font-sfnt",
}

// gzipPool is a pool of gzip writers using gzip.DefaultCompression.
// Note: Due to the inability to change the level after initialization, levels
// other than flate.DefaultCompression do not use pooled writers.
var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	}}

// compressHandler is a http.Handler that performs gzip/flate compression for
// HTTP responses.
type compressHandler struct {
	h     http.Handler
	level int
}

type compressResponseWriter struct {
	io.Writer
	http.ResponseWriter
	http.Hijacker
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
	ctype := h.Get("Content-Type")
	if ctype == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}

	// If the MIME type of the response is not compressable, use the underlying
	// (unmodified) writer instead.
	if !mimeExists(ctype, compressableTypes) {
		h.Del("Content-Encoding")
		return w.ResponseWriter.Write(b)
	}

	h.Del("Content-Length")
	w.Header().Add("Vary", "Accept-Encoding")

	return w.Writer.Write(b)
}

// CompressHandler gzip compresses HTTP responses for clients that support it
// via the 'Accept-Encoding' header.
//
func CompressHandler(h http.Handler) http.Handler {
	return &compressHandler{h, gzip.DefaultCompression}
}

// CompressHandlerLevel gzip compresses HTTP responses with specified compression level
// for clients that support it via the 'Accept-Encoding' header.
//
// The compression level should be gzip.DefaultCompression, gzip.NoCompression,
// or any integer value between gzip.BestSpeed and gzip.BestCompression inclusive.
// gzip.DefaultCompression is used in case of invalid compression level.
//
// Note that most users should use CompressHandler (which uses the default
// level), or be conscious of the increased client-side CPU requirements that
// higher compression levels have. The default compression level also uses a
// writer pool for gzip responses to reduce allocations across requests.
func CompressHandlerLevel(h http.Handler, level int) http.Handler {
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	return &compressHandler{h, level}
}

// ServeHTTP satisfies the http.Handler interface for compressHandler.
func (ch *compressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
L:
	for _, enc := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		switch strings.TrimSpace(enc) {
		case "gzip":
			w.Header().Set("Content-Encoding", "gzip")

			var gw *gzip.Writer
			switch ch.level {
			// Use a pooled writer where possible
			case gzip.DefaultCompression:
				gw = gzipPool.Get().(*gzip.Writer)
				// Explicitly reset the Writer before use.
				gw.Reset(w)
				// Only return a writer of the same level.
				defer gzipPool.Put(gw)
			default:
				gw, _ = gzip.NewWriterLevel(w, ch.level)
			}

			defer gw.Close()

			h, hok := w.(http.Hijacker)
			if !hok { /* w is not Hijacker... oh well... */
				h = nil
			}

			w = &compressResponseWriter{
				Writer:         gw,
				ResponseWriter: w,
				Hijacker:       h,
			}

			break L
		case "deflate":
			w.Header().Set("Content-Encoding", "deflate")

			fw, _ := flate.NewWriter(w, ch.level)
			defer fw.Close()

			h, hok := w.(http.Hijacker)
			if !hok { /* w is not Hijacker... oh well... */
				h = nil
			}

			w = &compressResponseWriter{
				Writer:         fw,
				ResponseWriter: w,
				Hijacker:       h,
			}

			break L
		}
	}

	// Call the wrapped handler
	ch.h.ServeHTTP(w, r)
}

// mimeExists checks whether the provided MIME type is in a list of MIME types.
func mimeExists(m string, accepted []string) bool {
	for _, ctype := range accepted {
		mtype, _, err := mime.ParseMediaType(m)
		if err != nil {
			log.Println(err)
			return false
		}

		// MIME type == compressable type
		if mtype == ctype {
			return true
		}

	}

	return false
}
