package whisper

import (
	"bytes"
	"context"
	"testing"
	"time"

	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	assert "github.com/stretchr/testify/assert"
)

func TestTask_NewTask(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	assert.NotNil(task)
	assert.Empty(task.model)
	assert.Nil(task.whisper)
	assert.Nil(task.result)
}

func TestTask_Close_Nil(t *testing.T) {
	assert := assert.New(t)
	var task *Task
	err := task.Close()
	assert.NoError(err)
}

func TestTask_Close_Uninit(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	err := task.Close()
	assert.NoError(err)
}

func TestTask_Init_NilModel(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	err := task.Init("", nil, 0, nil)
	assert.Error(err)
}

func TestTask_Init_InvalidPath(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	model := &schema.Model{
		Id:   "test",
		Path: "nonexistent.bin",
	}
	err := task.Init("/nonexistent", model, 0, nil)
	assert.Error(err)
}

func TestTask_Is(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()

	// Uninitialized task
	assert.False(task.Is(nil))
	assert.False(task.Is(&schema.Model{Id: "test"}))

	// Set model manually for testing
	task.model = "test-model"
	assert.False(task.Is(nil))
	assert.True(task.Is(&schema.Model{Id: "test-model"}))
	assert.False(task.Is(&schema.Model{Id: "other-model"}))
}

func TestTask_CopyParams(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()

	// CopyParams should initialize params and result
	task.CopyParams()
	assert.NotNil(task.result)
	assert.Equal("auto", task.params.Language())
}

func TestTask_SetTemperature(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	task.CopyParams()

	// Valid temperatures
	err := task.SetTemperature(0.0)
	assert.NoError(err)

	err = task.SetTemperature(0.5)
	assert.NoError(err)

	err = task.SetTemperature(1.0)
	assert.NoError(err)

	// Invalid temperatures
	err = task.SetTemperature(-0.1)
	assert.Error(err)

	err = task.SetTemperature(1.1)
	assert.Error(err)
}

func TestTask_SetLanguage(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	task.CopyParams()

	// Auto detection
	err := task.SetLanguage("")
	assert.NoError(err)
	assert.Equal("auto", task.Language())

	err = task.SetLanguage("auto")
	assert.NoError(err)
	assert.Equal("auto", task.Language())

	// Valid languages
	err = task.SetLanguage("en")
	assert.NoError(err)
	assert.Equal("en", task.Language())

	// Invalid language
	err = task.SetLanguage("invalid_language_xyz")
	assert.Error(err)
}

func TestTask_SetTranslate(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	task.CopyParams()

	task.SetTranslate(true)
	assert.True(task.Translate())

	task.SetTranslate(false)
	assert.False(task.Translate())
}

func TestTask_SetDiarize(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	task.CopyParams()

	task.SetDiarize(true)
	assert.True(task.Diarize())

	task.SetDiarize(false)
	assert.False(task.Diarize())
}

func TestTask_SetPrompt(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	task.CopyParams()

	err := task.SetPrompt("test prompt")
	assert.NoError(err)

	err = task.SetPrompt("")
	assert.NoError(err)
}

func TestTask_Result(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()

	// Before CopyParams
	assert.Nil(task.Result())

	// After CopyParams
	task.CopyParams()
	result := task.Result()
	assert.NotNil(result)
	assert.Equal("", result.Text)
	assert.Equal("", result.Language)
	assert.Equal("", result.Task)
	assert.Empty(result.Segments)
}

func TestTask_MarshalJSON(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	task.model = "test-model"
	task.CopyParams()

	data, err := task.MarshalJSON()
	assert.NoError(err)
	assert.NotEmpty(data)

	// Verify it's valid JSON
	assert.Contains(string(data), "test-model")
}

func TestTask_String(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	task.model = "test-model"
	task.CopyParams()

	str := task.String()
	assert.NotEmpty(str)
	assert.Contains(str, "test-model")
}

func TestWriteSegmentFunctions(t *testing.T) {
	assert := assert.New(t)

	segment := &schema.Segment{
		Id:    1,
		Text:  "Test segment",
		Start: schema.Timestamp(1 * time.Second),
		End:   schema.Timestamp(2 * time.Second),
	}

	// Test WriteSegmentSrt
	var buf bytes.Buffer
	WriteSegmentSrt(&buf, segment)
	assert.NotEmpty(buf.String())

	// Test WriteSegmentVtt
	buf.Reset()
	WriteSegmentVtt(&buf, segment)
	assert.NotEmpty(buf.String())

	// Test WriteSegmentText
	buf.Reset()
	WriteSegmentText(&buf, segment)
	assert.Equal("\nTest segment", buf.String())
}

func TestNewSegmentFunc_Type(t *testing.T) {
	assert := assert.New(t)

	// Test that NewSegmentFunc is a valid function type
	var fn NewSegmentFunc = func(seg *schema.Segment) {
		assert.NotNil(seg)
	}

	segment := &schema.Segment{
		Id:   1,
		Text: "Test",
	}

	fn(segment)
}

func TestTask_Transcribe_NilContext(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	task.CopyParams()

	// Should handle context properly
	ctx := context.Background()
	samples := make([]float32, 16000) // 1 second of silence at 16kHz

	// This will fail because task is not initialized with a model,
	// but it tests that the context handling works
	err := task.Transcribe(ctx, 0, samples, nil)
	assert.Error(err) // Expected to fail without initialized whisper context
}

func TestTask_Transcribe_CanceledContext(t *testing.T) {
	assert := assert.New(t)
	task := NewTask()
	task.CopyParams()

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	samples := make([]float32, 16000)

	err := task.Transcribe(ctx, 0, samples, nil)
	assert.Error(err) // Should error due to canceled context or nil whisper
}
