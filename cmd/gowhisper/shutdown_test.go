//go:build !client

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// A request in progress is cancelled when the server's context is
func Test_ShutdownHandler(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	handler := shutdownHandler(parent)(func(w http.ResponseWriter, r *http.Request) {
		cancel()
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
			t.Error("request wasn't cancelled when the server's context was")
		}
	})
	handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/whisper/transcribe", nil))
}

// A request isn't cancelled while the server's context isn't
func Test_ShutdownHandler_Running(t *testing.T) {
	handler := shutdownHandler(context.Background())(func(w http.ResponseWriter, r *http.Request) {
		if err := r.Context().Err(); err != nil {
			t.Errorf("request was cancelled: %v", err)
		}
	})
	handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/whisper/model", nil))
}
