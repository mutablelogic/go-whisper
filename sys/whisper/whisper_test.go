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
const SAMPLE_OLIVIERL = "../../samples/OlivierL.wav"
const MODEL_URL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/?download=true"
const MODEL_TINY = "ggml-tiny-q5_1.bin"
const MODEL_MEDIUM = "ggml-medium-q5_0.bin" // approx 540MB

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
	client := whisper.NewClient(MODEL_URL)
	assert.NotNil(client)

	// Get a model - save to file
	tmpfile, err := os.Create(filepath.Join(t.TempDir(), MODEL_TINY))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer tmpfile.Close()

	// Get the model
	t.Log("Downloading model", tmpfile.Name())
	_, err = client.Get(context.Background(), tmpfile, filepath.Base(tmpfile.Name()))
	assert.NoError(err)

	// Call full encode with callback
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

	// Call full encode with JFK sample
	t.Run("FullSampleJFK", func(t *testing.T) {
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

	// Call full encode with SAMPLE_OLIVIERL (fr)
	t.Run("FullSampleOlivierL", func(t *testing.T) {
		ctx := whisper.Whisper_init(tmpfile.Name())
		assert.NotNil(ctx)
		defer ctx.Whisper_free()

		params := whisper.NewParams(whisper.SAMPLING_GREEDY)
		assert.NotNil(params)
		defer params.Close()

		params.SetLanguage(ctx.Whisper_lang_id("fr"))

		// Get samples from WAV file
		samples, err := LoadSample(SAMPLE_OLIVIERL)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		err = ctx.Whisper_full(params, samples, nil, nil, nil)
		assert.NoError(err)

		// We should detect "en" in the "ggml-tiny-q5_1.bin" model
		lang := ctx.Whisper_full_lang_id()
		lang_str := whisper.Whisper_lang_str(lang)
		assert.Equal(ctx.Whisper_lang_id("fr"), lang)
		assert.Equal("fr", lang_str)

		// Get segments - should be one
		n := ctx.Whisper_full_n_segments()
		assert.GreaterOrEqual(n, 1)

		// Get tokens from segments
		for i := 0; i < n; i++ {
			ntokens := ctx.Whisper_full_n_tokens(i)
			assert.GreaterOrEqual(ntokens, 1)
			text := strings.TrimSpace(ctx.Whisper_full_get_segment_text(i))
			t.Logf("Segment %d: %q", i, text)
		}
	})
}

func Test_whisper_003(t *testing.T) {
	assert := assert.New(t)

	// Set logging
	whisper.Whisper_log_set(func(level whisper.LogLevel, text string) {
		t.Log(level, strings.TrimSpace(text))
	})

	// Create a client for downloading the model
	client := whisper.NewClient(MODEL_URL)
	assert.NotNil(client)

	// Get a model - save to file
	tmpfile, err := os.Create(filepath.Join(t.TempDir(), MODEL_TINY))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer tmpfile.Close()

	// Get the model
	t.Log("Downloading model", tmpfile.Name())
	_, err = client.Get(context.Background(), tmpfile, filepath.Base(tmpfile.Name()))
	assert.NoError(err)

	// Call full encode with callbacks SAMPLE_OLIVIERL (fr)
	t.Run("FullCallbackSampleOlivierL", func(t *testing.T) {
		ctx := whisper.Whisper_init(tmpfile.Name())
		assert.NotNil(ctx)
		defer ctx.Whisper_free()

		params := whisper.NewParams(whisper.SAMPLING_GREEDY)
		assert.NotNil(params)
		defer params.Close()

		params.SetLanguage(ctx.Whisper_lang_id("fr"))

		// Get samples from WAV file
		samples, err := LoadSample(SAMPLE_OLIVIERL)
		if !assert.NoError(err) {
			t.SkipNow()
		}

		// Set callbacks
		var calledBegin, calledSegment, calledProgress bool
		encodeBegin := func() bool {
			t.Log("Called encodeBegin")
			calledBegin = true
			return true
		}
		newSegment := func(s int) {
			t.Logf("Called newSegment %d", s)
			calledSegment = true
		}
		progress := func(p int) {
			t.Logf("Called progress %d", p)
			calledProgress = true
		}

		err = ctx.Whisper_full(params, samples, encodeBegin, newSegment, progress)
		assert.NoError(err)
		assert.True(calledBegin)
		assert.True(calledSegment)
		assert.True(calledProgress)

		// We should detect "fr" in the "ggml-tiny-q5_1.bin" model
		lang := ctx.Whisper_full_lang_id()
		lang_str := whisper.Whisper_lang_str(lang)
		assert.Equal(ctx.Whisper_lang_id("fr"), lang)
		assert.Equal("fr", lang_str)

		// Get segments - should be one
		n := ctx.Whisper_full_n_segments()
		assert.GreaterOrEqual(n, 1)

		// Get tokens from segments
		for i := 0; i < n; i++ {
			ntokens := ctx.Whisper_full_n_tokens(i)
			assert.GreaterOrEqual(ntokens, 1)
			text := strings.TrimSpace(ctx.Whisper_full_get_segment_text(i))
			t.Logf("Segment %d: %q", i, text)
		}
	})
}

//////////////////////////////////////////////////////////////////////////////

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
