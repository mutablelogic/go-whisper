package httphandler

import (
	"bytes"
	"context"
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
// TESTS - TRANSLATE

func TestTranslate_MethodNotAllowed(t *testing.T) {
	router := http.NewServeMux()
	RegisterTranslateHandlers(router, "/api", testManager, noopMiddleware())

	req := httptest.NewRequest(http.MethodGet, "/api/translate", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rw.Code)
	}
}

func TestTranslate_MissingAudio(t *testing.T) {
	router := http.NewServeMux()
	RegisterTranslateHandlers(router, "/api", testManager, noopMiddleware())

	// Create multipart form without audio file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("model", "ggml-tiny.en")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/translate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", rw.Code, rw.Body.String())
	}
}

func TestTranslate_MissingModel(t *testing.T) {
	router := http.NewServeMux()
	RegisterTranslateHandlers(router, "/api", testManager, noopMiddleware())

	// Create multipart form with audio but no model
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add a dummy audio file
	part, _ := writer.CreateFormFile("audio", "test.wav")
	part.Write([]byte("dummy audio data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/translate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	// A missing model should return 404 Not Found
	if rw.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d: %s", rw.Code, rw.Body.String())
	}
}

func TestTranslate_ActualTranslation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping actual translation test in short mode")
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

	// Use a multilingual audio sample for translation
	samplePath := filepath.Join("..", "..", "samples", "multi-lang.wav")
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		// Fallback to German podcast if multi-lang not available
		samplePath = filepath.Join("..", "..", "samples", "de-podcast.wav")
		if _, err := os.Stat(samplePath); os.IsNotExist(err) {
			t.Skipf("No multilingual sample files found")
		}
	}

	// Read the audio file
	audioData, err := os.ReadFile(samplePath)
	if err != nil {
		t.Fatalf("failed to read audio file: %v", err)
	}

	// Setup router
	router := http.NewServeMux()
	RegisterTranslateHandlers(router, "/api", testManager, noopMiddleware())

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add model field
	_ = writer.WriteField("model", "ggml-tiny.en")

	// Add audio file
	part, err := writer.CreateFormFile("audio", filepath.Base(samplePath))
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(audioData); err != nil {
		t.Fatalf("failed to write audio data: %v", err)
	}
	writer.Close()

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/translate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	rw := httptest.NewRecorder()

	// Execute translation
	t.Log("Starting translation...")
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
		t.Error("expected non-empty translation text")
	}

	t.Logf("Translation result: %s", result.Text)

	// Translation should be in English
	lowerText := strings.ToLower(result.Text)
	// Just verify we got some English-looking text (has common English words)
	hasEnglish := containsAny(lowerText, "the", "a", "is", "are", "to", "in", "of", "and")
	if !hasEnglish {
		t.Logf("Warning: translation may not contain English text: %s", result.Text)
	}
}

func TestTranslate_WithPrompt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping actual translation test in short mode")
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

	// Use a German sample
	samplePath := filepath.Join("..", "..", "samples", "de-podcast.wav")
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
	RegisterTranslateHandlers(router, "/api", testManager, noopMiddleware())

	// Create multipart form with prompt
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("model", "ggml-tiny.en")
	_ = writer.WriteField("prompt", "German podcast")

	part, _ := writer.CreateFormFile("audio", "de-podcast.wav")
	part.Write(audioData)
	writer.Close()

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/translate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	rw := httptest.NewRecorder()

	// Execute translation
	t.Log("Starting translation with prompt...")
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
		t.Error("expected non-empty translation text")
	}

	t.Logf("Translation with prompt: %s", result.Text)
}
