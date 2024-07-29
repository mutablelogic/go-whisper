package whisper_test

import (
	"context"
	"errors"
	"os"
	"testing"

	// Packages
	wav "github.com/go-audio/wav"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
	task "github.com/mutablelogic/go-whisper/pkg/whisper/task"
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

const MODEL_TINY = "ggml-tiny.en-q5_1.bin"
const SAMPLE_EN = "../../samples/jfk.wav"
const SAMPLE_FR = "../../samples/OlivierL.wav"
const SAMPLE_DE = "../../samples/ge-podcast.wav"

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
	service, err := whisper.New(t.TempDir(), whisper.OptDebug(), whisper.OptLog(func(text string) {
		t.Log(text)
	}))
	if !assert.Nil(err) {
		t.SkipNow()
	}
	defer service.Close()

	t.Run("NotFound", func(t *testing.T) {
		// Download a model - not found
		_, err = service.DownloadModel(context.Background(), "notfound.bin", nil)
		assert.ErrorIs(err, ErrNotFound)
	})

	t.Run("Download", func(t *testing.T) {
		// Download a model
		model, err := service.DownloadModel(context.Background(), MODEL_TINY, nil)
		assert.NoError(err)
		t.Log(model)
	})

	t.Run("Exists", func(t *testing.T) {
		// Get the model
		model := service.GetModelById(MODEL_TINY)
		assert.NotNil(model)
	})

	t.Run("Delete", func(t *testing.T) {
		// Delete the model
		err := service.DeleteModelById(MODEL_TINY)
		assert.NoError(err)
	})

	t.Run("NotExists", func(t *testing.T) {
		// Get the model
		model := service.GetModelById(MODEL_TINY)
		assert.Nil(model)
	})
}

func Test_whisper_003(t *testing.T) {
	assert := assert.New(t)
	service, err := whisper.New(t.TempDir(), whisper.OptDebug(), whisper.OptLog(func(text string) {
		t.Log(text)
	}))
	if !assert.Nil(err) {
		t.SkipNow()
	}
	defer service.Close()

	t.Run("Download", func(t *testing.T) {
		// Download a model
		model, err := service.DownloadModel(context.Background(), MODEL_TINY, nil)
		assert.NoError(err)
		t.Log(model)
	})

	t.Run("WithModelContext1", func(t *testing.T) {
		model := service.GetModelById(MODEL_TINY)
		assert.NotNil(model)

		// Get the model for the first time
		assert.NoError(service.WithModel(model, func(ctx *task.Context) error {
			assert.NotNil(ctx)
			return nil
		}))
	})

	t.Run("WithModelContext2", func(t *testing.T) {
		model := service.GetModelById(MODEL_TINY)
		assert.NotNil(model)

		// Get the model for the second time
		assert.NoError(service.WithModel(model, func(ctx *task.Context) error {
			assert.NotNil(ctx)
			return nil
		}))
	})

	t.Run("WithModelContext3", func(t *testing.T) {
		model := service.GetModelById(MODEL_TINY)
		assert.NotNil(model)

		// Get the model for the third time
		assert.NoError(service.WithModel(model, func(ctx *task.Context) error {
			assert.NotNil(ctx)
			return nil
		}))
	})
}

func Test_whisper_004(t *testing.T) {
	assert := assert.New(t)
	service, err := whisper.New(t.TempDir(), whisper.OptMaxConcurrent(2), whisper.OptDebug(), whisper.OptLog(func(text string) {
		t.Log(text)
	}))
	if !assert.Nil(err) {
		t.SkipNow()
	}
	defer service.Close()

	t.Run("Download", func(t *testing.T) {
		// Download a model
		model, err := service.DownloadModel(context.Background(), MODEL_TINY, nil)
		assert.NoError(err)
		t.Log(model)
	})

	var blocked int
	t.Run("InParallel", func(t *testing.T) {
		t.Run("WithModelContext1", func(t *testing.T) {
			t.Parallel()
			model := service.GetModelById(MODEL_TINY)
			assert.NotNil(model)

			err := service.WithModel(model, func(ctx *task.Context) error {
				assert.NotNil(ctx)
				return nil
			})
			if errors.Is(err, ErrChannelBlocked) {
				blocked++
			} else {
				assert.NoError(err)
			}
		})

		t.Run("WithModelContext2", func(t *testing.T) {
			t.Parallel()
			model := service.GetModelById(MODEL_TINY)
			assert.NotNil(model)

			err := service.WithModel(model, func(ctx *task.Context) error {
				assert.NotNil(ctx)
				return nil
			})
			if errors.Is(err, ErrChannelBlocked) {
				blocked++
			} else {
				assert.NoError(err)
			}
		})

		t.Run("WithModelContext3", func(t *testing.T) {
			t.Parallel()
			model := service.GetModelById(MODEL_TINY)
			assert.NotNil(model)

			err := service.WithModel(model, func(ctx *task.Context) error {
				assert.NotNil(ctx)
				return nil
			})
			if errors.Is(err, ErrChannelBlocked) {
				blocked++
			} else {
				assert.NoError(err)
			}
		})
	})

	// One of these should have been blocked
	assert.Equal(1, blocked)
}

func Test_whisper_005(t *testing.T) {
	assert := assert.New(t)
	service, err := whisper.New(t.TempDir(), whisper.OptMaxConcurrent(1), whisper.OptDebug(), whisper.OptLog(func(text string) {
		t.Log(text)
	}))
	if !assert.Nil(err) {
		t.SkipNow()
	}
	defer service.Close()

	t.Run("Download", func(t *testing.T) {
		// Download a model
		model, err := service.DownloadModel(context.Background(), MODEL_TINY, nil)
		assert.NoError(err)
		t.Log(model)
	})

	t.Run("TranscribeEN", func(t *testing.T) {
		model := service.GetModelById(MODEL_TINY)
		assert.NotNil(model)

		// Read samples
		samples, err := LoadSamples(SAMPLE_EN)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		assert.NoError(service.WithModel(model, func(task *task.Context) error {
			t.Log("Transcribing", len(samples), "samples")
			return task.Transcribe(context.Background(), samples)
		}))
	})

	t.Run("TranscribeDE", func(t *testing.T) {
		model := service.GetModelById(MODEL_TINY)
		assert.NotNil(model)

		// Read samples
		samples, err := LoadSamples(SAMPLE_DE)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		assert.NoError(service.WithModel(model, func(task *task.Context) error {
			t.Log("Transcribing", len(samples), "samples")
			return task.Transcribe(context.Background(), samples)
		}))
	})

	t.Run("TranscribeFR", func(t *testing.T) {
		model := service.GetModelById(MODEL_TINY)
		assert.NotNil(model)

		// Read samples
		samples, err := LoadSamples(SAMPLE_FR)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		assert.NoError(service.WithModel(model, func(task *task.Context) error {
			t.Log("Transcribing", len(samples), "samples")
			return task.Transcribe(context.Background(), samples)
		}))
	})
}

func Test_whisper_006(t *testing.T) {
	assert := assert.New(t)
	service, err := whisper.New(t.TempDir(), whisper.OptNoGPU(), whisper.OptMaxConcurrent(3), whisper.OptLog(func(text string) {
		t.Log(text)
	}))
	if !assert.Nil(err) {
		t.SkipNow()
	}

	// Run after all the other tests
	t.Cleanup(func() {
		service.Close()
	})

	// Download a model
	model, err := service.DownloadModel(context.Background(), MODEL_TINY, nil)
	assert.NoError(err)

	t.Run("Parallel", func(t *testing.T) {

		t.Run("TranscribeEN", func(t *testing.T) {
			t.Parallel()
			t.Log(service)

			// Read samples
			samples, err := LoadSamples(SAMPLE_EN)
			if !assert.NoError(err) {
				t.SkipNow()
			}

			assert.NoError(service.WithModel(model, func(task *task.Context) error {
				t.Log("Transcribing", len(samples), "samples")
				return task.Transcribe(context.Background(), samples)
			}))
		})

		t.Run("TranscribeDE", func(t *testing.T) {
			t.Parallel()

			model := service.GetModelById(MODEL_TINY)
			assert.NotNil(model)

			// Read samples
			samples, err := LoadSamples(SAMPLE_DE)
			if !assert.NoError(err) {
				t.SkipNow()
			}

			assert.NoError(service.WithModel(model, func(task *task.Context) error {
				t.Log("Transcribing", len(samples), "samples")
				return task.Transcribe(context.Background(), samples)
			}))
		})

		t.Run("TranscribeFR", func(t *testing.T) {
			t.Parallel()

			model := service.GetModelById(MODEL_TINY)
			assert.NotNil(model)

			// Read samples
			samples, err := LoadSamples(SAMPLE_FR)
			if !assert.NoError(err) {
				t.SkipNow()
			}

			assert.NoError(service.WithModel(model, func(task *task.Context) error {
				t.Log("Transcribing", len(samples), "samples")
				return task.Transcribe(context.Background(), samples)
			}))
		})
	})
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
