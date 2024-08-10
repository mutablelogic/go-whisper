package pool_test

import (
	"sync"
	"testing"

	// Packages
	"github.com/mutablelogic/go-whisper/pkg/pool"
)

type Item struct {
	*testing.T
	closed bool
}

func (i *Item) Close() error {
	i.Log("Closing item")
	i.closed = true
	return nil
}

func Test_basepool_001(t *testing.T) {
	t.Log("Creating pool with max=10")
	var pool = pool.NewPool(10, func() any {
		return &Item{t, false}
	})

	var items []*Item
	for i := 0; i < 100; i++ {
		t.Log("Getting pool", i)
		item, _ := pool.Get().(*Item)
		if item == nil {
			t.Log("Pool is full")
			break
		}
		t.Log("Got item", item)
		items = append(items, item)
	}

	for i := 0; i < len(items); i++ {
		t.Log("Putting item", i)
		pool.Put(items[i])
	}

	t.Log("Closing the pool")
	pool.Close()
}

func Test_basepool_002(t *testing.T) {
	t.Log("Creating pool with max=100")
	var pool = pool.NewPool(100, func() any {
		return &Item{t, false}
	})

	var wg sync.WaitGroup
	for i := 0; i < 120; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			t.Log("Getting pool", i)
			item, _ := pool.Get().(*Item)
			if item == nil {
				t.Log("Pool is full")
			}
			t.Log("Got item", item)
			defer pool.Put(item)
		}(i)
	}

	wg.Wait()

	t.Log("Closing the pool")
	pool.Close()
}
