package pool

import (

	// Packages

	"path/filepath"

	model "github.com/mutablelogic/go-whisper/pkg/whisper/model"
	"github.com/mutablelogic/go-whisper/sys/whisper"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Context is used for running the transcription or translation
type Context struct {
	Model *model.Model

	whisper *whisper.Context
	params  whisper.FullParams
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new context object
func New(path string, model *model.Model) (*Context, error) {
	ctx := new(Context)

	// Init the context
	if err := ctx.Init(path, model); err != nil {
		return nil, err
	}

	// Init the transcription with default parameters
	ctx.params = whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)

	// Return success
	return ctx, nil
}

// Init the context
func (m *Context) Init(path string, model *model.Model) error {
	// Check parameters
	if model == nil {
		return ErrBadParameter
	}

	// Get a context
	ctx := whisper.Whisper_init_from_file_with_params(filepath.Join(path, model.Path), whisper.DefaultContextParams())
	if ctx == nil {
		return ErrInternalAppError.With("whisper_init_from_file_with_params")
	}

	// Set resources
	m.whisper = ctx
	m.Model = model

	// Return success
	return nil
}

// Close the context and release all resources
func (m *Context) Close() error {
	var result error

	// Do nothing if nil
	if m == nil {
		return nil
	}

	// Release resources
	if m.whisper != nil {
		whisper.Whisper_free(m.whisper)
	}
	m.whisper = nil
	m.Model = nil

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Transcribe samples. The samples should be 16KHz float32 samples in
// a single channel.
// TODO: We need a low-latency streaming version of this function.
// TODO: We need a callback for segment progress.
func (ctx *Context) Transcribe(samples []float32) error {
	// Perform the transcription
	return whisper.Whisper_full(ctx.whisper, ctx.params, samples)
}
