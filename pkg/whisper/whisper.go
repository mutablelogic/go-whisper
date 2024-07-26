package whisper

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
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

type Transcription struct {
	Task     string                 `json:"task,omitempty"`
	Language string                 `json:"language,omitempty"`
	Duration float64                `json:"duration,omitempty"`
	Text     string                 `json:"text"`
	Segments []TranscriptionSegment `json:"segments,omitempty"`
}

type TranscriptionSegment struct {
	Id    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// This is the extension of the model files
	extModel = ".bin"

	// This is where the model is downloaded from
	modelUrl = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/?download=true"
)

const (
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

	// Close the modelpools
	for _, pool := range w.pool {
		result = errors.Join(result, pool.Close())
	}

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (w *Whisper) ListModels() []*Model {
	return w.models.models
}

func (w *Whisper) GetModelById(id string) *Model {
	return w.models.ById(id)
}

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

func (w *Whisper) DownloadModel(ctx context.Context, name, dest string) (*Model, error) {
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

	// Download the model
	// TODO: track progress
	if err := w.client.Get(ctx, f, name); err != nil {
		return nil, err
	}
	f.Close()

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

// Run with a context
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

	log.Println("Pool connections for model", model.Id, ":", pool.N())

	// Execute the function
	return fn(ctx)
}

// Transcribe decoded samples with a context
func (w *Whisper) Transcribe(ctx *Context, samples []float32) (*Transcription, error) {
	var result Transcription

	// Set parameters for transcription
	ctx.SetTranslate(false)
	ctx.SetLanguage("")

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
