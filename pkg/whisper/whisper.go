package whisper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	// Packages
	model "github.com/mutablelogic/go-whisper/pkg/whisper/model"
	pool "github.com/mutablelogic/go-whisper/pkg/whisper/pool"
	task "github.com/mutablelogic/go-whisper/pkg/whisper/task"
	whisper "github.com/mutablelogic/go-whisper/sys/whisper"

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

	if pool := pool.NewContextPool(path, o.MaxConcurrent, o.gpu); pool == nil {
		return nil, ErrInternalAppError
	} else {
		w.pool = pool
	}

	// Logging
	if o.logfn != nil {
		whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
			if !o.debug && level > whisper.LogLevelError {
				return
			}
			o.logfn(fmt.Sprintf("[%s] %s", level, strings.TrimSpace(text)))
		})
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
// STRINGIFY

func (w *Whisper) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Store *model.Store      `json:"store"`
		Pool  *pool.ContextPool `json:"pool"`
	}{
		Store: w.store,
		Pool:  w.pool,
	})
}

func (w *Whisper) String() string {
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
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

// Get a task for the specified model, which may load the model or
// return an existing one. The context can then be used to run the Transcribe
// function, and after the context is returned to the pool.
func (w *Whisper) WithModel(model *model.Model, fn func(task *task.Context) error) error {
	if model == nil || fn == nil {
		return ErrBadParameter
	}

	// Get a context from the pool
	task, err := w.pool.Get(model)
	if err != nil {
		return err
	}
	defer w.pool.Put(task)

	// Execute the function
	return fn(task)
}
