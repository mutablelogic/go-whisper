package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

// Helper function to create a fake model file (must be at least 8MB for store to recognize it)
func createFakeModel(t *testing.T, path string) {
	data := make([]byte, 8*1024*1024) // 8MB
	err := os.WriteFile(path, data, 0644)
	assert.NoError(t, err)
}

func TestStore_NewStore(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)
	assert.NotNil(store)
	assert.Equal(path, store.path)
	assert.Equal(".bin", store.ext)
}

func TestStore_NewStore_InvalidPath(t *testing.T) {
	assert := assert.New(t)

	store, err := NewStore("/nonexistent/path", ".bin", "https://example.com")
	assert.Error(err)
	assert.Nil(store)
}

func TestStore_NewStore_NotDirectory(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	// Create a file, not a directory
	filePath := filepath.Join(path, "notdir")
	err := os.WriteFile(filePath, []byte("test"), 0644)
	assert.NoError(err)

	store, err := NewStore(filePath, ".bin", "https://example.com")
	assert.Error(err)
	assert.Nil(store)
}

func TestStore_List_Empty(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	models := store.List()
	assert.Empty(models)
}

func TestStore_List_WithModels(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	// Create a fake model file
	modelPath := filepath.Join(path, "test-model.bin")
	createFakeModel(t, modelPath)

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	models := store.List()
	assert.Len(models, 1)
	assert.Equal("test-model", models[0].Id)
	assert.Equal("test-model.bin", models[0].Path)
}

func TestStore_ById(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	// Create fake model files
	modelPath1 := filepath.Join(path, "model1.bin")
	modelPath2 := filepath.Join(path, "model2.bin")
	createFakeModel(t, modelPath1)
	createFakeModel(t, modelPath2)

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	// Find existing model
	model := store.ById("model1")
	assert.NotNil(model)
	assert.Equal("model1", model.Id)

	// Find non-existent model
	model = store.ById("nonexistent")
	assert.Nil(model)
}

func TestStore_Delete(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	// Create a fake model file
	modelPath := filepath.Join(path, "test-model.bin")
	createFakeModel(t, modelPath)

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	// Verify model exists
	models := store.List()
	assert.Len(models, 1)

	// Delete the model
	err = store.Delete("test-model")
	assert.NoError(err)

	// Verify model is gone
	models = store.List()
	assert.Empty(models)

	// Verify file is deleted
	_, err = os.Stat(modelPath)
	assert.True(os.IsNotExist(err))
}

func TestStore_Delete_NonExistent(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	err = store.Delete("nonexistent")
	assert.Error(err)
}

func TestStore_Rescan(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)
	assert.Empty(store.List())

	// Add a model file after creation
	modelPath := filepath.Join(path, "new-model.bin")
	createFakeModel(t, modelPath)

	// Rescan
	err = store.Rescan()
	assert.NoError(err)

	// Verify new model is found
	models := store.List()
	assert.Len(models, 1)
	assert.Equal("new-model", models[0].Id)
}

func TestStore_MarshalJSON(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	// Create fake model files
	modelPath := filepath.Join(path, "test-model.bin")
	createFakeModel(t, modelPath)

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	data, err := store.MarshalJSON()
	assert.NoError(err)
	assert.NotEmpty(data)
	assert.Contains(string(data), "test-model")
	assert.Contains(string(data), ".bin")
}

func TestStore_String(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	str := store.String()
	assert.NotEmpty(str)
	assert.Contains(str, path)
	assert.Contains(str, ".bin")
}

func TestStore_Download_InvalidPath(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Try to download with invalid path
	model, err := store.Download(ctx, "", nil)
	assert.Error(err)
	assert.Nil(model)
}

func TestStore_Download_AlreadyExists(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	// Create an existing model
	modelPath := filepath.Join(path, "existing-model.bin")
	createFakeModel(t, modelPath)

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	ctx := context.Background()

	// Try to download a model that already exists
	model, err := store.Download(ctx, "existing-model.bin", nil)
	assert.NoError(err)
	assert.NotNil(model)
	assert.Equal("existing-model", model.Id)
}

func TestStore_WithNestedDirectories(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	// Create nested directory structure
	nestedPath := filepath.Join(path, "subdir")
	err := os.MkdirAll(nestedPath, 0755)
	assert.NoError(err)

	// Create model in nested directory
	modelPath := filepath.Join(nestedPath, "nested-model.bin")
	createFakeModel(t, modelPath)

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	// Should find nested model
	models := store.List()
	assert.Len(models, 1)
	assert.Equal("nested-model", models[0].Id)
	assert.Equal("subdir/nested-model.bin", models[0].Path)

	// Model ID is the filename without extension (not including directory)
	model := store.ById("nested-model")
	assert.NotNil(model)
}

func TestStore_MultipleExtensions(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	// Create files with different extensions
	createFakeModel(t, filepath.Join(path, "model1.bin"))
	createFakeModel(t, filepath.Join(path, "model2.txt"))
	createFakeModel(t, filepath.Join(path, "model3.bin"))

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	// Should only find .bin files
	models := store.List()
	assert.Len(models, 2)
}

func TestStore_ConcurrentAccess(t *testing.T) {
	assert := assert.New(t)
	path := t.TempDir()

	// Create a model
	modelPath := filepath.Join(path, "test-model.bin")
	createFakeModel(t, modelPath)

	store, err := NewStore(path, ".bin", "https://example.com")
	assert.NoError(err)

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			models := store.List()
			assert.Len(models, 1)
			model := store.ById("test-model")
			assert.NotNil(model)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
