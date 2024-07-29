package whisper

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// Packages
	"github.com/mutablelogic/go-whisper/sys/whisper"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Whisper struct {
	models models
	client whisper.Client
	pool   map[string]*ModelPool
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// This is the extension of the model files
	extModel = ".bin"

	// This is where the model is downloaded from
	modelUrl = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/?download=true"

	// This is the expected sample rate for transcription and translation (16KHz)
	SampleRate = whisper.SampleRate
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new whisper instance with the specified path to the models directory
func New(models string) (*Whisper, error) {
	// Check model path exists and is writable
	if info, err := os.Stat(models); err != nil {
		return nil, os.ErrNotExist
	} else if !info.IsDir() {
		return nil, os.ErrNotExist
	}

	// Create whisper instance
	w := new(Whisper)
	w.client = whisper.NewClient(modelUrl)
	w.pool = make(map[string]*ModelPool)

	// Get a listing of the models
	w.models.path = models
	if err := w.models.Rescan(); err != nil {
		return nil, err
	}

	// Return success
	return w, nil
}

// Close the whisper instance and release resources
func (w *Whisper) Close() error {
	var result error

	// Close the model pools
	for _, pool := range w.pool {
		result = errors.Join(result, pool.Close())
	}

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all models in the models directory
func (w *Whisper) ListModels() []*Model {
	return w.models.models
}

// Get a model by its Id, returns nil if the model does not exist
func (w *Whisper) GetModelById(id string) *Model {
	return w.models.ById(id)
}

// Delete a model by its id
func (w *Whisper) DeleteModelById(id string) error {
	model := w.models.ById(id)
	if model == nil {
		return os.ErrNotExist
	}

	// Delete the model pool
	if pool, ok := w.pool[model.Id]; ok {
		pool.Close()
		delete(w.pool, model.Id)
	}

	// Delete the model
	path := filepath.Join(w.models.path, model.Path)
	if err := os.Remove(path); err != nil {
		return err
	}

	// Rescan the models directory
	if err := w.models.Rescan(); err != nil {
		return err
	}

	// Return success
	return nil
}

// Download a model to the models directory. If the model already exists, it will be returned.
// The destination directory is relative to the models directory. A function can be provided to
// track the progress of the download. If no Content-Length is provided by the server, the total
// bytes will be unknown and is set to zero.
func (w *Whisper) DownloadModel(ctx context.Context, name, dest string, fn func(curBytes, totalBytes uint64)) (*Model, error) {
	// If the model already exists, return it
	model := w.models.ByPath(name, dest)
	if model != nil {
		return model, nil
	}

	// Create the destination directory if it's not empty or a '.'
	dest_ := filepath.Join(w.models.path, dest)
	if info, err := os.Stat(dest_); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dest_, 0755); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if !info.IsDir() {
		return nil, os.ErrNotExist
	}

	// Create the destination file
	f, err := os.Create(filepath.Join(dest_, name))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Download the model, with callback. If an error occurs, the model is deleted again
	if _, err := w.client.Get(ctx, &writer{Writer: f, fn: fn}, name); err != nil {
		return nil, errors.Join(err, os.Remove(f.Name()))
	}

	// Rescan the models directory
	if err := w.models.Rescan(); err != nil {
		return nil, err
	}

	// Get a model by path
	model = w.models.ByPath(name, dest)
	if model == nil {
		return nil, os.ErrNotExist
	}

	// Return success
	return model, nil
}

// Get a context for the specified model, which may load the model or return an existing one.
// The context can then be used to run the Transcribe function. Returns os.ErrNotExist error
// if the model does not exist, or if the context could not be created.
func (w *Whisper) WithModelContext(model *Model, fn func(ctx *Context) error) error {
	// Check model parameter
	if model == nil {
		return os.ErrNotExist
	}

	// Create a new pool for the model if it doesn't exist
	if _, ok := w.pool[model.Id]; !ok {
		w.pool[model.Id] = NewModelPool(w.models.path, model)
	}

	// Retrieve the pool
	pool := w.pool[model.Id]
	if pool == nil {
		return os.ErrNotExist
	}

	// Get a context from the pool and return later
	ctx := pool.Get()
	if ctx == nil {
		return os.ErrNotExist
	}
	defer pool.Put(ctx)

	// Execute the function
	return fn(ctx)
}

// Transcribe samples with a context. The samples should be 16KHz float32 samples in
// a single channel. TODO: We need a low-latency streaming version of this function.
// TODO: We need a callback for segment progress.
func (w *Whisper) Transcribe(ctx *Context, samples []float32) (*Transcription, error) {
	var result Transcription

	// Perform the transcription
	log.Print(ctx.Params())
	if err := ctx.Whisper_full(ctx.Params(), samples, nil, nil, nil); err != nil {
		return nil, err
	}

	// Set language and duration
	result.Language = whisper.Whisper_lang_str(ctx.Whisper_full_lang_id())
	result.Duration = float64(len(samples)) / float64(SampleRate)

	// Set segments
	for i := 0; i < ctx.Whisper_full_n_segments(); i++ {
		segment := TranscriptionSegment{
			Id:    i,
			Text:  strings.TrimSpace(ctx.Whisper_full_get_segment_text(i)),
			Start: float64(ctx.Whisper_full_get_segment_t0(i)) / 100.0,
			End:   float64(ctx.Whisper_full_get_segment_t1(i)) / 100.0,
		}
		if i > 0 {
			result.Text += " "
		}
		result.Text += segment.Text
		result.Segments = append(result.Segments, segment)
	}

	// Return the result
	return &result, nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

type writer struct {
	io.Writer

	// Current and total bytes
	curBytes, totalBytes uint64

	// Callback function
	fn func(curBytes, totalBytes uint64)
}

// Collect number of bytes written
func (w *writer) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if err == nil && w.fn != nil {
		w.curBytes += uint64(n)
		w.fn(w.curBytes, w.totalBytes)
	}
	return n, nil
}

// Collect total number of bytes
func (w *writer) Header(h http.Header) error {
	if contentLength := h.Get("Content-Length"); contentLength != "" {
		if v, err := strconv.ParseUint(contentLength, 10, 64); err != nil {
			return err
		} else {
			w.totalBytes = v
		}
	}
	return nil
}
