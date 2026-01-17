package httphandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TESTS - TRANSCRIBE

func TestTranscribe_MethodNotAllowed(t *testing.T) {
	router := http.NewServeMux()
	RegisterTranscribeHandlers(router, "/api", testManager, noopMiddleware())

	req := httptest.NewRequest(http.MethodGet, "/api/transcribe", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rw.Code)
	}
}

func TestTranscribe_MissingAudio(t *testing.T) {
	router := http.NewServeMux()
	RegisterTranscribeHandlers(router, "/api", testManager, noopMiddleware())

	// Create multipart form without audio file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("model", "ggml-tiny.en")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/transcribe", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", rw.Code, rw.Body.String())
	}
}

func TestTranscribe_MissingModel(t *testing.T) {
	router := http.NewServeMux()
	RegisterTranscribeHandlers(router, "/api", testManager, noopMiddleware())

	// Create multipart form with audio but no model
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add a dummy audio file
	part, _ := writer.CreateFormFile("audio", "test.wav")
	part.Write([]byte("dummy audio data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/transcribe", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	// A missing model should return 404 Not Found
	if rw.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d: %s", rw.Code, rw.Body.String())
	}
}

func TestTranscribe_ActualTranscription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping actual transcription test in short mode")
	}

	// Download the model if not already present
	models := testManager.ListModels(context.Background())
	hasModel := false
	for _, m := range models {
		if m.Id == "ggml-tiny.en" {
			hasModel = true
			break
		}
	}

	if !hasModel {
		t.Log("Downloading ggml-tiny.en model...")
		if _, err := testManager.DownloadModel(context.Background(), "ggml-tiny.en.bin", nil); err != nil {
			t.Fatalf("failed to download model: %v", err)
		}
	}

	// Find an audio sample
	samplePath := filepath.Join("..", "..", "samples", "jfk.wav")
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", samplePath)
	}

	// Read the audio file
	audioData, err := os.ReadFile(samplePath)
	if err != nil {
		t.Fatalf("failed to read audio file: %v", err)
	}

	// Setup router
	router := http.NewServeMux()
	RegisterTranscribeHandlers(router, "/api", testManager, noopMiddleware())

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add model field
	_ = writer.WriteField("model", "ggml-tiny.en")

	// Add audio file
	part, err := writer.CreateFormFile("audio", "jfk.wav")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(audioData); err != nil {
		t.Fatalf("failed to write audio data: %v", err)
	}
	writer.Close()

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/transcribe", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rw := httptest.NewRecorder()

	// Execute transcription
	t.Log("Starting transcription...")
	router.ServeHTTP(rw, req)

	// Check response
	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}

	// Parse response
	var result schema.Transcription
	if err := parseJSON(rw.Body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify we got text
	if result.Text == "" {
		t.Error("expected non-empty transcription text")
	}

	t.Logf("Transcription result: %s", result.Text)

	// For JFK sample, we expect something about nation/country/ask
	lowerText := strings.ToLower(result.Text)
	if !containsAny(lowerText, "nation", "country", "ask", "what") {
		t.Logf("Warning: transcription may not be accurate: %s", result.Text)
	}
}

func TestTranscribe_WithPrompt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping actual transcription test in short mode")
	}

	// Ensure model is downloaded
	models := testManager.ListModels(context.Background())
	hasModel := false
	for _, m := range models {
		if m.Id == "ggml-tiny.en" {
			hasModel = true
			break
		}
	}

	if !hasModel {
		t.Log("Downloading ggml-tiny.en model...")
		if _, err := testManager.DownloadModel(context.Background(), "ggml-tiny.en.bin", nil); err != nil {
			t.Fatalf("failed to download model: %v", err)
		}
	}

	// Find an audio sample
	samplePath := filepath.Join("..", "..", "samples", "jfk.wav")
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", samplePath)
	}

	// Read the audio file
	audioData, err := os.ReadFile(samplePath)
	if err != nil {
		t.Fatalf("failed to read audio file: %v", err)
	}

	// Setup router
	router := http.NewServeMux()
	RegisterTranscribeHandlers(router, "/api", testManager, noopMiddleware())

	// Create multipart form with prompt
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("model", "ggml-tiny.en")
	_ = writer.WriteField("prompt", "President Kennedy")

	part, _ := writer.CreateFormFile("audio", "jfk.wav")
	part.Write(audioData)
	writer.Close()

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/transcribe", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rw := httptest.NewRecorder()

	// Execute transcription
	t.Log("Starting transcription with prompt...")
	router.ServeHTTP(rw, req)

	// Check response
	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}

	// Parse response
	var result schema.Transcription
	if err := parseJSON(rw.Body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Text == "" {
		t.Error("expected non-empty transcription text")
	}

	t.Logf("Transcription with prompt: %s", result.Text)
}

///////////////////////////////////////////////////////////////////////////////
// HELPER FUNCTIONS

func parseJSON(r io.Reader, v interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
