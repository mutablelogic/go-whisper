package whisper_test

import (
	"context"
	"os"
	"testing"

	// Packages
	"github.com/go-audio/wav"
	"github.com/mutablelogic/go-whisper/pkg/whisper"
	"github.com/stretchr/testify/assert"
)

const MODEL_TINY = "ggml-tiny-q5_1.bin"
const SAMPLES_DE = "../../samples/ge-podcast.wav"

func Test_whisper_001(t *testing.T) {
	assert := assert.New(t)
	service, err := whisper.New(t.TempDir())
	if !assert.Nil(err) {
		t.SkipNow()
	}
	assert.NotNil(service)
	assert.NoError(service.Close())
}

func Test_whisper_002(t *testing.T) {
	assert := assert.New(t)
	service, err := whisper.New(t.TempDir())
	if !assert.Nil(err) {
		t.SkipNow()
	}
	assert.NotNil(service)
	defer assert.NoError(service.Close())

	// Load a model
	model, err := service.DownloadModel(context.Background(), MODEL_TINY, "", func(cur, total uint64) {
		t.Logf("Downloaded %d bytes of %d", cur, total)
	})
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(model)

	// Get the model by ID
	model2 := service.GetModelById(model.Id)
	assert.NotNil(model2)
	assert.Equal(model, model2)

	// Delete the model by ID
	assert.NoError(service.DeleteModelById(model.Id))

}

func Test_whisper_003(t *testing.T) {
	assert := assert.New(t)
	service, err := whisper.New(t.TempDir())
	if !assert.Nil(err) {
		t.SkipNow()
	}
	assert.NotNil(service)
	defer assert.NoError(service.Close())

	// Load a model
	model, err := service.DownloadModel(context.Background(), MODEL_TINY, "", func(cur, total uint64) {
		t.Logf("Downloaded %d bytes of %d", cur, total)
	})
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(model)

	// Load samples from a file
	samples, err := LoadSamples(SAMPLES_DE)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Transcribe a file
	if err := service.WithModelContext(model, func(ctx *whisper.Context) error {
		ctx.SetLanguage("de")
		ctx.SetTranslate(false)
		transcription, err := service.Transcribe(ctx, samples)
		if err != nil {
			return err
		}
		t.Log(transcription)
		return nil
	}); !assert.NoError(err) {
		t.SkipNow()
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
