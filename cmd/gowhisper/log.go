//go:build !client

package main

import (
	"log/slog"
	"net/http"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// statusWriter records the status of a response, and flushes, and unwraps
// to, the response writer it wraps, so streamed responses, which check for
// an http.Flusher, are still flushed as they're written
type statusWriter struct {
	http.ResponseWriter
	status int
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (w *statusWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func (w *statusWriter) Flush() {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// logHandler returns middleware which logs each request, with its status,
// and how long it took, as an error when it's a server error
func logHandler(logger *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w}
			next(sw, r)
			if sw.status == 0 {
				sw.status = http.StatusOK
			}
			level := slog.LevelInfo
			if sw.status >= http.StatusInternalServerError {
				level = slog.LevelError
			}
			logger.Log(r.Context(), level, r.Method+" "+r.URL.Path, "status", sw.status, "duration", time.Since(start).Round(time.Millisecond).String())
		}
	}
}
