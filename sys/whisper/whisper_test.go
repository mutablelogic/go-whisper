package whisper_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-audio/wav"
	"github.com/mutablelogic/go-whisper/sys/whisper"
	"github.com/stretchr/testify/assert"
)

const SAMPLE_JFK = "../../samples/jfk.wav"

func Test_whisper_001(t *testing.T) {
	assert := assert.New(t)

	// Set logging
	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		t.Log(level, strings.TrimSpace(text))
	})

	ctx := whisper.Whisper_init("path")
	assert.Nil(ctx)
	ctx.Whisper_free()
}

func Test_whisper_002(t *testing.T) {
	assert := assert.New(t)

	// Set logging
	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		t.Log(level, strings.TrimSpace(text))
	})

	// Create a client for downloading the model
	client := whisper.NewClient("https://huggingface.co/ggerganov/whisper.cpp/resolve/main/?download=true")
	assert.NotNil(client)

	// Get a model - save to file
	tmpfile, err := os.Create(filepath.Join(t.TempDir(), "ggml-tiny-q5_1.bin"))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer tmpfile.Close()

	// Get the model
	t.Log("Downloading model", tmpfile.Name())
	err = client.Get(context.Background(), tmpfile, filepath.Base(tmpfile.Name()))
	assert.NoError(err)

	// Load the model
	t.Run("LoadModel", func(t *testing.T) {
		ctx := whisper.Whisper_init(tmpfile.Name())
		assert.NotNil(ctx)
		ctx.Whisper_free()
	})

	// Create parameters
	t.Run("CreateParams", func(t *testing.T) {
		ctx := whisper.Whisper_init(tmpfile.Name())
		assert.NotNil(ctx)
		defer ctx.Whisper_free()

		params := whisper.NewParams(whisper.SAMPLING_GREEDY)
		assert.NotNil(params)
		defer params.Close()

		t.Log(params)

	})

	// Call full encode with silence
	t.Run("FullSilence", func(t *testing.T) {
		ctx := whisper.Whisper_init(tmpfile.Name())
		assert.NotNil(ctx)
		defer ctx.Whisper_free()

		params := whisper.NewParams(whisper.SAMPLING_GREEDY)
		assert.NotNil(params)
		defer params.Close()

		// 10 seconds of silence
		samples := make([]float32, 10*whisper.SampleRate)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		err = ctx.Whisper_full(params, samples, nil, nil, nil)
		assert.NoError(err)

		// We should detect "en" in the "ggml-tiny-q5_1.bin" model
		lang := ctx.Whisper_full_lang_id()
		lang_str := whisper.Whisper_lang_str(lang)
		assert.Equal(0, lang)
		assert.Equal("en", lang_str)

		// Get segments - should be one
		n := ctx.Whisper_full_n_segments()
		assert.GreaterOrEqual(n, 1)

		// Get tokens from segments
		for i := 0; i < n; i++ {
			ntokens := ctx.Whisper_full_n_tokens(i)
			assert.GreaterOrEqual(ntokens, 1)
			text := ctx.Whisper_full_get_segment_text(i)
			t.Logf("Segment %d: %q", i, text)
			for j := 0; j < ntokens; j++ {
				token := ctx.Whisper_full_get_token_data(i, j)
				t.Logf("Token %d: %q", j, token)
			}
		}
	})

	// Call full encode with sample
	t.Run("FullSample", func(t *testing.T) {
		ctx := whisper.Whisper_init(tmpfile.Name())
		assert.NotNil(ctx)
		defer ctx.Whisper_free()

		params := whisper.NewParams(whisper.SAMPLING_GREEDY)
		assert.NotNil(params)
		defer params.Close()

		// Get samples from WAV file
		samples, err := LoadSample(SAMPLE_JFK)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		err = ctx.Whisper_full(params, samples, nil, nil, nil)
		assert.NoError(err)

		// We should detect "en" in the "ggml-tiny-q5_1.bin" model
		lang := ctx.Whisper_full_lang_id()
		lang_str := whisper.Whisper_lang_str(lang)
		assert.Equal(0, lang)
		assert.Equal("en", lang_str)

		// Get segments - should be one
		n := ctx.Whisper_full_n_segments()
		assert.GreaterOrEqual(n, 1)

		// Get tokens from segments
		for i := 0; i < n; i++ {
			ntokens := ctx.Whisper_full_n_tokens(i)
			assert.GreaterOrEqual(ntokens, 1)
			text := strings.TrimSpace(ctx.Whisper_full_get_segment_text(i))
			t.Logf("Segment %d: %q", i, text)
			assert.Contains(strings.ToLower(text), "ask not what your country can do for you")
			assert.Contains(strings.ToLower(text), "ask what you can do for your country")
		}
	})
}

// Return samples as []float32
func LoadSample(path string) ([]float32, error) {
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
