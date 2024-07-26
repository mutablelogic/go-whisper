package whisper

import (
	"errors"
	"sync"
	"sync/atomic"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ModelPool struct {
	// Base path
	path string

	// Model
	model *Model

	// Pool of context objects
	pool sync.Pool

	// Number of contexts in the pool
	n int32

	// Maximum number of contexts in the pool
	max int32
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultMaxPoolSize = 10
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewModelPool(path string, model *Model) *ModelPool {
	return &ModelPool{
		path:  path,
		model: model,
		max:   defaultMaxPoolSize,
	}
}

func (m *ModelPool) Close() error {
	var result error

	// We repeatedly call Get until we get nil
	m.max = 0
	for {
		ctx := m.pool.Get().(*Context)
		if ctx == nil {
			break
		}
		result = errors.Join(result, ctx.Close())
	}

	// Return any error
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Returns a newly loaded model context, or nil if there is an error
func (m *ModelPool) Get() *Context {
	// Get a context from the pool
	ctx, _ := m.pool.Get().(*Context)

	// Check for maximum number of contexts
	if ctx == nil && m.N() >= int(m.max) {
		return nil
	}

	// Create a new context object
	if ctx == nil {
		ctx = NewContextWithModel(m.path, m.model)
	}

	// Increment counter
	if ctx != nil {
		atomic.AddInt32(&m.n, 1)
	}

	// Return the context
	return ctx
}

// Puts the context back in the pool
func (m *ModelPool) Put(ctx *Context) {
	if ctx != nil {
		atomic.AddInt32(&m.n, -1)
		m.pool.Put(ctx)
	}
}

// Return the number of contexts in the pool
func (m *ModelPool) N() int {
	return int(atomic.LoadInt32(&m.n))
}
