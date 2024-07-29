package pool

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Pool is a pool of context objects, up to a maximum number
// This acts as a cache so we don't need to reload models
// If the pool is full, then Get will return nil
type Pool struct {
	// Pool of context objects
	pool sync.Pool

	// Number of contexts in the pool
	n int32

	// Maximum number of contexts in the pool
	max, empty int32
}

// Create a new object to place in the pool
type NewFunc func() any

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new pool of context objects, up to 'max'
// objects
func NewPool(max int32, fn NewFunc) *Pool {
	// Max needs to be one or more
	if max <= 0 {
		return nil
	}
	// Create pool
	pool := &Pool{max: max}
	pool.pool.New = func() any {
		if pool.n >= pool.max {
			return nil
		}
		if int(atomic.LoadInt32(&pool.empty)) != 0 {
			return nil
		}
		return fn()
	}

	// Return success
	return pool
}

func (m *Pool) Close() error {
	var result error

	// We repeatedly call Get until we get nil
	atomic.StoreInt32(&m.empty, 1)
	for {
		ctx := m.pool.Get()
		if ctx == nil {
			break
		}
		// If it is an io.Closer, then close it
		if ctx, ok := ctx.(io.Closer); ok {
			result = errors.Join(result, ctx.Close())
		}
	}

	// Return any error
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Returns an unused context, or nil if the maximum number of contexts has been reached
func (m *Pool) Get() any {
	// Get a context from the pool
	if ctx := m.pool.Get(); ctx != nil {
		atomic.AddInt32(&m.n, 1)
		return ctx
	}
	return nil
}

// Puts the context back in the pool
func (m *Pool) Put(ctx any) {
	if ctx != nil {
		atomic.AddInt32(&m.n, -1)
		m.pool.Put(ctx)
	}
}

// Return the number of contexts in the pool
func (m *Pool) N() int {
	return int(atomic.LoadInt32(&m.n))
}
