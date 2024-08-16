package segmenter_test

import (
	"context"
	"os"
	"testing"
	"time"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	segmenter "github.com/mutablelogic/go-whisper/pkg/segmenter"
	assert "github.com/stretchr/testify/assert"
)

func Test_denoiser_001(t *testing.T) {
	assert := assert.New(t)

	// Open sample file
	r, err := os.Open(SAMPLE)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	// Open denoiser
	denoiser, err := segmenter.NewDenoiseReader(r, "fltp", 16000)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer denoiser.Close()

	// Start timestamp for segment
	start := time.Duration(-1)
	silence_dur := time.Second
	silence_threshold := float32(0.5)
	assert.NoError(denoiser.Decode(context.Background(), func(ts time.Duration, p float32, frame *ffmpeg.Frame) error {
		// Determine if we should start a new segment
		if p > silence_threshold {
			if start == -1 {
				t.Log("Start new segment")
			}
			start = ts
		} else if ts-start > silence_dur && start != -1 {
			start = -1
			t.Log("End segment")
		}

		// Write segment
		if start >= 0 {
			t.Log(p, "=>", frame)
		}

		// Return success
		return nil
	}))
}
