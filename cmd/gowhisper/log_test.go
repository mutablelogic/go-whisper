//go:build !client

package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A streamed response is flushed through the logging middleware, and the
// request is logged with its status
func Test_LogHandler_Flush(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	handler := logHandler(logger)(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer isn't an http.Flusher")
		}
		w.Write([]byte("event"))
		flusher.Flush()
	})

	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodPost, "/api/whisper/transcribe", nil))
	if !w.Flushed {
		t.Error("response wasn't flushed")
	}
	if !strings.Contains(logs.String(), `msg="POST /api/whisper/transcribe" status=200`) {
		t.Errorf("request wasn't logged: %q", logs.String())
	}
}

// A server error is logged as an error
func Test_LogHandler_Error(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	handler := logHandler(logger)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/whisper/model", nil))
	if !strings.Contains(logs.String(), `level=ERROR msg="GET /api/whisper/model" status=500`) {
		t.Errorf("server error wasn't logged as an error: %q", logs.String())
	}
}
