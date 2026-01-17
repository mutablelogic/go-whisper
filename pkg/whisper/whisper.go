package whisper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	store "github.com/mutablelogic/go-whisper/pkg/whisper/store"
	whisper "github.com/mutablelogic/go-whisper/sys/whisper"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// contextPool is a simple pool of Task objects
type contextPool struct {
	sync.Mutex
	path      string
	gpu       int
	max       int
	available []*Task
	inUse     int
}

// Manager provides a global whisper manager with automatic cleanup
type Manager struct {
	sync.RWMutex
	pool  *contextPool
	store *store.Store
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// Global manager instance
	globalManager = &Manager{}
)

const (
	// This is the extension of the model files
	extModel = ".bin"

	// This is where the model is downloaded from
	defaultModelUrl = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/?download=true"

	// Sample Rate
	SampleRate = whisper.SampleRate
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New initializes the global whisper manager with the specified path to
// the models directory and optional parameters. Returns an error if the
// manager is already initialized.
func New(path string, opt ...Opt) (*Manager, error) {
	var o opts
	globalManager.Lock()
	defer globalManager.Unlock()

	// Set options
	o.MaxConcurrent = runtime.NumCPU()
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}

	// If already initialized, then return error
	if globalManager.store != nil {
		return nil, ErrInternalAppError.With("whisper manager already initialized")
	}

	// Create a model store
	if store, err := store.NewStore(path, extModel, defaultModelUrl); err != nil {
		return nil, err
	} else {
		globalManager.store = store
	}

	// Create a context pool
	if pool := newContextPool(path, o.MaxConcurrent, o.gpu); pool == nil {
		return nil, ErrInternalAppError.With("unable to create context pool")
	} else {
		globalManager.pool = pool
	}

	// Logging - always set up to override whisper's default stderr logging
	// Only log errors by default, or all levels if debug is enabled
	logfn := o.logfn
	if logfn == nil {
		// Suppress all logging if no log function provided
		logfn = func(string) {}
	}

	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		if !o.debug && level > whisper.LogLevelError {
			return
		}
		logfn(fmt.Sprintf("[%s] %s", level, strings.TrimSpace(text)))
	})
	ffmpeg.SetLogging(o.debug, func(text string) {
		logfn(text)
	})

	// Return success
	return globalManager, nil
}

// Close closes the whisper manager and releases all resources.
// It is safe to call multiple times.
func (m *Manager) Close() error {
	m.Lock()
	defer m.Unlock()

	// Check if initialized
	if m.store == nil {
		return nil
	}

	// Release pool resources
	var result error
	if m.pool != nil {
		result = errors.Join(result, m.pool.close())
	}

	// Clean up sys/whisper resources
	whisper.Whisper_log_set(nil)
	whisper.CleanupAllCallbacks()

	// Set all to nil
	m.pool = nil
	m.store = nil

	// Return any errors
	return result
}

// Close closes the global whisper manager. This is a convenience function
// that calls Close() on the global manager instance.
func Close() error {
	return globalManager.Close()
}

// newContextPool creates a simple context pool
func newContextPool(path string, max int, gpu int) *contextPool {
	if max <= 0 {
		return nil
	}
	return &contextPool{
		path:      path,
		gpu:       gpu,
		max:       max,
		available: make([]*Task, 0, max),
	}
}

// close releases all tasks in the pool
func (p *contextPool) close() error {
	p.Lock()
	defer p.Unlock()

	var result error
	for _, task := range p.available {
		result = errors.Join(result, task.Close())
	}
	p.available = nil
	p.inUse = 0
	return result
}

// stats returns pool statistics for JSON marshaling
func (p *contextPool) stats() map[string]int {
	p.Lock()
	defer p.Unlock()
	return map[string]int{
		"available": len(p.available),
		"in_use":    p.inUse,
		"max":       p.max,
		"gpu":       p.gpu,
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *Manager) MarshalJSON() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()

	poolStats := map[string]int{}
	if m.pool != nil {
		poolStats = m.pool.stats()
	}

	return json.Marshal(struct {
		Store *store.Store   `json:"store"`
		Pool  map[string]int `json:"pool"`
	}{
		Store: m.store,
		Pool:  poolStats,
	})
}

func (m *Manager) String() string {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListModels returns all models in the models directory
func (m *Manager) ListModels() []*schema.Model {
	m.RLock()
	defer m.RUnlock()
	return m.store.List()
}

// GetModelById returns a model by its Id, returns nil if the model does not exist
func (m *Manager) GetModelById(id string) *schema.Model {
	m.RLock()
	defer m.RUnlock()
	return m.store.ById(id)
}

// DeleteModelById deletes a model by its id
func (m *Manager) DeleteModelById(id string) error {
	m.Lock()
	defer m.Unlock()

	model := m.store.ById(id)
	if model == nil {
		return ErrNotFound.Withf("%q", id)
	}

	// Delete the model
	if err := m.store.Delete(model.Id); err != nil {
		return err
	}

	// Return success
	return nil
}

// DownloadModel downloads a model by path, where the directory is the root of the model
// within the models directory. The model is returned immediately if it
// already exists in the store
func (m *Manager) DownloadModel(ctx context.Context, path string, fn func(curBytes, totalBytes uint64)) (*schema.Model, error) {
	m.RLock()
	defer m.RUnlock()
	return m.store.Download(ctx, path, fn)
}

// WithModel gets a task for the specified model and executes the function.
// The task is automatically returned to the pool when done.
func (m *Manager) WithModel(model *schema.Model, fn func(task *Task) error) error {
	if model == nil || fn == nil {
		return ErrBadParameter
	}

	m.RLock()
	pool := m.pool
	m.RUnlock()

	if pool == nil {
		return ErrInternalAppError.With("pool not initialized")
	}

	// Get a task from the pool
	task, err := pool.get(model)
	if err != nil {
		return err
	}
	defer pool.put(task)

	// Copy parameters
	task.CopyParams()

	// Execute the function
	return fn(task)
}

// get retrieves a task from the pool or creates a new one
func (p *contextPool) get(model *schema.Model) (*Task, error) {
	if model == nil {
		return nil, ErrBadParameter
	}

	p.Lock()
	defer p.Unlock()

	// Try to reuse an existing task
	for i, task := range p.available {
		if task.Is(model) {
			// Remove from available and return
			p.available = append(p.available[:i], p.available[i+1:]...)
			p.inUse++
			return task, nil
		}
	}

	// Check if we can create a new task
	if p.inUse+len(p.available) >= p.max {
		// Try to reuse any available task
		if len(p.available) > 0 {
			task := p.available[0]
			p.available = p.available[1:]
			p.inUse++

			// Close old model and init new one
			if err := task.Close(); err != nil {
				p.inUse--
				return nil, err
			}
			if err := task.Init(p.path, model, p.gpu); err != nil {
				p.inUse--
				return nil, err
			}
			return task, nil
		}
		return nil, ErrChannelBlocked.With("pool at capacity, try again later")
	}

	// Create new task
	task := NewTask()
	if err := task.Init(p.path, model, p.gpu); err != nil {
		return nil, err
	}
	p.inUse++
	return task, nil
}

// put returns a task to the pool
func (p *contextPool) put(task *Task) {
	if task == nil {
		return
	}

	p.Lock()
	defer p.Unlock()

	p.available = append(p.available, task)
	p.inUse--
}
