package rnnoise_test

import (
	"os"
	"testing"

	// Packages
	"github.com/go-audio/wav"
	"github.com/mutablelogic/go-whisper/sys/rnnoise"
)

const SAMPLE_EN = "../../samples/jfk.wav"
const SAMPLE_FR = "../../samples/OlivierL.wav"
const SAMPLE_DE = "../../samples/de-podcast.wav"

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
	t.Cleanup(func() {
		rnnoise.Rnnoise_destroy(state)
	})

	samples := make([]float32, rnnoise.Rnnoise_get_frame_size())
	for i := 0; i < 10; i++ {
		prob := rnnoise.Rnnoise_process_frame(state, samples)
		t.Log(prob)
	}
}

func Test_rnnoise_004(t *testing.T) {
	state, err := rnnoise.Rnnoise_create(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		rnnoise.Rnnoise_destroy(state)
	})

	samples, err := LoadSamples(SAMPLE_EN)
	if err != nil {
		t.Fatal(err)
	}
	i := 0
	frame_size := rnnoise.Rnnoise_get_frame_size()
	for {
		if i >= len(samples) {
			break
		}
		data := samples[i:min(i+frame_size, len(samples))]
		if len(data) < frame_size {
			data = append(data, make([]float32, frame_size-len(data))...)
		}
		prob := rnnoise.Rnnoise_process_frame(state, data)
		t.Log(prob)
		i += len(data)
	}

}

//////////////////////////////////////////////////////////////////////////////

// Return samples as []float32
func LoadSamples(path string) ([]float32, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	// Read samples
	d := wav.NewDecoder(fh)
	if buf, err := d.FullPCMBuffer(); err != nil {
		return nil, err
	} else {
		return buf.AsFloat32Buffer().Data, nil
	}
}
