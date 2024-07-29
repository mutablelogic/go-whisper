package pool

import (
	"encoding/json"
	"fmt"

	// Packages
	model "github.com/mutablelogic/go-whisper/pkg/whisper/model"
	task "github.com/mutablelogic/go-whisper/pkg/whisper/task"

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

	// GPU flags
	gpu int
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new context pool of context objects, up to 'max' items
// Set the path for the model storage
// If GPU is -1 then disable, if 0 then use default, if >0 then enable
// and use the specified device
func NewContextPool(path string, max int, gpu int) *ContextPool {
	pool := new(ContextPool)
	pool.Pool = NewPool(max, func() any {
		return task.New()
	})
	pool.path = path
	pool.gpu = gpu

	// Return success
	return pool
}

// Close the pool and release all resources
func (m *ContextPool) Close() error {
	return m.Pool.Close()
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *ContextPool) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Gpu int `json:"gpu"`
		N   int `json:"n"`
		Max int `json:"max"`
	}{
		Gpu: m.gpu,
		N:   m.N(),
		Max: m.max,
	})
}

func (m *ContextPool) String() string {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get a context from the pool, for a model
func (m *ContextPool) Get(model *model.Model) (*task.Context, error) {
	// Check parameters
	if model == nil {
		return nil, ErrBadParameter
	}

	// Get a context from the pool
	t, ok := m.Pool.Get().(*task.Context)
	if !ok || t == nil {
		return nil, ErrChannelBlocked.With("unable to get a context from the pool, try again later")
	}

	// If the model matches, return it, or else release the resources
	if t.Is(model) {
		return t, nil
	} else if err := t.Close(); err != nil {
		return nil, err
	}

	// Initialise the context
	if err := t.Init(m.path, model, m.gpu); err != nil {
		return nil, err
	}

	// Return the context
	return t, nil
}

// Put a context back into the pool
func (m *ContextPool) Put(ctx *task.Context) {
	m.Pool.Put(ctx)
}

// Drain the pool of all contexts for a model, freeing resources
func (m *ContextPool) Drain(model *model.Model) error {
	fmt.Println("TODO: DRAIN", model.Id)
	return nil
}
