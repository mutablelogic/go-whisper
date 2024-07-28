package whisper_test

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/mutablelogic/go-whisper/sys/whisper"
	"github.com/stretchr/testify/assert"
)

const MODEL_URL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/?download=true"
const MODEL_TINY = "ggml-tiny-q5_1.bin"
const MODEL_MEDIUM = "ggml-medium-q5_0.bin" // approx 540MB

func Test_model_001(t *testing.T) {
	assert := assert.New(t)
	client := whisper.NewClient(MODEL_URL)
	assert.NotNil(client)
}

func Test_model_002(t *testing.T) {
	assert := assert.New(t)
	client := whisper.NewClient(MODEL_URL)
	assert.NotNil(client)

	// Basic writer
	w := &writer{t: t}
	_, err := client.Get(context.Background(), w, MODEL_TINY)
	assert.NoError(err)
}

func Test_model_003(t *testing.T) {
	assert := assert.New(t)
	client := whisper.NewClient(MODEL_URL)
	assert.NotNil(client)

	// Cancel download after 1s
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	w := &writer{t: t}
	_, err := client.Get(ctx, w, MODEL_MEDIUM)
	assert.ErrorIs(context.DeadlineExceeded, err)
}

////////////////////////////////////////////////////////////////////////////////
// WRITER

type writer struct {
	t          *testing.T
	curBytes   uint64
	totalBytes uint64
}

// Collect number of bytes written
func (w *writer) Write(p []byte) (n int, err error) {
	w.curBytes += uint64(len(p))
	w.t.Log("Written", w.curBytes, " bytes of", w.totalBytes)
	return len(p), nil
}

// Collect total number of bytes
func (w *writer) Header(h http.Header) error {
	contentLength := h.Get("Content-Length")
	if contentLength != "" {
		v, err := strconv.ParseUint(contentLength, 10, 64)
		if err != nil {
			return err
		}
		w.totalBytes = v
	}
	return nil
}
