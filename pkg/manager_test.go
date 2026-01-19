package pkg_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	// Packages
	pkg "github.com/mutablelogic/go-whisper/pkg"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
)

///////////////////////////////////////////////////////////////////////////////
// TESTS

func TestManager_New(t *testing.T) {
	// Create temporary directory for models
	tmpDir := t.TempDir()

	// Create manager without any API keys
	manager, err := pkg.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Verify ListModels works and returns a slice
	models := manager.ListModels(context.Background())
	if models == nil {
		t.Error("expected non-nil models slice from ListModels()")
	}
}

func TestManager_NewWithElevenLabsKey(t *testing.T) {
	// Create temporary directory for models
	tmpDir := t.TempDir()

	// Create manager with elevenlabs API key
	manager, err := pkg.New(tmpDir, pkg.OptElevenLabsKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Verify ListModels includes both whisper and elevenlabs models
	models := manager.ListModels(context.Background())
	hasElevenLabs := false
	for _, model := range models {
		if model.OwnedBy == "elevenlabs" {
			hasElevenLabs = true
			break
		}
	}
	if !hasElevenLabs {
		t.Error("expected elevenlabs models to be present")
	}
}

func TestManager_NewWithOpenAIKey(t *testing.T) {
	// Create temporary directory for models
	tmpDir := t.TempDir()

	// Create manager with openai API key
	manager, err := pkg.New(tmpDir, pkg.OptOpenAIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Verify ListModels includes both whisper and openai models
	models := manager.ListModels(context.Background())
	hasOpenAI := false
	for _, model := range models {
		if model.OwnedBy == "openai" {
			hasOpenAI = true
			break
		}
	}
	if !hasOpenAI {
		t.Error("expected openai models to be present")
	}
}

func TestManager_NewWithAllKeys(t *testing.T) {
	// Create temporary directory for models
	tmpDir := t.TempDir()

	// Create manager with both API keys
	manager, err := pkg.New(tmpDir,
		pkg.OptElevenLabsKey("elevenlabs-key"),
		pkg.OptOpenAIKey("openai-key"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Verify ListModels includes models from all three sources
	models := manager.ListModels(context.Background())
	hasElevenLabs := false
	hasOpenAI := false
	for _, model := range models {
		if model.OwnedBy == "elevenlabs" {
			hasElevenLabs = true
		}
		if model.OwnedBy == "openai" {
			hasOpenAI = true
		}
	}
	if !hasElevenLabs {
		t.Error("expected elevenlabs models to be present")
	}
	if !hasOpenAI {
		t.Error("expected openai models to be present")
	}
}

func TestManager_NewWithWhisperOpts(t *testing.T) {
	// Create temporary directory for models
	tmpDir := t.TempDir()

	// Create manager with whisper options
	manager, err := pkg.New(tmpDir,
		pkg.WithWhisperOpt(
			whisper.OptMaxConcurrent(2),
			whisper.OptNoGPU(),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Verify manager is created
	if manager == nil {
		t.Error("expected manager to be initialized")
	}
}

func TestManager_Close(t *testing.T) {
	// Create temporary directory for models
	tmpDir := t.TempDir()

	// Create manager
	manager, err := pkg.New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Close should not error
	if err := manager.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Multiple closes should not panic
	if err := manager.Close(); err != nil {
		t.Errorf("second Close() returned error: %v", err)
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS FOR MANAGER METHODS

///////////////////////////////////////////////////////////////////////////////
// TESTS FOR OPENAI ROUTES

func TestManager_Transcribe_OpenAI_StreamUnsupported(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptOpenAIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Transcribe with no model ID should fail
	req := &schema.TranscribeRequest{}
	_, err = manager.Transcribe(context.Background(), nil, bytes.NewReader([]byte{}), req)
}

func TestManager_Transcribe_OpenAI_DiarizeUnsupported(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptOpenAIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Diarize should be rejected by OpenAI
	diarizeTrue := true
	req := &schema.TranscribeRequest{
		TranslateRequest: schema.TranslateRequest{
			Model:   "whisper-1",
			Diarize: &diarizeTrue,
		},
	}
	_, err = manager.Transcribe(context.Background(), nil, bytes.NewReader([]byte{}), req)
	if err == nil {
		t.Error("expected error for diarize parameter in OpenAI transcription")
	}
}

func TestManager_Translate_OpenAI_DiarizeUnsupported(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptOpenAIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Diarize should be rejected by OpenAI
	diarizeTrue := true
	req := &schema.TranslateRequest{
		Model:   "whisper-1",
		Diarize: &diarizeTrue,
	}
	_, err = manager.Translate(context.Background(), nil, bytes.NewReader([]byte{}), req)
	if err == nil {
		t.Error("expected error for diarize parameter in OpenAI translation")
	}
}

func TestManager_Transcribe_OpenAI_ModelNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptOpenAIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Non-existent model should return error
	req := &schema.TranscribeRequest{
		TranslateRequest: schema.TranslateRequest{
			Model: "nonexistent-model",
		},
	}
	_, err = manager.Transcribe(context.Background(), nil, bytes.NewReader([]byte{}), req)
	if err == nil {
		t.Error("expected error for nonexistent model")
	}
}

func TestManager_Translate_OpenAI_ModelNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptOpenAIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Non-existent model should return error
	req := &schema.TranslateRequest{
		Model: "nonexistent-model",
	}
	_, err = manager.Translate(context.Background(), nil, bytes.NewReader([]byte{}), req)
	if err == nil {
		t.Error("expected error for nonexistent model")
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS FOR ELEVENLABS ROUTES

func TestManager_Transcribe_ElevenLabs_StreamUnsupported(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptElevenLabsKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Stream should be rejected by ElevenLabs
	req := &schema.TranscribeRequest{
		TranslateRequest: schema.TranslateRequest{
			Model: "scribe_v2",
		},
	}
	_, err = manager.Transcribe(context.Background(), nil, bytes.NewReader([]byte{}), req)
	if err == nil {
		t.Error("expected error for stream parameter in ElevenLabs transcription")
	}
}

func TestManager_Transcribe_ElevenLabs_PromptUnsupported(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptElevenLabsKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Prompt should be rejected by ElevenLabs
	promptText := "test prompt"
	req := &schema.TranscribeRequest{
		TranslateRequest: schema.TranslateRequest{
			Model:  "scribe_v2",
			Prompt: &promptText,
		},
	}
	_, err = manager.Transcribe(context.Background(), nil, bytes.NewReader([]byte{}), req)
	if err == nil {
		t.Error("expected error for prompt parameter in ElevenLabs transcription")
	}
}

func TestManager_Translate_ElevenLabs_NotSupported(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptElevenLabsKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// ElevenLabs doesn't support translation at all
	req := &schema.TranslateRequest{
		Model: "scribe_v2",
	}
	_, err = manager.Translate(context.Background(), nil, bytes.NewReader([]byte{}), req)
	if err == nil {
		t.Error("expected error for translation with ElevenLabs")
	}
}

func TestManager_Transcribe_ElevenLabs_ModelNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptElevenLabsKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Non-existent model should return error
	req := &schema.TranscribeRequest{
		TranslateRequest: schema.TranslateRequest{
			Model: "nonexistent-model",
		},
	}
	_, err = manager.Transcribe(context.Background(), nil, bytes.NewReader([]byte{}), req)
	if err == nil {
		t.Error("expected error for nonexistent model")
	}
}

///////////////////////////////////////////////////////////////////////////////
// INTEGRATION TESTS - OPENAI
// These tests make actual API calls to OpenAI and require OPENAI_API_KEY to be set
// Run with: OPENAI_API_KEY=your-key go test -run Integration ./pkg

func TestManager_Transcribe_OpenAI_Integration(t *testing.T) {
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptOpenAIKey(openaiKey))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Open sample audio file - use .mp3 format which OpenAI supports
	file, err := os.Open("../samples/en-office.mp3")
	if err != nil {
		t.Fatalf("failed to open sample file: %v", err)
	}
	defer file.Close()

	// Transcribe with OpenAI
	filename := "en-office.mp3"
	req := &schema.TranscribeRequest{
		TranslateRequest: schema.TranslateRequest{
			Model:    "whisper-1",
			Filename: &filename,
		},
	}
	result, err := manager.Transcribe(context.Background(), nil, file, req)
	if err != nil {
		t.Fatalf("transcription failed: %v", err)
	}

	// Verify result has text
	if result == nil {
		t.Error("expected non-nil result")
	} else if result.Text == "" {
		t.Error("expected non-empty transcription text")
	} else {
		t.Logf("Transcription: %s\n", result.Text)
	}
}

func TestManager_Translate_OpenAI_Integration(t *testing.T) {
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptOpenAIKey(openaiKey))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Open sample audio file in non-English language - use .mp3 format which OpenAI reliably accepts
	file, err := os.Open("../samples/OlivierL.wav")
	if err != nil {
		t.Fatalf("failed to open sample file: %v", err)
	}
	defer file.Close()

	// Translate with OpenAI
	filename := "OlivierL.wav"
	req := &schema.TranslateRequest{
		Model:    "whisper-1",
		Filename: &filename,
	}
	result, err := manager.Translate(context.Background(), nil, file, req)
	if err != nil {
		t.Fatalf("translation failed: %v", err)
	}

	// Verify result has English text
	if result == nil {
		t.Error("expected non-nil result")
	} else if result.Text == "" {
		t.Error("expected non-empty translated text")
	} else {
		t.Logf("Translation: %s\n", result.Text)
	}
}

///////////////////////////////////////////////////////////////////////////////
// INTEGRATION TESTS - ELEVENLABS
// These tests make actual API calls to ElevenLabs and require ELEVENLABS_API_KEY to be set
// Run with: ELEVENLABS_API_KEY=your-key go test -run Integration ./pkg

func TestManager_Transcribe_ElevenLabs_Integration(t *testing.T) {
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		t.Skip("ELEVENLABS_API_KEY not set, skipping integration test")
	}

	tmpDir := t.TempDir()
	manager, err := pkg.New(tmpDir, pkg.OptElevenLabsKey(elevenLabsKey))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	// Open sample audio file
	file, err := os.Open("../samples/jfk.wav")
	if err != nil {
		t.Fatalf("failed to open sample file: %v", err)
	}
	defer file.Close()

	// Transcribe with ElevenLabs
	req := &schema.TranscribeRequest{
		TranslateRequest: schema.TranslateRequest{
			Model: "scribe_v2",
		},
	}
	result, err := manager.Transcribe(context.Background(), nil, file, req)
	if err != nil {
		t.Fatalf("transcription failed: %v", err)
	}

	// Verify result has text
	if result == nil {
		t.Error("expected non-nil result")
	} else if result.Text == "" {
		t.Error("expected non-empty transcription text")
	} else {
		t.Logf("Transcription: %s\n", result.Text)
	}
}
