package task

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	// Packages
	schema "github.com/mutablelogic/go-whisper/pkg/whisper/schema"
	whisper "github.com/mutablelogic/go-whisper/sys/whisper"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Context is used for running the transcription or translation
type Context struct {
	sync.Mutex

	// Model Id and whisper context
	model   string
	whisper *whisper.Context

	// Parameters for the next transcription
	params whisper.FullParams

	// Collect the transcription
	result *schema.Transcription
}

// Callback for new segments during the transcription process
type NewSegmentFunc func(*schema.Segment)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new context object
func New() *Context {
	return new(Context)
}

// Init the context
func (m *Context) Init(path string, model *schema.Model, gpu int) error {
	m.Lock()
	defer m.Unlock()

	// Check parameters
	if model == nil {
		return ErrBadParameter
	}

	// Get default parameters
	params := whisper.DefaultContextParams()

	// If gpu is -1, then disable
	// If gpu is 0, then use whatever the default is
	// If gpu is >0, then enable and set the device
	if gpu == -1 {
		params.SetUseGpu(false)
	} else if gpu > 0 {
		params.SetUseGpu(true)
		params.SetGpuDevice(gpu)
	}

	// Get a context
	ctx := whisper.Whisper_init_from_file_with_params(filepath.Join(path, model.Path), params)
	if ctx == nil {
		return ErrInternalAppError.With("whisper_init")
	}

	// Set resources
	m.whisper = ctx
	m.model = model.Id

	// Return success
	return nil
}

// Close the context and release all resources. The context
// itself can be re-used by calling Init again
func (ctx *Context) Close() error {
	// Do nothing if nil
	if ctx == nil {
		return nil
	}

	// Release resources
	if ctx.whisper != nil {
		whisper.Whisper_free(ctx.whisper)
	}
	ctx.whisper = nil
	ctx.model = ""

	// Return success
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *Context) MarshalJSON() ([]byte, error) {
	type j struct {
		Model   string             `json:"model"`
		Params  whisper.FullParams `json:"params"`
		Context string             `json:"context"`
	}
	return json.Marshal(j{
		Model:   ctx.model,
		Params:  ctx.params,
		Context: fmt.Sprintf("%p", ctx.whisper),
	})
}

func (ctx *Context) String() string {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Context has a loaded model that matches the argument
func (ctx *Context) Is(model *schema.Model) bool {
	if ctx.model == "" {
		return false
	}
	if model == nil {
		return false
	}
	return ctx.model == model.Id
}

// Reset task context for re-use
func (task *Context) CopyParams() {
	task.params = whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)
	task.params.SetLanguage("auto")
	task.result = nil
}

// Model is multilingual and can translate
func (task *Context) CanTranslate() bool {
	return whisper.Whisper_is_multilingual(task.whisper)
}

// Transcribe samples. The samples should be 16KHz float32 samples in
// a single channel. Appends the transcription to the result, and includes
// segment data if segments is true.
func (task *Context) Transcribe(ctx context.Context, ts time.Duration, samples []float32, segments bool, fn NewSegmentFunc) error {
	// Set the 'abort' function
	task.params.SetAbortCallback(task.whisper, func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	})

	// Set the new segment function
	if fn != nil {
		task.params.SetSegmentCallback(task.whisper, func(new_segments int) {
			num_segments := task.whisper.NumSegments()
			for i := num_segments - new_segments; i < num_segments; i++ {
				fn(newSegment(ts, task.whisper.Segment(i)))
			}
		})
	}

	// TODO: Set the initial prompt tokens from any previous transcription call

	// Perform the transcription
	if err := whisper.Whisper_full(task.whisper, task.params, samples); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		} else {
			return err
		}
	}

	// Remove the callbacks
	task.params.SetAbortCallback(task.whisper, nil)
	task.params.SetSegmentCallback(task.whisper, nil)

	// Append the transcription
	if task.result == nil {
		task.result = new(schema.Transcription)
	}
	task.appendResult(ts, segments)

	// Return success
	return nil
}

// Set the language. For transcription, this is the language of the
// audio samples. For translation, this is the language to translate
// to. If you set this to "auto" then the language will be detected
func (ctx *Context) SetLanguage(v string) error {
	if v == "" || v == "auto" {
		ctx.params.SetLanguage("auto")
		return nil
	}
	id := whisper.Whisper_lang_id(v)
	if id == -1 {
		return ErrBadParameter.Withf("invalid language: %q", v)
	}
	ctx.params.SetLanguage(v)
	return nil
}

// Set translate to true or false
func (ctx *Context) SetTranslate(v bool) {
	ctx.params.SetTranslate(v)
}

// Return the transcription result
func (ctx *Context) Result() *schema.Transcription {
	return ctx.result
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (ctx *Context) appendResult(ts time.Duration, segments bool) {
	// Append text
	for i := 0; i < ctx.whisper.NumSegments(); i++ {
		seg := ctx.whisper.Segment(i)
		ctx.result.Text += seg.Text
	}
	if segments {
		// Append segments
		for i := 0; i < ctx.whisper.NumSegments(); i++ {
			ctx.result.Segments = append(ctx.result.Segments, newSegment(ts, ctx.whisper.Segment(i)))
		}
	}
}
