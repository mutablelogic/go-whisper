package whisper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_NewAndClose(t *testing.T) {
	assert := assert.New(t)
	Close()

	path := t.TempDir()
	mgr, err := New(path)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(mgr)

	globalManager.RLock()
	assert.NotNil(globalManager.store)
	assert.NotNil(globalManager.pool)
	globalManager.RUnlock()

	err = Close()
	assert.NoError(err)

	globalManager.RLock()
	assert.Nil(globalManager.store)
	assert.Nil(globalManager.pool)
	globalManager.RUnlock()
}

func TestManager_DoubleInitialize(t *testing.T) {
	assert := assert.New(t)
	Close()

	path := t.TempDir()
	_, err := New(path)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Try to initialize again
	_, err = New(path)
	assert.Error(err)
	assert.Contains(err.Error(), "already initialized")

	Close()
}

func TestManager_CloseMultipleTimes(t *testing.T) {
	assert := assert.New(t)
	Close()

	path := t.TempDir()
	_, err := New(path)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Close multiple times should be safe
	assert.NoError(Close())
	assert.NoError(Close())
	assert.NoError(Close())
}

func TestManager_WithMaxConcurrent(t *testing.T) {
	assert := assert.New(t)
	Close()

	path := t.TempDir()
	mgr, err := New(path, OptMaxConcurrent(4))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(mgr)

	globalManager.RLock()
	assert.Equal(4, globalManager.pool.max)
	globalManager.RUnlock()

	Close()
}

func TestManager_WithNoGPU(t *testing.T) {
	assert := assert.New(t)
	Close()

	path := t.TempDir()
	mgr, err := New(path, OptNoGPU())
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(mgr)

	globalManager.RLock()
	assert.Equal(-1, globalManager.pool.gpu)
	globalManager.RUnlock()

	Close()
}

func TestManager_WithDebug(t *testing.T) {
	assert := assert.New(t)
	Close()

	path := t.TempDir()
	called := false
	mgr, err := New(path, OptDebug(), OptLog(func(s string) {
		called = true
	}))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(mgr)

	// Note: called may or may not be true depending on whether whisper logs during init
	_ = called

	Close()
}

func TestManager_ListModels_Empty(t *testing.T) {
	assert := assert.New(t)
	Close()

	path := t.TempDir()
	mgr, err := New(path)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	models := mgr.ListModels()
	assert.NotNil(models)
	assert.Empty(models)

	Close()
}

func TestManager_GetModelById_NotFound(t *testing.T) {
	assert := assert.New(t)
	Close()

	path := t.TempDir()
	mgr, err := New(path)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	model := mgr.GetModelById("nonexistent")
	assert.Nil(model)

	Close()
}

func TestManager_DeleteModelById_NotFound(t *testing.T) {
	assert := assert.New(t)
	Close()

	path := t.TempDir()
	mgr, err := New(path)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	err = mgr.DeleteModelById("nonexistent")
	assert.Error(err)
	assert.Contains(err.Error(), "nonexistent")

	Close()
}

func TestContextPool_Stats(t *testing.T) {
	assert := assert.New(t)

	pool := newContextPool("/tmp", 5, 0)
	assert.NotNil(pool)

	stats := pool.stats()
	assert.Equal(0, stats["available"])
	assert.Equal(0, stats["in_use"])
	assert.Equal(5, stats["max"])
	assert.Equal(0, stats["gpu"])
}

func TestContextPool_NilCreation(t *testing.T) {
	assert := assert.New(t)

	pool := newContextPool("/tmp", 0, 0)
	assert.Nil(pool)

	pool = newContextPool("/tmp", -1, 0)
	assert.Nil(pool)
}
