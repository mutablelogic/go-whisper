package segmenter_test

import (
	"context"
	"os"
	"testing"

	// Packages
	segmenter "github.com/mutablelogic/go-whisper/pkg/whisper/segmenter"
	assert "github.com/stretchr/testify/assert"
)

const SAMPLE_EN = "../../../samples/jfk.wav"
const SAMPLE_FR = "../../../samples/OlivierL.wav"
const SAMPLE_DE = "../../../samples/ge-podcast.wav"

func Test_segmenter_001(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open(SAMPLE_EN)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	segmenter, err := segmenter.NewSegmenter(f)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer segmenter.Close()

	assert.NoError(segmenter.Decode(context.Background()))
}
