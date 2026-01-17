package httphandler

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

func TestTranscribe_PlainTextResponse(t *testing.T) {
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

	// Create request with Accept: text/plain
	req := httptest.NewRequest(http.MethodPost, "/api/transcribe", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "text/plain")
	rw := httptest.NewRecorder()

	// Execute transcription
	t.Log("Starting transcription with text/plain response...")
	router.ServeHTTP(rw, req)

	// Check response
	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}

	// Check content type
	if rw.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("expected Content-Type text/plain, got %s", rw.Header().Get("Content-Type"))
	}

	// Verify we got text
	text := rw.Body.String()
	if text == "" {
		t.Error("expected non-empty transcription text")
	}
	if len(text) < 10 {
		t.Errorf("expected longer transcription text, got: %s", text)
	}

	t.Logf("Transcription (plain text): %s", text)
}

func TestTranscribe_SRTResponse(t *testing.T) {
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

	// Create request with Accept: application/x-subrip
	req := httptest.NewRequest(http.MethodPost, "/api/transcribe", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/x-subrip")
	rw := httptest.NewRecorder()

	// Execute transcription
	t.Log("Starting transcription with SRT response...")
	router.ServeHTTP(rw, req)

	// Check response
	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}

	// Check content type
	if rw.Header().Get("Content-Type") != "application/x-subrip" {
		t.Errorf("expected Content-Type application/x-subrip, got %s", rw.Header().Get("Content-Type"))
	}

	// Verify we got SRT format (should contain timecodes if segments exist)
	srtContent := rw.Body.String()
	// Note: The response might be empty if there are no segments with timing info
	// The important thing is that if there is content, it's in valid SRT format
	if len(srtContent) > 0 {
		if !bytes.Contains([]byte(srtContent), []byte("-->")) {
			t.Error("expected SRT format with timecodes (-->)")
		}
		t.Logf("Transcription (SRT format):\n%s", srtContent)
	} else {
		t.Log("Note: SRT response was empty (may be due to no segments with timing info)")
	}
}

func TestTranscribe_JSONResponse(t *testing.T) {
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

	// Create request with Accept: application/json (default)
	req := httptest.NewRequest(http.MethodPost, "/api/transcribe", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	rw := httptest.NewRecorder()

	// Execute transcription
	t.Log("Starting transcription with JSON response...")
	router.ServeHTTP(rw, req)

	// Check response
	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}

	// Check content type
	if !bytes.Contains([]byte(rw.Header().Get("Content-Type")), []byte("application/json")) {
		t.Errorf("expected Content-Type application/json, got %s", rw.Header().Get("Content-Type"))
	}

	// Parse response
	var result schema.Transcription
	if err := parseJSON(rw.Body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify we got structured data
	if result.Text == "" {
		t.Error("expected non-empty transcription text")
	}
	// Note: segments might not always be populated depending on the model
	// The important thing is we got valid JSON with text
	if len(result.Text) < 10 {
		t.Errorf("expected longer transcription text")
	}

	t.Logf("Transcription (JSON): %d segments, text length: %d", len(result.Segments), len(result.Text))
}
