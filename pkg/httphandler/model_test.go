package httphandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	// Packages
	pkg "github.com/mutablelogic/go-whisper/pkg"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var testManager *pkg.Manager

///////////////////////////////////////////////////////////////////////////////
// TEST MAIN

func TestMain(m *testing.M) {
	// Create temporary directory for models
	tmpDir, err := os.MkdirTemp("", "whisper-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manager once for all tests
	testManager, err = pkg.New(tmpDir)
	if err != nil {
		panic(err)
	}
	defer testManager.Close()

	// Run tests
	os.Exit(m.Run())
}

///////////////////////////////////////////////////////////////////////////////
// HELPER FUNCTIONS

// noopMiddleware returns middleware that does nothing
func noopMiddleware() HTTPMiddlewareFuncs {
	return HTTPMiddlewareFuncs{}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - LIST MODELS

func TestModelList_Success(t *testing.T) {
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	req := httptest.NewRequest(http.MethodGet, "/api/model", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	var models []*schema.Model
	if err := json.NewDecoder(rw.Body).Decode(&models); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Models may be empty if whisper hasn't registered any yet
	if len(models) == 0 {
		t.Skip("no models registered - test environment issue")
	}
}

func TestModelList_MethodNotAllowed(t *testing.T) {
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	req := httptest.NewRequest(http.MethodPut, "/api/model", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rw.Code)
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - GET MODEL

func TestModelGet_Success(t *testing.T) {
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	// Get a known model ID
	models := testManager.ListModels(context.Background())
	if len(models) == 0 {
		t.Skip("no models available for testing")
	}
	modelID := models[0].Id

	req := httptest.NewRequest(http.MethodGet, "/api/model/"+modelID, nil)
	req.SetPathValue("id", modelID)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}

	var model schema.Model
	if err := json.NewDecoder(rw.Body).Decode(&model); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if model.Id != modelID {
		t.Errorf("expected model ID %s, got %s", modelID, model.Id)
	}
}

func TestModelGet_NotFound(t *testing.T) {
	// Use global testManager
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	req := httptest.NewRequest(http.MethodGet, "/api/model/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d: %s", rw.Code, rw.Body.String())
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - DOWNLOAD MODEL

func TestModelDownload_BadRequest_EmptyBody(t *testing.T) {
	// Use global testManager
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	req := httptest.NewRequest(http.MethodPost, "/api/model", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rw.Code)
	}
}

func TestModelDownload_BadRequest_MissingModel(t *testing.T) {
	// Use global testManager
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	reqBody := schema.DownloadModelRequest{
		Model: "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/model", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", rw.Code, rw.Body.String())
	}
}

func TestModelDownload_NotFound(t *testing.T) {
	// Use global testManager
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	reqBody := schema.DownloadModelRequest{
		Model: "nonexistent-model",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/model", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	// Nonexistent model returns 400 because the model name is invalid
	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", rw.Code, rw.Body.String())
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - DELETE MODEL

func TestModelDelete_NotFound(t *testing.T) {
	// Use global testManager
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	req := httptest.NewRequest(http.MethodDelete, "/api/model/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d: %s", rw.Code, rw.Body.String())
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - ERROR MAPPING

func TestHttperr_Nil(t *testing.T) {
	err := httperr(nil)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestHttperr_ReturnsError(t *testing.T) {
	err := httperr(ErrNotFound)
	if err == nil {
		t.Fatal("expected error to be returned")
	}

	err = httperr(ErrBadParameter.With("test"))
	if err == nil {
		t.Fatal("expected error to be returned")
	}

	err = httperr(errors.New("generic"))
	if err == nil {
		t.Fatal("expected error to be returned")
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - MIDDLEWARE

func TestMiddleware_Wrap(t *testing.T) {
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	middleware := HTTPMiddlewareFuncs{
		func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test", "value")
				next(w, r)
			}
		},
	}

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	wrapped(rw, req)

	if !called {
		t.Error("expected handler to be called")
	}

	if rw.Header().Get("X-Test") != "value" {
		t.Error("expected middleware to set header")
	}
}

func TestMiddleware_WrapEmpty(t *testing.T) {
	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
	}

	middleware := HTTPMiddlewareFuncs{}
	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	wrapped(rw, req)

	if !called {
		t.Error("expected handler to be called")
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - STREAMING

func TestModelDownload_Streaming(t *testing.T) {
	// Use global testManager
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	reqBody := schema.DownloadModelRequest{
		Model: "nonexistent-model",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/model", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	// Check that response is streaming
	if contentType := rw.Header().Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %s", contentType)
	}

	// For a non-existent model, we should get an error event
	body = rw.Body.Bytes()
	if len(body) == 0 {
		t.Error("expected response body")
	}

	// Should contain an error event
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "event: error") && !strings.Contains(bodyStr, "error") {
		t.Logf("Response: %s", bodyStr)
		t.Error("expected error event in streaming response")
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS - MODEL DOWNLOAD (ACTUAL DOWNLOAD)

func TestModelDownload_ActualDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping actual download test in short mode")
	}

	// Use the correct model name from huggingface: ggml-tiny.en.bin
	targetModel := "ggml-tiny.en.bin"

	t.Logf("Testing download of model: %s", targetModel)

	// Setup HTTP handler
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	// Create download request
	reqBody := schema.DownloadModelRequest{
		Model: targetModel,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/model", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()

	// Execute download
	t.Log("Starting model download...")
	router.ServeHTTP(rw, req)

	// Check response
	if rw.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rw.Code, rw.Body.String())
	}

	var result schema.Model
	if err := json.NewDecoder(rw.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Model downloaded successfully: %s", result.Id)

	// Verify model is now available
	downloadedModel, err := testManager.GetModel(context.Background(), result.Id)
	if err != nil {
		t.Fatalf("failed to get downloaded model: %v", err)
	}

	if downloadedModel.Id != result.Id {
		t.Errorf("expected model ID %s, got %s", result.Id, downloadedModel.Id)
	}

	t.Logf("Verified model is available: %s", downloadedModel.Id)
}

func TestModelDownload_ActualDownload_Streaming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping actual download test in short mode")
	}

	// Use the correct model name from huggingface: ggml-tiny.en.bin
	targetModel := "ggml-tiny.en.bin"

	t.Logf("Testing streaming download of model: %s", targetModel)

	// Setup HTTP handler
	router := http.NewServeMux()
	RegisterModelHandlers(router, "/api", testManager, noopMiddleware())

	// Create download request with streaming
	reqBody := schema.DownloadModelRequest{
		Model: targetModel,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/model", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	rw := httptest.NewRecorder()

	// Execute download
	t.Log("Starting streaming model download...")
	router.ServeHTTP(rw, req)

	// Check response is streaming
	if contentType := rw.Header().Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %s", contentType)
	}

	// Parse streaming response
	response := rw.Body.String()
	if len(response) == 0 {
		t.Error("expected streaming response body")
	}

	t.Logf("Received streaming response with %d bytes", len(response))

	// Should contain progress or done events
	if !containsAny(response, "event: progress", "event: done", "progress", "done") {
		t.Errorf("expected progress or done events in response, got: %s", response)
	}

	t.Log("Streaming download completed successfully")
}

// Helper function to check if string contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
