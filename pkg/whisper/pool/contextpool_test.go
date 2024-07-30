package pool_test

import (
	"testing"

	// Packages
	pool "github.com/mutablelogic/go-whisper/pkg/whisper/pool"
	schema "github.com/mutablelogic/go-whisper/pkg/whisper/schema"
)

func Test_contextpool_001(t *testing.T) {
	var pool = pool.NewContextPool(t.TempDir(), 2, 0)

	model1, err := pool.Get(&schema.Model{
		Id: "model1",
	})
	if err != nil {
		t.Error(err)
	}
	if model1 == nil {
		t.Error("Expected model1")
	}
	t.Log("Got model1", model1)

	model2, err := pool.Get(&schema.Model{
		Id: "model2",
	})
	if err != nil {
		t.Error(err)
	}
	if model2 == nil {
		t.Error("Expected model2")
	}
	t.Log("Got model2", model2)

	pool.Put(model1)

	model3, err := pool.Get(&schema.Model{
		Id: "model1",
	})
	if err != nil {
		t.Error(err)
	}
	if model3 == nil {
		t.Error("Expected model3")
	}
	t.Log("Got model3", model3)

	pool.Put(model2)
	pool.Put(model3)

	t.Log("Closing the pool")
	pool.Close()
}
