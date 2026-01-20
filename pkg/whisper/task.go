package whisper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/sys/whisper"
	attribute "go.opentelemetry.io/otel/attribute"
	trace "go.opentelemetry.io/otel/trace"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Task is used for running the transcription or translation
type Task struct {
	sync.Mutex

	// Model Id and whisper context
	model   string
	whisper *whisper.Context

	// Parameters for the next transcription
	params whisper.FullParams

	// Collect the transcription
	result *schema.Transcription

	// OTEL tracer
	tracer trace.Tracer
}

// Callback for new segments during the transcription process
type NewSegmentFunc func(*schema.Segment)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewTask creates a new task object
func NewTask() *Task {
	return new(Task)
}

// Init the task
func (t *Task) Init(path string, model *schema.Model, gpu int) error {
	t.Lock()
	defer t.Unlock()

	// Check parameters
	if model == nil {
		return ErrBadParameter
	}

	// Get default parameters
	params := whisper.DefaultContextParams()

	// If gpu is -1, then disable
	// If gpu is 0, then use whatever the default is
	// If gpu is >0, then enable and set the device
	if gpu == -1 {
		params.SetUseGpu(false)
	} else if gpu > 0 {
		params.SetUseGpu(true)
		params.SetGpuDevice(gpu)
	}

	// Get a context
	ctx := whisper.Whisper_init_from_file_with_params(filepath.Join(path, model.Path), params)
	if ctx == nil {
		return ErrInternalAppError.With("whisper_init")
	}

	// Set resources
	t.whisper = ctx
	t.model = model.Id

	// Return success
	return nil
}

// Close the task and release all resources. The task
// itself can be re-used by calling Init again
func (t *Task) Close() error {
	// Do nothing if nil
	if t == nil {
		return nil
	}

	// Release resources
	if t.whisper != nil {
		whisper.Whisper_free(t.whisper)
	}
	t.whisper = nil
	t.model = ""

	// Return success
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *Task) MarshalJSON() ([]byte, error) {
	type j struct {
		Model   string             `json:"model"`
		Params  whisper.FullParams `json:"params"`
		Context string             `json:"context"`
	}
	return json.Marshal(j{
		Model:   t.model,
		Params:  t.params,
		Context: fmt.Sprintf("%p", t.whisper),
	})
}

func (t *Task) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Task has a loaded model that matches the argument
func (t *Task) Is(model *schema.Model) bool {
	if t.model == "" {
		return false
	}
	if model == nil {
		return false
	}
	return t.model == model.Id
}

// Reset task for re-use
func (t *Task) CopyParams() {
	t.params = whisper.DefaultFullParams(whisper.SAMPLING_BEAM_SEARCH)
	t.params.SetLanguage("auto")
	t.result = new(schema.Transcription)
}

// Model is multilingual and can translate
func (t *Task) CanTranslate() bool {
	return whisper.Whisper_is_multilingual(t.whisper)
}

// Transcribe samples. The samples should be 16KHz float32 samples in
// a single channel. Appends the transcription to the result, and includes
// segment data if the new segment function is not nil
func (t *Task) Transcribe(ctx context.Context, ts time.Duration, samples []float32, fn NewSegmentFunc) error {
	// Set the 'abort' function
	t.params.SetAbortCallback(t.whisper, func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	})

	// Set the new segment function
	if fn != nil {
		t.params.SetSegmentCallback(t.whisper, func(new_segments int) {
			t.result.Language = whisper.Whisper_lang_str_full(t.whisper.DefaultLangId())
			num_segments := t.whisper.NumSegments()
			offset := len(t.result.Segments)
			for i := num_segments - new_segments; i < num_segments; i++ {
				fn(newSegment(ts, int32(offset), t.whisper.Segment(i)))
			}
		})
	}

	// Perform the transcription
	if err := whisper.Whisper_full(t.whisper, t.params, samples); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		} else {
			return err
		}
	}

	// Set the task, language and duration
	if t.params.Translate() {
		t.result.Task = "translate"
	} else {
		t.result.Task = "transcribe"
	}
	t.result.Language = whisper.Whisper_lang_str_full(t.whisper.DefaultLangId())
	t.result.Duration = schema.Timestamp(float64(len(samples)) * float64(time.Second) / float64(whisper.SampleRate))

	// Remove the callbacks
	t.params.SetAbortCallback(t.whisper, nil)
	t.params.SetSegmentCallback(t.whisper, nil)

	// Append the transcription and segments (always include segments for output formats like VTT/SRT)
	t.appendResult(ts, true)

	// Return success
	return nil
}

// TranscribeReader transcribes audio from an io.Reader using the segmenter
// to automatically handle audio decoding and segmentation. This is a higher-level
// convenience function that wraps Transcribe.
//
// The reader can be any audio format supported by FFmpeg (mp3, wav, etc.).
// Audio is automatically decoded, resampled to 16kHz, and converted to mono.
//
// Parameters:
//   - ctx: Context for cancellation
//   - r: Audio source (any format supported by FFmpeg)
//   - fn: Optional callback for each segment during transcription
//   - segmenterOpts: Optional segmenter options (e.g., segmenter.WithSegmentSize(30*time.Second))
//
// Returns the final transcription result or an error.
func (t *Task) TranscribeReader(ctx context.Context, r io.Reader, fn NewSegmentFunc, segmenterOpts ...segmenter.Opt) error {
	// Create segmenter with Whisper's sample rate (16kHz)
	seg, err := segmenter.NewFromReader(r, whisper.SampleRate, segmenterOpts...)
	if err != nil {
		return err
	}
	defer seg.Close()

	// Process each audio segment - and report in a span
	err = seg.DecodeFloat32(ctx, func(start time.Duration, samples []float32) (result error) {
		end := start + time.Duration(float64(len(samples))*float64(time.Second)/float64(whisper.SampleRate))
		childctx, endfunc := otel.StartSpan(t.tracer, ctx, "whisper.TranscribeReader.Segment",
			attribute.String("start", start.String()),
			attribute.String("end", end.String()),
		)
		defer func() { endfunc(result) }()
		return t.Transcribe(childctx, start, samples, fn)
	})

	if err == io.EOF {
		err = nil
	}
	return err
}

// Set temperature for sampling
func (t *Task) SetTemperature(v float64) error {
	if v < 0 || v > 1 {
		return ErrBadParameter.Withf("temperature must be between 0 and 1, got %f", v)
	}
	t.params.SetTemperature(float32(v))
	return nil
}

// Set initial prompt tokens for the transcription
func (t *Task) SetPrompt(prompt string) error {
	t.params.SetPrompt(prompt)
	return nil
}

// Set the language. For transcription, this is the language of the
// audio samples. For translation, this is the language to translate
// to. If you set this to "auto" then the language will be detected
func (t *Task) SetLanguage(v string) error {
	if v == "" || v == "auto" {
		t.params.SetLanguage("auto")
		return nil
	}
	id := whisper.Whisper_lang_id(v)
	if id == -1 {
		return ErrBadParameter.Withf("invalid language: %q", v)
	}
	t.params.SetLanguage(v)
	return nil
}

func (t *Task) Language() string {
	return t.params.Language()
}

// Set translate to true or false
func (t *Task) SetTranslate(v bool) {
	t.params.SetTranslate(v)
}

// Return the translate flag
func (t *Task) Translate() bool {
	return t.params.Translate()
}

// Set diarize flag
func (t *Task) SetDiarize(v bool) {
	t.params.SetDiarize(v)
}

// Return the diarize flag
func (t *Task) Diarize() bool {
	return t.params.Diarize()
}

// Return the transcription result
func (t *Task) Result() *schema.Transcription {
	return t.result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

func WriteSegmentSrt(w io.Writer, seg *schema.Segment) {
	seg.WriteSRT(w, 0)
}

func WriteSegmentVtt(w io.Writer, seg *schema.Segment) {
	seg.WriteVTT(w, 0)
}

func WriteSegmentText(w io.Writer, seg *schema.Segment) {
	seg.WriteText(w)
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (t *Task) appendResult(ts time.Duration, segments bool) {
	offset := len(t.result.Segments)

	// Append text
	for i := 0; i < t.whisper.NumSegments(); i++ {
		seg := t.whisper.Segment(i)
		t.result.Text += seg.Text
	}

	// Trim leading/trailing whitespace from the full text
	t.result.Text = strings.TrimSpace(t.result.Text)

	if segments {
		// Append segments
		for i := 0; i < t.whisper.NumSegments(); i++ {
			t.result.Segments = append(t.result.Segments, newSegment(ts, int32(offset), t.whisper.Segment(i)))
		}
	}
}

func newSegment(ts time.Duration, offset int32, seg *whisper.Segment) *schema.Segment {
	tokens := make([]string, 0, len(seg.Tokens))
	for _, token := range seg.Tokens {
		if token.Text != "" {
			tokens = append(tokens, token.Text)
		}
	}
	return &schema.Segment{
		Id:          offset + seg.Id,
		Text:        seg.Text,
		Tokens:      tokens,
		Start:       schema.Timestamp(seg.T0 + ts),
		End:         schema.Timestamp(seg.T1 + ts),
		SpeakerTurn: seg.SpeakerTurn,
	}
}
