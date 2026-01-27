package whisper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Packages
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

var (
	testManager    *Manager
	testModelsPath string
	cleanupNeeded  bool
)

// TestMain sets up the test environment by downloading models and initializing
// the whisper manager. It cleans up after all tests complete.
func TestMain(m *testing.M) {
	var exitCode int

	// Setup
	if err := setupTestEnvironment(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup test environment: %v\n", err)
		exitCode = 1
	} else {
		// Run tests
		exitCode = m.Run()
	}

	// Cleanup
	cleanupTestEnvironment()

	os.Exit(exitCode)
}

func setupTestEnvironment() error {
	// Create models directory
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return fmt.Errorf("could not create models directory: %w", err)
	}
	testModelsPath = modelsDir

	// Download model if needed
	modelPath := filepath.Join(modelsDir, modelTiny)
	if info, err := os.Stat(modelPath); err != nil || info.Size() < 8*1024*1024 {
		fmt.Printf("Downloading model: %s (this may take a minute...)\n", modelTiny)

		if err := downloadModel(modelPath); err != nil {
			fmt.Printf("Warning: Could not download model: %v\n", err)
			fmt.Println("Tests requiring models will be skipped")
			return nil // Don't fail, just skip model-dependent tests
		}

		cleanupNeeded = true // Mark for cleanup since we downloaded it
		fmt.Printf("Model downloaded successfully: %s\n", modelTiny)
	} else {
		fmt.Printf("Using existing model: %s\n", modelTiny)
	}

	// Initialize whisper manager
	mgr, err := New(modelsDir)
	if err != nil {
		return fmt.Errorf("could not initialize whisper manager: %w", err)
	}
	testManager = mgr

	fmt.Printf("Test environment ready with %d model(s)\n", len(testManager.ListModels()))
	return nil
}

func cleanupTestEnvironment() {
	// Close whisper manager
	if testManager != nil {
		testManager.Close()
		fmt.Println("Closed whisper manager")
	}

	// Delete downloaded model if we downloaded it
	if cleanupNeeded && testModelsPath != "" {
		modelPath := filepath.Join(testModelsPath, modelTiny)
		if err := os.Remove(modelPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not remove downloaded model: %v\n", err)
		} else {
			fmt.Printf("Cleaned up downloaded model: %s\n", modelTiny)
		}
	}
}

func downloadModel(modelPath string) error {
	// Create file for writing
	f, err := os.Create(modelPath)
	if err != nil {
		return fmt.Errorf("could not create model file: %w", err)
	}
	defer f.Close()

	// Download the model
	client := whisper.NewClient(modelURL)
	if client == nil {
		return fmt.Errorf("could not create HTTP client")
	}

	ctx := context.Background()
	_, err = client.Get(ctx, f, modelTiny)
	if err != nil {
		os.Remove(modelPath) // Clean up partial download
		return fmt.Errorf("download failed: %w", err)
	}

	// Verify download
	if info, err := os.Stat(modelPath); err != nil || info.Size() < 8*1024*1024 {
		os.Remove(modelPath)
		return fmt.Errorf("downloaded model is invalid")
	}

	return nil
}

// Helper to check if manager and models are available
func skipIfNoManager(t *testing.T) *Manager {
	if testManager == nil {
		t.Skip("Whisper manager not available")
		return nil
	}
	models := testManager.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
		return nil
	}
	return testManager
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
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	// Get available models
	models := mgr.ListModels()
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
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	// Get a model
	models := mgr.ListModels()
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
	err := mgr.WithModel(model, func(task *Task) error {
		return task.Transcribe(ctx, 0, samples, segmentCallback)
	})

	assert.NoError(err)
	// Segments may or may not be generated for silence
	t.Logf("Generated %d segment(s)", len(segments))
}

func TestIntegration_LanguageDetection(t *testing.T) {
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()
	model := models[0]

	// Create test samples
	samples := make([]float32, 16000)

	ctx := context.Background()
	var detectedLang string

	err := mgr.WithModel(model, func(task *Task) error {
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
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()

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

	err := mgr.WithModel(model, func(task *Task) error {
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
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()
	model := models[0]

	// Test different temperature values
	temperatures := []float64{0.0, 0.2, 0.8, 1.0}

	for _, temp := range temperatures {
		t.Run(string(rune(int(temp*10))), func(t *testing.T) {
			samples := make([]float32, 16000)
			ctx := context.Background()

			err := mgr.WithModel(model, func(task *Task) error {
				err := task.SetTemperature(temp)
				assert.NoError(err)

				return task.Transcribe(ctx, 0, samples, nil)
			})

			assert.NoError(err)
		})
	}
}

func TestIntegration_ContextReuse(t *testing.T) {
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()
	model := models[0]

	samples := make([]float32, 16000)
	ctx := context.Background()

	// Perform multiple transcriptions with the same model
	// This tests that the context pool properly reuses contexts
	for i := 0; i < 3; i++ {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			err := mgr.WithModel(model, func(task *Task) error {
				return task.Transcribe(ctx, 0, samples, nil)
			})
			assert.NoError(err)
		})
	}
}

func TestIntegration_ConcurrentTranscriptions(t *testing.T) {
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()
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

func TestIntegration_HighConcurrencyLimit(t *testing.T) {
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()
	model := models[0]

	samples := make([]float32, 16000)
	ctx := context.Background()

	// Run only 3 concurrent transcriptions
	// The shared manager has default capacity (NumCPU), but we're testing
	// that when we run fewer tasks than capacity, it works correctly
	const numTasks = 3
	done := make(chan error, numTasks)

	t.Logf("Running %d concurrent tasks with pool max=%d", numTasks, mgr.pool.max)

	for i := 0; i < numTasks; i++ {
		go func(taskNum int) {
			err := mgr.WithModel(model, func(task *Task) error {
				t.Logf("Task %d starting", taskNum)
				err := task.Transcribe(ctx, 0, samples, nil)
				t.Logf("Task %d completed", taskNum)
				return err
			})
			done <- err
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < numTasks; i++ {
		err := <-done
		assert.NoError(err)
	}
}

func TestIntegration_ConcurrencyLimitEnforced(t *testing.T) {
	if testModelsPath == "" {
		t.Skip("No models path available")
	}
	assert := assert.New(t)

	// Create a manager with a very low concurrency limit
	const maxConcurrent = 2

	// First, close the global test manager to free up resources
	if testManager != nil {
		testManager.Close()
		testManager = nil
	}

	mgr, err := New(testModelsPath, OptMaxConcurrent(maxConcurrent))
	assert.NoError(err)
	assert.NotNil(mgr)

	defer func() {
		if mgr != nil {
			mgr.Close()
		}
		// Reinitialize the global manager for other tests
		setupTestEnvironment()
	}()

	models := mgr.ListModels()
	if len(models) == 0 {
		t.Skip("No models available")
	}
	model := models[0]

	// Use a channel to hold tasks - we'll block them to ensure they overlap
	blocker := make(chan struct{})
	started := make(chan int, maxConcurrent+1)
	results := make(chan error, maxConcurrent+1)

	// Start maxConcurrent tasks that will block
	for i := 0; i < maxConcurrent; i++ {
		go func(id int) {
			err := mgr.WithModel(model, func(task *Task) error {
				started <- id
				<-blocker // Block until we release
				return nil
			})
			results <- err
		}(i)
	}

	// Wait for all blocking tasks to start
	for i := 0; i < maxConcurrent; i++ {
		select {
		case id := <-started:
			t.Logf("Task %d acquired a slot", id)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for tasks to start")
		}
	}

	// Now try to start one more task - it should fail immediately with "pool at capacity"
	err = mgr.WithModel(model, func(task *Task) error {
		t.Error("This task should not have started - pool should be at capacity")
		return nil
	})

	// Should get a "pool at capacity" error
	assert.Error(err)
	assert.Contains(err.Error(), "pool at capacity")
	t.Logf("Got expected error: %v", err)

	// Release the blocking tasks
	close(blocker)

	// Wait for blocked tasks to complete
	for i := 0; i < maxConcurrent; i++ {
		err := <-results
		assert.NoError(err)
	}

	// Now a new task should succeed
	err = mgr.WithModel(model, func(task *Task) error {
		t.Log("Task succeeded after pool freed up")
		return nil
	})
	assert.NoError(err)
}

func TestIntegration_TranscribeReaderJFK(t *testing.T) {
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()
	model := models[0]

	// Open the JFK sample file
	samplePath := getSamplePath(t, "jfk.wav")
	file, err := os.Open(samplePath)
	assert.NoError(err)
	defer file.Close()

	ctx := context.Background()
	var segments []*schema.Segment
	var result *schema.Transcription

	err = mgr.WithModel(model, func(task *Task) error {
		task.SetLanguage("en")

		// Use TranscribeReader to process the audio
		err := task.TranscribeReader(ctx, file, func(seg *schema.Segment) {
			segments = append(segments, seg)
			t.Logf("Segment: %s", seg.Text)
		})

		if err == nil {
			result = task.Result()
		}
		return err
	})

	assert.NoError(err)
	assert.NotNil(result)
	assert.NotEmpty(result.Text, "Should have transcribed text")
	assert.Equal("transcribe", result.Task)
	assert.NotEmpty(segments, "Should have generated segments")

	t.Logf("Full transcription: %s", result.Text)
	t.Logf("Language: %s", result.Language)
	t.Logf("Total segments: %d", len(segments))
}

func TestIntegration_TranscribeReaderMP3(t *testing.T) {
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()
	model := models[0]

	// Open an MP3 sample file
	samplePath := getSamplePath(t, "en-office.mp3")
	file, err := os.Open(samplePath)
	assert.NoError(err)
	defer file.Close()

	ctx := context.Background()
	var segmentCount int

	err = mgr.WithModel(model, func(task *Task) error {
		task.SetLanguage("en")

		// Use TranscribeReader with segmentation
		return task.TranscribeReader(ctx, file, func(seg *schema.Segment) {
			segmentCount++
			t.Logf("Segment %d: %s", segmentCount, seg.Text)
		})
	})

	assert.NoError(err)
	assert.Greater(segmentCount, 0, "Should have generated segments")

	t.Logf("Total segments: %d", segmentCount)
}

func TestIntegration_TranscribeReaderMultiLang(t *testing.T) {
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()

	// Find a multilingual model
	var model *schema.Model
	for _, m := range models {
		if m.Id != "tiny.en" && m.Id != "base.en" {
			model = m
			break
		}
	}
	if model == nil {
		t.Skip("No multilingual model available")
	}

	// Open the multi-language sample
	samplePath := getSamplePath(t, "multi-lang.wav")
	file, err := os.Open(samplePath)
	assert.NoError(err)
	defer file.Close()

	ctx := context.Background()
	var result *schema.Transcription

	err = mgr.WithModel(model, func(task *Task) error {
		// Auto-detect language
		task.SetLanguage("auto")

		err := task.TranscribeReader(ctx, file, func(seg *schema.Segment) {
			t.Logf("Segment: %s", seg.Text)
		})

		if err == nil {
			result = task.Result()
		}
		return err
	})

	assert.NoError(err)
	assert.NotNil(result)
	assert.NotEmpty(result.Language, "Should have detected language")
	t.Logf("Detected language: %s", result.Language)
}

func TestIntegration_ContextCancellation(t *testing.T) {
	mgr := skipIfNoManager(t)
	assert := assert.New(t)

	models := mgr.ListModels()
	model := models[0]

	// Create a context that we'll cancel
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This should be cancelled before completing (or complete very quickly with silence)
	samples := make([]float32, 16000*10) // 10 seconds

	err := mgr.WithModel(model, func(task *Task) error {
		return task.Transcribe(ctx, 0, samples, nil)
	})

	// Error could be nil if transcription of silence is very fast,
	// or context.DeadlineExceeded if it was cancelled
	if err != nil {
		assert.Contains(err.Error(), "deadline")
		t.Logf("Transcription properly cancelled: %v", err)
	}
}
