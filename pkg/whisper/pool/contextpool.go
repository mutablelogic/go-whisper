package pool

import (
	"fmt"

	// Packages
	model "github.com/mutablelogic/go-whisper/pkg/whisper/model"
	context "github.com/mutablelogic/go-whisper/pkg/whisper/task"

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

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new context pool of context objects, up to 'max' items
// Set the path for the model storage
func NewContextPool(path string, max int32) *ContextPool {
	pool := new(ContextPool)
	pool.Pool = NewPool(max, func() any {
		return context.New()
	})
	pool.path = path

	// Return success
	return pool
}

// Close the pool and release all resources
func (m *ContextPool) Close() error {
	return m.Pool.Close()
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get a context from the pool, for a model
func (m *ContextPool) Get(model *model.Model) (*context.Context, error) {
	// Check parameters
	if model == nil {
		return nil, ErrBadParameter
	}

	// Get a context from the pool
	ctx, ok := m.Pool.Get().(*context.Context)
	if !ok || ctx == nil {
		return nil, ErrChannelBlocked.With("unable to get a context from the pool, try again later")
	}

	// If the model matches, return it
	if ctx.Is(model) {
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
func (m *ContextPool) Put(ctx *context.Context) {
	m.Pool.Put(ctx)
}

// Drain the pool of all contexts for a model, freeing resources
func (m *ContextPool) Drain(model *model.Model) error {
	fmt.Println("TODO: DRAIN", model.Id)
	return nil
}
