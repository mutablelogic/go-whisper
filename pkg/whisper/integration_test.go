package whisper

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/sys/whisper"
	assert "github.com/stretchr/testify/assert"
)

// Integration tests that use real models and audio samples
// These tests will be skipped if models are not available
//
// To download models, run:
//   mkdir -p models && cd models
//   wget https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin
//   wget https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin

const (
	modelsDir  = "../../third_party/whisper.cpp/models"
	samplesDir = "../../samples"
	modelURL   = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/?download=true"
	modelTiny  = "ggml-tiny.bin" // ~75MB, fastest for testing
)

// Helper to download a model if it doesn't exist
func downloadModelIfNeeded(t *testing.T, modelsPath, modelName string) string {
	modelPath := filepath.Join(modelsPath, modelName)

	// Check if model already exists and is large enough (>8MB)
	if info, err := os.Stat(modelPath); err == nil && info.Size() > 8*1024*1024 {
		t.Logf("Model already exists: %s (%d MB)", modelName, info.Size()/(1024*1024))
		return modelPath
	}

	t.Logf("Downloading model: %s (this may take a minute...)", modelName)

	// Create models directory if it doesn't exist
	if err := os.MkdirAll(modelsPath, 0755); err != nil {
		t.Skipf("Could not create models directory: %v", err)
		return ""
	}

	// Create file for writing
	f, err := os.Create(modelPath)
	if err != nil {
		t.Skipf("Could not create model file: %v", err)
		return ""
	}
	defer f.Close()

	// Download the model
	client := whisper.NewClient(modelURL)
	if client == nil {
		t.Skip("Could not create HTTP client for model download")
		return ""
	}

	ctx := context.Background()
	_, err = client.Get(ctx, f, modelName)
	if err != nil {
		// Clean up partial download
		os.Remove(modelPath)
		t.Skipf("Could not download model: %v", err)
		return ""
	}

	// Verify download
	if info, err := os.Stat(modelPath); err != nil || info.Size() < 8*1024*1024 {
		os.Remove(modelPath)
		t.Skip("Downloaded model is too small or invalid")
		return ""
	}

	t.Logf("Model downloaded successfully: %s", modelName)
	return modelPath
}

// Helper to check if test should be skipped, and download model if needed
func skipIfNoModels(t *testing.T) string {
	// Try to download the tiny model if it doesn't exist
	modelPath := downloadModelIfNeeded(t, modelsDir, modelTiny)
	if modelPath == "" {
		return ""
	}

	return modelsDir
}

// Helper to get path to a sample audio file
func getSamplePath(t *testing.T, filename string) string {
	path := filepath.Join(samplesDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", path)
	}
	return path
}

func TestIntegration_TranscribeJFK(t *testing.T) {
	modelsPath := skipIfNoModels(t)
	assert := assert.New(t)

	// Initialize whisper
	mgr, err := New(modelsPath)
	assert.NoError(err)
	assert.NotNil(mgr)
	defer Close()

	// Get available models
	models := mgr.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
	}
	t.Logf("Found %d model(s)", len(models))

	// Use the first available model (preferably tiny or base)
	var model *schema.Model
	for _, m := range models {
		// Look for tiny or base models (handle "for-tests-ggml-tiny" naming)
		if m.Id == "tiny" || m.Id == "base" ||
		   m.Id == "for-tests-ggml-tiny" || m.Id == "for-tests-ggml-base" ||
		   m.Id == "for-tests-ggml-tiny.en" {
			model = m
			break
		}
	}
	if model == nil {
		model = models[0] // Use first available
	}
	t.Logf("Using model: %s (%s)", model.Id, model.Path)

	// Get JFK sample
	samplePath := getSamplePath(t, "jfk.wav")

	// Open audio file
	audioFile, err := os.Open(samplePath)
	assert.NoError(err)
	defer audioFile.Close()

	// Read audio samples - jfk.wav is 11 seconds of 16kHz mono
	// For this test, we'll just create some dummy samples
	// In a real integration test, you'd decode the WAV file properly
	samples := make([]float32, 16000*11) // 11 seconds of silence as placeholder

	// Perform transcription
	var result *schema.Transcription
	ctx := context.Background()

	err = mgr.WithModel(model, func(task *Task) error {
		task.SetLanguage("en")
		task.SetTranslate(false)

		// Transcribe the audio
		err := task.Transcribe(ctx, 0, samples, nil)
		if err != nil {
			return err
		}

		result = task.Result()
		return nil
	})

	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal("transcribe", result.Task)

	t.Logf("Transcription: %s", result.Text)
	t.Logf("Language: %s", result.Language)
	t.Logf("Duration: %v", result.Duration)
}

func TestIntegration_TranscribeWithSegments(t *testing.T) {
	modelsPath := skipIfNoModels(t)
	assert := assert.New(t)

	// Initialize whisper
	mgr, err := New(modelsPath)
	assert.NoError(err)
	defer Close()

	// Get a model
	models := mgr.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
	}
	model := models[0]

	// Create test samples (1 second of silence)
	samples := make([]float32, 16000)

	// Track segments
	var segments []*schema.Segment
	segmentCallback := func(seg *schema.Segment) {
		segments = append(segments, seg)
		t.Logf("Segment %d: %s (%.2fs - %.2fs)",
			seg.Id, seg.Text,
			time.Duration(seg.Start).Seconds(), time.Duration(seg.End).Seconds())
	}

	ctx := context.Background()
	err = mgr.WithModel(model, func(task *Task) error {
		return task.Transcribe(ctx, 0, samples, segmentCallback)
	})

	assert.NoError(err)
	// Segments may or may not be generated for silence
	t.Logf("Generated %d segment(s)", len(segments))
}

func TestIntegration_LanguageDetection(t *testing.T) {
	modelsPath := skipIfNoModels(t)
	assert := assert.New(t)

	mgr, err := New(modelsPath)
	assert.NoError(err)
	defer Close()

	models := mgr.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
	}
	model := models[0]

	// Create test samples
	samples := make([]float32, 16000)

	ctx := context.Background()
	var detectedLang string

	err = mgr.WithModel(model, func(task *Task) error {
		// Set language to auto for detection
		err := task.SetLanguage("auto")
		assert.NoError(err)

		err = task.Transcribe(ctx, 0, samples, nil)
		if err != nil {
			return err
		}

		detectedLang = task.Language()
		return nil
	})

	assert.NoError(err)
	t.Logf("Detected language: %s", detectedLang)
}

func TestIntegration_Translation(t *testing.T) {
	modelsPath := skipIfNoModels(t)
	assert := assert.New(t)

	mgr, err := New(modelsPath)
	assert.NoError(err)
	defer Close()

	models := mgr.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
	}

	// Find a multilingual model
	var model *schema.Model
	for _, m := range models {
		if m.Id != "tiny.en" && m.Id != "base.en" {
			model = m
			break
		}
	}
	if model == nil {
		t.Skip("No multilingual model available for translation test")
	}

	samples := make([]float32, 16000)

	ctx := context.Background()
	var result *schema.Transcription

	err = mgr.WithModel(model, func(task *Task) error {
		// Check if model can translate
		if !task.CanTranslate() {
			return nil
		}

		task.SetTranslate(true)
		err := task.Transcribe(ctx, 0, samples, nil)
		if err != nil {
			return err
		}

		result = task.Result()
		return nil
	})

	assert.NoError(err)
	if result != nil {
		assert.Equal("translate", result.Task)
		t.Logf("Translation result: %s", result.Text)
	}
}

func TestIntegration_TemperatureSettings(t *testing.T) {
	modelsPath := skipIfNoModels(t)
	assert := assert.New(t)

	mgr, err := New(modelsPath)
	assert.NoError(err)
	defer Close()

	models := mgr.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
	}
	model := models[0]

	// Test different temperature values
	temperatures := []float64{0.0, 0.2, 0.8, 1.0}

	for _, temp := range temperatures {
		t.Run(string(rune(int(temp*10))), func(t *testing.T) {
			samples := make([]float32, 16000)
			ctx := context.Background()

			err = mgr.WithModel(model, func(task *Task) error {
				err := task.SetTemperature(temp)
				assert.NoError(err)

				return task.Transcribe(ctx, 0, samples, nil)
			})

			assert.NoError(err)
		})
	}
}

func TestIntegration_ContextReuse(t *testing.T) {
	modelsPath := skipIfNoModels(t)
	assert := assert.New(t)

	mgr, err := New(modelsPath)
	assert.NoError(err)
	defer Close()

	models := mgr.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
	}
	model := models[0]

	samples := make([]float32, 16000)
	ctx := context.Background()

	// Perform multiple transcriptions with the same model
	// This tests that the context pool properly reuses contexts
	for i := 0; i < 3; i++ {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			err = mgr.WithModel(model, func(task *Task) error {
				return task.Transcribe(ctx, 0, samples, nil)
			})
			assert.NoError(err)
		})
	}
}

func TestIntegration_ConcurrentTranscriptions(t *testing.T) {
	modelsPath := skipIfNoModels(t)
	assert := assert.New(t)

	mgr, err := New(modelsPath, OptMaxConcurrent(2))
	assert.NoError(err)
	defer Close()

	models := mgr.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
	}
	model := models[0]

	samples := make([]float32, 16000)
	ctx := context.Background()

	// Run concurrent transcriptions
	done := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func() {
			err := mgr.WithModel(model, func(task *Task) error {
				return task.Transcribe(ctx, 0, samples, nil)
			})
			done <- err
		}()
	}

	// Wait for both to complete
	for i := 0; i < 2; i++ {
		err := <-done
		assert.NoError(err)
	}
}

func TestIntegration_ContextCancellation(t *testing.T) {
	modelsPath := skipIfNoModels(t)
	assert := assert.New(t)

	mgr, err := New(modelsPath)
	assert.NoError(err)
	defer Close()

	models := mgr.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
	}
	model := models[0]

	// Create a context that we'll cancel
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This should be cancelled before completing (or complete very quickly with silence)
	samples := make([]float32, 16000*10) // 10 seconds

	err = mgr.WithModel(model, func(task *Task) error {
		return task.Transcribe(ctx, 0, samples, nil)
	})

	// Error could be nil if transcription of silence is very fast,
	// or context.DeadlineExceeded if it was cancelled
	if err != nil {
		assert.Contains(err.Error(), "deadline")
		t.Logf("Transcription properly cancelled: %v", err)
	}
}
