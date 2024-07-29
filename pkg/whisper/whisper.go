package whisper

import (
	"context"
	"errors"

	// Packages
	"github.com/mutablelogic/go-whisper/pkg/whisper/model"
	"github.com/mutablelogic/go-whisper/pkg/whisper/pool"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Whisper represents a whisper service for running transcription and translation
type Whisper struct {
	pool  *pool.ContextPool
	store *model.Store
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// This is the extension of the model files
	extModel = ".bin"

	// This is where the model is downloaded from
	defaultModelUrl = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/?download=true"
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new whisper service with the path to the models directory
// and optional parameters
func New(path string, opt ...Opt) (*Whisper, error) {
	var o opts

	// Set options
	o.MaxConcurrent = 1
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}

	// Create a new whisper service
	w := new(Whisper)
	if store, err := model.NewStore(path, extModel, defaultModelUrl); err != nil {
		return nil, err
	} else {
		w.store = store
	}
	if pool := pool.NewContextPool(path, int32(o.MaxConcurrent)); pool == nil {
		return nil, ErrInternalAppError
	} else {
		w.pool = pool
	}

	// Return success
	return w, nil
}

// Release all resources
func (w *Whisper) Close() error {
	var result error

	// Release pool resources
	if w.pool != nil {
		result = errors.Join(result, w.pool.Close())
	}

	// Set all to nil
	w.pool = nil
	w.store = nil

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all models in the models directory
func (w *Whisper) ListModels() []*model.Model {
	return w.store.List()
}

// Get a model by its Id, returns nil if the model does not exist
func (w *Whisper) GetModelById(id string) *model.Model {
	return w.store.ById(id)
}

// Delete a model by its id
func (w *Whisper) DeleteModelById(id string) error {
	model := w.store.ById(id)
	if model == nil {
		return ErrNotFound.Withf("%q", id)
	}

	// Empty the pool of this model
	if err := w.pool.Drain(model); err != nil {
		return err
	}

	// Delete the model
	if err := w.store.Delete(model.Id); err != nil {
		return err
	}

	// Return success
	return nil
}

// Download a model by path, where the directory is the root of the model
// within the models directory. The model is returned immediately if it
// already exists in the store
func (w *Whisper) DownloadModel(ctx context.Context, path string, fn func(curBytes, totalBytes uint64)) (*model.Model, error) {
	return w.store.Download(ctx, path, fn)
}

// Get a context for the specified model, which may load the model or return an existing one.
// The context can then be used to run the Transcribe function.
func (w *Whisper) WithModelContext(model *model.Model, fn func(ctx *pool.Context) error) error {
	if model == nil || fn == nil {
		return ErrBadParameter
	}

	// Get a context from the pool
	ctx, err := w.pool.Get(model)
	if err != nil {
		return err
	}
	defer w.pool.Put(ctx)

	// Execute the function
	return fn(ctx)
}
