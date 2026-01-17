package whisper

import (
	"context"
	"encoding/json"

	// Packages
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Whisper represents a whisper service for running transcription and translation
// This is a compatibility wrapper around pkg/whisper.Manager
type Whisper struct {
	mgr *whisper.Manager
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Sample Rate
	SampleRate = whisper.SampleRate
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new whisper service with the path to the models directory
// and optional parameters
func New(path string, opt ...Opt) (*Whisper, error) {
	mgr, err := whisper.New(path, opt...)
	if err != nil {
		return nil, err
	}

	return &Whisper{mgr: mgr}, nil
}

// Release all resources
func (w *Whisper) Close() error {
	return whisper.Close()
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (w *Whisper) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.mgr)
}

func (w *Whisper) String() string {
	return w.mgr.String()
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all models in the models directory
func (w *Whisper) ListModels() []*schema.Model {
	return w.mgr.ListModels()
}

// Get a model by its Id, returns nil if the model does not exist
func (w *Whisper) GetModelById(id string) *schema.Model {
	return w.mgr.GetModelById(id)
}

// Delete a model by its id
func (w *Whisper) DeleteModelById(id string) error {
	return w.mgr.DeleteModelById(id)
}

// Download a model by path, where the directory is the root of the model
// within the models directory. The model is returned immediately if it
// already exists in the store
func (w *Whisper) DownloadModel(ctx context.Context, path string, fn func(curBytes, totalBytes uint64)) (*schema.Model, error) {
	return w.mgr.DownloadModel(ctx, path, fn)
}

// WithModel gets a task for the specified model and executes the function.
// The task is automatically returned to the pool when done.
func (w *Whisper) WithModel(model *schema.Model, fn func(task *whisper.Task) error) error {
	return w.mgr.WithModel(model, fn)
}
