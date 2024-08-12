package rnnoise_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-whisper/sys/rnnoise"
)

func Test_rnnoise_001(t *testing.T) {
	size := rnnoise.Rnnoise_get_size()
	t.Log("Size:", size)
}

func Test_rnnoise_002(t *testing.T) {
	size := rnnoise.Rnnoise_get_frame_size()
	t.Log("Frame size:", size)
}

func Test_rnnoise_003(t *testing.T) {
	state, err := rnnoise.Rnnoise_create(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rnnoise.Rnnoise_destroy(state)
	if err := rnnoise.Rnnoise_init(state, nil); err != nil {
		t.Fatal(err)
	}
	samples := make([]float32, rnnoise.Rnnoise_get_frame_size())

	for i := 0; i < 10; i++ {
		prob := rnnoise.Rnnoise_process_frame(state, samples, samples)
		t.Log(prob)
	}
}
