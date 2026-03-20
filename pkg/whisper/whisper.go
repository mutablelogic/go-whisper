package whisper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	store "github.com/mutablelogic/go-whisper/pkg/whisper/store"
	whisper "github.com/mutablelogic/go-whisper/sys/whisper"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// contextPool is a simple pool of Task objects
type contextPool struct {
	sync.Mutex
	*opts
	path      string
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
	globalManager.Lock()
	defer globalManager.Unlock()

	// If already initialized, then return error, else create the model store
	if globalManager.store != nil {
		return nil, httpresponse.ErrInternalError.With("whisper manager already initialized")
	} else if store, err := store.NewStore(path, extModel, defaultModelUrl); err != nil {
		return nil, err
	} else {
		globalManager.store = store
	}

	// Set options
	o, err := applyOpts(opt...)
	if err != nil {
		return nil, err
	}

	// Create a context pool
	if pool := newContextPool(path, &o); pool == nil {
		return nil, httpresponse.ErrInternalError.With("unable to create context pool")
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

// newContextPool creates a simple context pool
func newContextPool(path string, opt *opts) *contextPool {
	if opt.max <= 0 {
		return nil
	}
	return &contextPool{
		opts:      opt,
		path:      path,
		available: make([]*Task, 0, opt.max),
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
		return httpresponse.ErrNotFound.Withf("%q", id)
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

// AcquireTask gets a task for the specified model from the pool without
// scoping it to a callback. The caller MUST call the returned release
// function when done (typically via defer) to return the task to the pool.
// This is intended for long-lived sessions such as streaming transcription.
func (m *Manager) AcquireTask(model *schema.Model) (*Task, func(), error) {
	if model == nil {
		return nil, nil, httpresponse.ErrBadRequest.With("model must be non-nil")
	}

	m.RLock()
	pool := m.pool
	m.RUnlock()

	if pool == nil {
		return nil, nil, httpresponse.ErrInternalError.With("pool not initialized")
	}

	task, err := pool.get(model)
	if err != nil {
		return nil, nil, err
	}

	task.CopyParams()

	release := func() { pool.put(task) }
	return task, release, nil
}

// WithModel gets a task for the specified model and executes the function.
// The task is automatically returned to the pool when done.
func (m *Manager) WithModel(model *schema.Model, fn func(task *Task) error, opts ...Opt) error {
	if model == nil || fn == nil {
		return httpresponse.ErrBadRequest.With("model and function must be non-nil")
	}

	m.RLock()
	pool := m.pool
	m.RUnlock()

	if pool == nil {
		return httpresponse.ErrInternalError.With("pool not initialized")
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
	p.Lock()
	defer p.Unlock()

	// Check parameters
	if model == nil {
		return nil, httpresponse.ErrBadRequest.With("model is nil")
	}

	// Try to take a task from available pool (prefers matching model)
	// No available task - need to create one
	task, needsInit := p.takeTask(model)
	if task == nil {
		if p.inUse >= p.max {
			return nil, httpresponse.Err(http.StatusServiceUnavailable).With("pool at capacity, try again later")
		}
		task = NewTask()
		needsInit = true
	}

	// Initialize if needed
	if needsInit {
		if err := task.Init(p.path, model, p.gpu, p.tracer); err != nil {
			return nil, err
		}
	}

	p.inUse++
	return task, nil
}

// takeTask removes and returns a task from the available pool.
// Prefers tasks that already have the model loaded (needsInit=false).
// Returns nil if no tasks are available.
func (p *contextPool) takeTask(model *schema.Model) (task *Task, needsInit bool) {
	// First, look for a task with the same model already loaded
	for i, t := range p.available {
		if t.Is(model) {
			p.available = append(p.available[:i], p.available[i+1:]...)
			return t, false
		}
	}

	// Take any available task (will need reinitialization)
	if len(p.available) > 0 {
		task = p.available[len(p.available)-1]
		p.available = p.available[:len(p.available)-1]
		task.Close()
		return task, true
	}

	return nil, true
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
