package segmenter_test

import (
	"context"
	"os"
	"testing"
	"time"

	// Packages
	segmenter "github.com/mutablelogic/go-whisper/pkg/whisper/segmenter"
	assert "github.com/stretchr/testify/assert"
)

const SAMPLE = "../../../samples/OlivierL.wav"

func Test_segmenter_001(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open(SAMPLE)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	segmenter, err := segmenter.NewSegmenter(f, 200*time.Millisecond, 16000)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer segmenter.Close()

	assert.NoError(segmenter.Decode(context.Background(), func(ts time.Duration, buf []float32) {
		t.Log(ts, len(buf))
	}))
}

func Test_segmenter_002(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open(SAMPLE)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// No segmentation, just output the audio
	segmenter, err := segmenter.NewSegmenter(f, 0, 16000)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer segmenter.Close()

	assert.NoError(segmenter.Decode(context.Background(), func(ts time.Duration, buf []float32) {
		t.Log(ts, len(buf))
	}))
}
