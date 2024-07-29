package whisper_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	// Packages
	"github.com/go-audio/wav"
	"github.com/mutablelogic/go-whisper/sys/whisper"
	"github.com/stretchr/testify/assert"
)

const SAMPLE_EN = "../../samples/jfk.wav"
const SAMPLE_FR = "../../samples/OlivierL.wav"
const SAMPLE_DE = "../../samples/ge-podcast.wav"

func Test_whisper_00(t *testing.T) {
	assert := assert.New(t)

	// Set logging
	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		t.Log(level, strings.TrimSpace(text))
	})

	// Create a file for the model
	w, err := os.Create(filepath.Join(t.TempDir(), MODEL_TINY))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer w.Close()

	// Read the model
	client := whisper.NewClient(MODEL_URL)
	if !assert.NotNil(client) {
		t.SkipNow()
	}
	if _, err := client.Get(context.Background(), w, MODEL_TINY); !assert.NoError(err) {
		t.SkipNow()
	}

	t.Run("InitFromFileWithParams", func(t *testing.T) {
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), whisper.DefaultContextParams())
		assert.NotNil(ctx)
		// Free memory
		whisper.Whisper_free(ctx)
	})

	t.Run("InitFromMemoryWithParams", func(t *testing.T) {
		data, err := os.ReadFile(w.Name())
		if !assert.NoError(err) {
			t.SkipNow()
		}

		ctx := whisper.Whisper_init_from_buffer_with_params(data, whisper.DefaultContextParams())
		assert.NotNil(ctx)
		// Free memory
		whisper.Whisper_free(ctx)
	})

}

func Test_whisper_01(t *testing.T) {
	assert := assert.New(t)

	// Set logging
	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		t.Log(level, strings.TrimSpace(text))
	})

	// Get maximum language id
	max := whisper.Whisper_lang_max_id()
	assert.Greater(max, 0)

	for i := 0; i < max; i++ {
		short := whisper.Whisper_lang_str(i)
		long := whisper.Whisper_lang_str_full(i)
		assert.NotEmpty(short)
		assert.NotEmpty(long)
		assert.Equal(i, whisper.Whisper_lang_id(short))
		assert.Equal(i, whisper.Whisper_lang_id(long))
		t.Logf("Language %d: %s (%s)", i, short, long)
	}
}

func Test_whisper_02(t *testing.T) {
	assert := assert.New(t)

	// Set logging
	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		t.Log(level, strings.TrimSpace(text))
	})

	// Create a file for the model
	w, err := os.Create(filepath.Join(t.TempDir(), MODEL_TINY))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer w.Close()

	// Read the model
	client := whisper.NewClient(MODEL_URL)
	if !assert.NotNil(client) {
		t.SkipNow()
	}
	if _, err := client.Get(context.Background(), w, MODEL_TINY); !assert.NoError(err) {
		t.SkipNow()
	}

	// Let's run three in parallel
	params := whisper.DefaultContextParams()
	params.SetUseGpu(false)

	t.Run("Full_en", func(t *testing.T) {
		t.Parallel()

		// Create a context
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), params)
		if !assert.NotNil(ctx) {
			t.SkipNow()
		}

		// Load samples
		data, err := LoadSamples(SAMPLE_EN)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Set parameters
		params := whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)
		params.SetLanguage("auto")
		params.SetTranslate(false)
		t.Log(params)

		// Run the model
		err = whisper.Whisper_full(ctx, params, data)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Free memory
		whisper.Whisper_free(ctx)
	})

	t.Run("Full_fr", func(t *testing.T) {
		t.Parallel()

		// Create a context
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), params)
		if !assert.NotNil(ctx) {
			t.SkipNow()
		}

		// Load samples
		data, err := LoadSamples(SAMPLE_FR)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Set parameters
		params := whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)
		params.SetLanguage("auto")
		params.SetTranslate(false)
		t.Log(params)

		// Run the model
		err = whisper.Whisper_full(ctx, params, data)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Free memory
		whisper.Whisper_free(ctx)
	})

	t.Run("Full_de", func(t *testing.T) {
		t.Parallel()

		// Create a context
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), params)
		if !assert.NotNil(ctx) {
			t.SkipNow()
		}

		// Load samples
		data, err := LoadSamples(SAMPLE_DE)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Set parameters
		params := whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)
		params.SetLanguage("auto")
		params.SetTranslate(false)
		t.Log(params)

		// Run the model
		err = whisper.Whisper_full(ctx, params, data)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Free memory
		whisper.Whisper_free(ctx)
	})

}

//////////////////////////////////////////////////////////////////////////////

func Test_whisper_03(t *testing.T) {
	assert := assert.New(t)

	// Set logging
	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		t.Log(level, strings.TrimSpace(text))
	})

	// Create a file for the model
	w, err := os.Create(filepath.Join(t.TempDir(), MODEL_TINY))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer w.Close()

	// Read the model
	client := whisper.NewClient(MODEL_URL)
	if !assert.NotNil(client) {
		t.SkipNow()
	}
	if _, err := client.Get(context.Background(), w, MODEL_TINY); !assert.NoError(err) {
		t.SkipNow()
	}

	// Let's run three in parallel
	params := whisper.DefaultContextParams()
	params.SetUseGpu(false)

	t.Run("Progress_en", func(t *testing.T) {
		t.Parallel()

		// Create a context
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), params)
		if !assert.NotNil(ctx) {
			t.SkipNow()
		}

		// Load samples
		data, err := LoadSamples(SAMPLE_EN)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Set parameters
		params := whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)
		params.SetLanguage("auto")
		params.SetTranslate(false)
		params.SetProgressCallback(ctx, func(progress int) {
			t.Logf("Progress: %d%%", progress)
		})
		params.SetAbortCallback(ctx, func() bool {
			t.Logf("Abort Callback called")
			return false
		})
		params.SetSegmentCallback(ctx, func(segment int) {
			t.Logf("Segment %d", segment)
		})

		t.Log(params)

		// Run the model
		err = whisper.Whisper_full(ctx, params, data)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Free memory
		whisper.Whisper_free(ctx)
	})

	t.Run("Progress_fr", func(t *testing.T) {
		t.Parallel()

		// Create a context
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), params)
		if !assert.NotNil(ctx) {
			t.SkipNow()
		}

		// Load samples
		data, err := LoadSamples(SAMPLE_FR)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Set parameters
		params := whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)
		params.SetLanguage("auto")
		params.SetTranslate(false)
		params.SetProgressCallback(ctx, func(progress int) {
			t.Logf("Progress: %d%%", progress)
		})
		params.SetAbortCallback(ctx, func() bool {
			t.Logf("Abort Callback called")
			return false
		})
		params.SetSegmentCallback(ctx, func(segment int) {
			t.Logf("Segment %d", segment)
		})

		t.Log(params)

		// Run the model
		err = whisper.Whisper_full(ctx, params, data)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Free memory
		whisper.Whisper_free(ctx)
	})

	t.Run("Progress_de", func(t *testing.T) {
		t.Parallel()

		// Create a context
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), params)
		if !assert.NotNil(ctx) {
			t.SkipNow()
		}

		// Load samples
		data, err := LoadSamples(SAMPLE_DE)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Set parameters
		params := whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)
		params.SetLanguage("auto")
		params.SetTranslate(false)
		params.SetProgressCallback(ctx, func(progress int) {
			t.Logf("Progress: %d%%", progress)
		})
		params.SetAbortCallback(ctx, func() bool {
			t.Logf("Abort Callback called")
			return false
		})
		params.SetSegmentCallback(ctx, func(segment int) {
			t.Logf("Segment %d", segment)
		})

		t.Log(params)

		// Run the model
		err = whisper.Whisper_full(ctx, params, data)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Free memory
		whisper.Whisper_free(ctx)
	})
}

func Test_whisper_04(t *testing.T) {
	assert := assert.New(t)

	// Set logging
	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		t.Log(level, strings.TrimSpace(text))
	})

	// Create a file for the model
	w, err := os.Create(filepath.Join(t.TempDir(), MODEL_TINY))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer w.Close()

	// Read the model
	client := whisper.NewClient(MODEL_URL)
	if !assert.NotNil(client) {
		t.SkipNow()
	}
	if _, err := client.Get(context.Background(), w, MODEL_TINY); !assert.NoError(err) {
		t.SkipNow()
	}

	// Let's run three in parallel
	params := whisper.DefaultContextParams()
	params.SetUseGpu(false)

	t.Run("Abort_fr", func(t *testing.T) {
		// Create a context
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), params)
		if !assert.NotNil(ctx) {
			t.SkipNow()
		}

		// Load samples
		data, err := LoadSamples(SAMPLE_FR)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Progress counter
		var p int

		// Set parameters
		params := whisper.DefaultFullParams(whisper.SAMPLING_BEAM_SEARCH)
		params.SetLanguage("auto")
		params.SetTranslate(false)
		params.SetProgressCallback(ctx, func(progress int) {
			t.Logf("Progress: %d%%", progress)
			p = progress
		})
		params.SetAbortCallback(ctx, func() bool {
			if p > 20 {
				t.Logf("Aborting")
				return true
			}
			return false
		})

		// Run the model
		err = whisper.Whisper_full(ctx, params, data)
		if !assert.ErrorIs(err, whisper.ErrTranscriptionFailed) {
			t.SkipNow()
		}

		// Free memory
		whisper.Whisper_free(ctx)
	})
}

func Test_whisper_05(t *testing.T) {
	assert := assert.New(t)

	// Set logging
	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		t.Log(level, strings.TrimSpace(text))
	})

	// Create a file for the model
	w, err := os.Create(filepath.Join(t.TempDir(), MODEL_TINY))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer w.Close()

	// Read the model
	client := whisper.NewClient(MODEL_URL)
	if !assert.NotNil(client) {
		t.SkipNow()
	}
	if _, err := client.Get(context.Background(), w, MODEL_TINY); !assert.NoError(err) {
		t.SkipNow()
	}

	// Let's run three in parallel
	params := whisper.DefaultContextParams()
	params.SetUseGpu(false)

	t.Run("Tokens_fr", func(t *testing.T) {
		// Create a context
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), params)
		if !assert.NotNil(ctx) {
			t.SkipNow()
		}

		// Load samples
		data, err := LoadSamples(SAMPLE_FR)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Set parameters
		params := whisper.DefaultFullParams(whisper.SAMPLING_BEAM_SEARCH)
		params.SetLanguage("auto")
		params.SetTranslate(false)
		params.SetTokenTimestamps(true)
		params.SetDiarizeEnable(true)
		params.SetSegmentCallback(ctx, func(new_segments int) {
			num_segments := ctx.NumSegments()
			for i := num_segments - new_segments; i < num_segments; i++ {
				t.Logf("Segment %d: %v", i, ctx.SegmentText(i))
			}
		})

		// Run the model
		err = whisper.Whisper_full(ctx, params, data)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Free memory
		whisper.Whisper_free(ctx)
	})

	t.Run("Tokens_de", func(t *testing.T) {
		// Create a context
		ctx := whisper.Whisper_init_from_file_with_params(w.Name(), params)
		if !assert.NotNil(ctx) {
			t.SkipNow()
		}

		// Load samples
		data, err := LoadSamples(SAMPLE_DE)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Set parameters
		params := whisper.DefaultFullParams(whisper.SAMPLING_BEAM_SEARCH)
		params.SetLanguage("auto")
		params.SetTranslate(false)
		params.SetTokenTimestamps(true)
		params.SetDiarizeEnable(true)
		params.SetSegmentCallback(ctx, func(new_segments int) {
			num_segments := ctx.NumSegments()
			for i := num_segments - new_segments; i < num_segments; i++ {
				t.Logf("Segment %d: %v", i, ctx.Segment(i))
			}
		})

		// Run the model
		err = whisper.Whisper_full(ctx, params, data)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Free memory
		whisper.Whisper_free(ctx)
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
