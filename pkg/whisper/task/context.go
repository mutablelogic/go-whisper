package task

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	// Packages
	model "github.com/mutablelogic/go-whisper/pkg/whisper/model"
	whisper "github.com/mutablelogic/go-whisper/sys/whisper"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Context is used for running the transcription or translation
type Context struct {
	sync.Mutex

	model   string
	whisper *whisper.Context

	// Parameters for the next transcription
	params whisper.FullParams
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new context object
func New() *Context {
	return new(Context)
}

// Init the context
func (m *Context) Init(path string, model *model.Model, gpu int) error {
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
		fmt.Printf("Release model resources %v\n", ctx)
		whisper.Whisper_free(ctx.whisper)
	}
	ctx.whisper = nil
	ctx.model = ""

	// Return success
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx Context) MarshalJSON() ([]byte, error) {
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

func (ctx Context) String() string {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Context has a loaded model that matches the argument
func (ctx *Context) Is(model *model.Model) bool {
	if ctx.model == "" {
		return false
	}
	if model == nil {
		return false
	}
	return ctx.model == model.Id
}

// Copy task parameters from the default
func (task *Context) CopyParams() {
	task.params = whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)
	task.params.SetLanguage("auto")
}

// Model is multilingual and can translate
func (task *Context) CanTranslate() bool {
	return whisper.Whisper_is_multilingual(task.whisper)
}

// Transcribe samples. The samples should be 16KHz float32 samples in
// a single channel.
// TODO: We need a low-latency streaming version of this function.
// TODO: We need a callback for segment progress.
func (task *Context) Transcribe(ctx context.Context, samples []float32) error {
	// Set the 'abort' function
	/*task.params.SetAbortCallback(task.whisper, func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	})*/

	// Set the 'progress' function
	//task.params.SetProgressCallback(task.whisper, func(percent int) {
	//	fmt.Printf("Progress: %v\n", percent)
	//})

	// Perform the transcription
	if err := whisper.Whisper_full(task.whisper, task.params, samples); err != nil {
		return err
	}

	// Get segments
	for i := 0; i < task.whisper.NumSegments(); i++ {
		segment := task.whisper.Segment(i)
		fmt.Printf("Segment: %v\n", segment.Text)
	}

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

func (ctx *Context) SetTranslate(v bool) {
	ctx.params.SetTranslate(v)
}
