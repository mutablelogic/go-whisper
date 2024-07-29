package pool

import (
	"errors"
	"io"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Pool is a pool of context objects, up to a maximum number
// This acts as a cache so we don't need to reload models
// If the pool is full, then Get will return nil
type Pool struct {
	sync.RWMutex

	// The pool
	pool  []any
	fn    NewFunc
	n     int
	max   int
	empty bool
}

// Create a new object to place in the pool
type NewFunc func() any

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new pool of context objects, up to 'max' objects
func NewPool(max int, fn NewFunc) *Pool {
	// Max needs to be one or more
	if max <= 0 {
		return nil
	}
	// Create pool
	pool := new(Pool)
	pool.max = max
	pool.fn = func() any {
		if pool.atCapacity() {
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
	m.setEmpty(true)
	for {
		ctx := m.Get()
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

// Returns an item, or nil if the maximum number of contexts has been reached
func (m *Pool) Get() any {
	m.Lock()
	defer m.Unlock()

	// Create a new item
	var item any
	if len(m.pool) > 0 {
		item, m.pool = m.pool[0], m.pool[1:]
	} else {
		item = m.fn()
		if item != nil {
			m.n++
		}
	}
	return item
}

// Puts the context back in the pool
func (m *Pool) Put(ctx any) {
	m.Lock()
	defer m.Unlock()

	if ctx != nil {
		m.pool = append(m.pool, ctx)
		m.n--
	}
}

// Return the number of contexts in the pool
func (m *Pool) N() int {
	m.RLock()
	defer m.RUnlock()
	return m.n
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Return true if pool is at capacity
func (m *Pool) atCapacity() bool {
	return m.n >= m.max || m.empty
}

// Set pool in drain mode, no more contexts will be added
func (m *Pool) setEmpty(v bool) {
	m.Lock()
	defer m.Unlock()
	m.empty = v
}
