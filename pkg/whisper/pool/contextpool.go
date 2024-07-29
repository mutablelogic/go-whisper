package pool

import (

	// Packages
	"fmt"
	"path/filepath"

	model "github.com/mutablelogic/go-whisper/pkg/whisper/model"
	"github.com/mutablelogic/go-whisper/sys/whisper"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// ContextPool is a pool of context objects
type ContextPool struct {
	// Pool of context objects
	*Pool

	// Base path for models
	path string
}

// Context is used for running the transcription or translation
type Context struct {
	Model   *model.Model
	Context *whisper.Context
	Params  whisper.FullParams
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new context pool of context objects, up to 'max' items
// Set the path for the model storage
func NewContextPool(path string, max int32) *ContextPool {
	pool := new(ContextPool)
	pool.Pool = NewPool(max, func() any {
		return &Context{}
	})
	pool.path = path

	// Return success
	return pool
}

// Close the pool and release all resources
func (m *ContextPool) Close() error {
	return m.Pool.Close()
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
	m.Context = ctx
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
	if m.Context != nil {
		whisper.Whisper_free(m.Context)
	}
	m.Context = nil
	m.Model = nil

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get a context from the pool, for a model
func (m *ContextPool) Get(model *model.Model) (*Context, error) {
	// Check parameters
	if model == nil {
		return nil, ErrBadParameter
	}

	// Get a context from the pool
	ctx, ok := m.Pool.Get().(*Context)
	if !ok || ctx == nil {
		return nil, ErrChannelBlocked.With("unable to get a context from the pool, try again later")
	}

	// If the model matches, return it
	if ctx.Model != nil && ctx.Model.Id == model.Id {
		return ctx, nil
	}

	// Model didn't match: close the context
	if err := ctx.Close(); err != nil {
		return nil, err
	}

	// Initialise the context
	if err := ctx.Init(m.path, model); err != nil {
		return nil, err
	}

	// Return the context
	return ctx, nil
}

// Put a context back into the pool
func (m *ContextPool) Put(ctx *Context) {
	m.Pool.Put(ctx)
}

// Drain the pool of all contexts for a model, freeing resources
func (m *ContextPool) Drain(model *model.Model) error {
	fmt.Println("TODO: DRAIN", model.Id)
	return nil
}
