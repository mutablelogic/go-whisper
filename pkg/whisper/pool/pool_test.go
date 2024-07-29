package pool_test

import (
	"fmt"
	"testing"

	"github.com/mutablelogic/go-whisper/pkg/whisper/pool"
)

type Item struct {
	closed bool
}

func (i *Item) Close() error {
	fmt.Println("Closing item")
	i.closed = true
	return nil
}

func Test_basepool_001(t *testing.T) {
	var pool = pool.NewPool(10, func() any {
		return &Item{}
	})

	for i := 0; i < 100; i++ {
		item, _ := pool.Get().(*Item)
		if item == nil {
			t.Log("Pool is full")
			break
		}
		t.Log("Got item", item)
		defer item.Close()
	}

	t.Log("Closing the pool")
	pool.Close()
}
