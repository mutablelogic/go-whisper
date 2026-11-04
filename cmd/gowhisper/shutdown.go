//go:build !client

package main

import (
	"context"
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// shutdownHandler returns middleware which cancels each request when the
// server's context is, so a transcription in progress is aborted when the
// server is stopped, and its client gets an error, rather than waiting until
// its connection is closed
func shutdownHandler(parent context.Context) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()
			stop := context.AfterFunc(parent, cancel)
			defer stop()
			next(w, r.WithContext(ctx))
		}
	}
}
