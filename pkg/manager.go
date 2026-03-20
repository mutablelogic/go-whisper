package pkg

import (
	"context"
	"errors"
	"io"
	"slices"
	"strings"

	// Packages
	goclient "github.com/mutablelogic/go-client"
	multipart "github.com/mutablelogic/go-client/pkg/multipart"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
	elevenlabs "github.com/mutablelogic/go-whisper/pkg/elevenlabs"
	openai "github.com/mutablelogic/go-whisper/pkg/openai"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
	attribute "go.opentelemetry.io/otel/attribute"
	trace "go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	whisper    *whisper.Manager
	elevenlabs *elevenlabs.Client
	openai     *openai.Client
	segopts    []segmenter.Opt
	tracer     trace.Tracer
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new manager with a whisper client and optionally
// elevenlabs and openai clients if API keys are provided via Opts
func New(modelsPath string, opt ...Opt) (*Manager, error) {
	var o opts
	self := new(Manager)

	// Apply options
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}

	// Create whisper client (required - never nil)
	whisperManager, err := whisper.New(modelsPath, o.whisperopts...)
	if err != nil {
		return nil, err
	} else {
		self.whisper = whisperManager
	}

	// Store segmenter options
	self.segopts = o.segOpts

	// Add tracer if provided
	if o.tracer != nil {
		self.tracer = o.tracer
		o.clientOpts = append(o.clientOpts, goclient.OptTracer(o.tracer))
	}

	// Create elevenlabs client if API key provided
	if o.elevenLabsKey != "" {
		if client, err := elevenlabs.New(o.elevenLabsKey, o.clientOpts...); err != nil {
			return nil, errors.Join(err, whisperManager.Close())
		} else {
			self.elevenlabs = client
		}
	}

	// Create openai client if API key provided
	if o.openAIKey != "" {
		if client, err := openai.New(o.openAIKey, o.clientOpts...); err != nil {
			return nil, errors.Join(err, whisperManager.Close())
		} else {
			self.openai = client
		}
	}

	return self, nil
}

// Close closes the manager and all clients
func (m *Manager) Close() error {
	var result error
	if m.whisper != nil {
		result = errors.Join(result, m.whisper.Close())
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListModels returns all available models from whisper, openai, and elevenlabs
func (m *Manager) ListModels(ctx context.Context) []*schema.Model {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "manager.ListModels")
	defer func() { endSpan(nil) }()

	models := make([]*schema.Model, 0)

	// Add whisper models
	if m.whisper != nil {
		if whisperModels := m.whisper.ListModels(); whisperModels != nil {
			models = append(models, whisperModels...)
		}
	}

	// Add OpenAI models
	if m.openai != nil {
		for _, modelID := range openai.Models {
			models = append(models, &schema.Model{
				Id:      modelID,
				Object:  "model",
				OwnedBy: "openai",
			})
		}
		for _, modelID := range openai.DiarizeModels {
			models = append(models, &schema.Model{
				Id:      modelID,
				Object:  "model",
				OwnedBy: "openai",
			})
		}
	}

	// Add ElevenLabs models
	if m.elevenlabs != nil {
		for _, modelID := range elevenlabs.Models {
			models = append(models, &schema.Model{
				Id:      modelID,
				Object:  "model",
				OwnedBy: "elevenlabs",
			})
		}
	}

	return models
}

// GetModel retrieves a model by ID from all available sources
func (m *Manager) GetModel(ctx context.Context, modelID string) (*schema.Model, error) {
	// OTEL request tracing
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "manager.GetModel",
		attribute.String("model.id", modelID),
	)
	var err error
	defer func() { endSpan(err) }()

	// List models and find by ID
	models := m.ListModels(ctx)
	for _, model := range models {
		if model.Id == modelID {
			return model, nil
		}
	}
	err = httpresponse.ErrNotFound.With(modelID)
	return nil, err
}

// DownloadModel downloads a whisper model by path
func (m *Manager) DownloadModel(ctx context.Context, path string, fn func(cur, total uint64)) (*schema.Model, error) {
	// OTEL request tracing
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "manager.DownloadModel",
		attribute.String("model.path", path),
	)
	var err error
	defer func() { endSpan(err) }()

	// Download the model
	result, err := m.whisper.DownloadModel(ctx, path, fn)
	return result, err
}

// DeleteModel deletes a whisper model by id
func (m *Manager) DeleteModel(ctx context.Context, modelID string) error {
	// OTEL request tracing
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "manager.DeleteModel",
		attribute.String("model.id", modelID),
	)
	var err error
	defer func() { endSpan(err) }()

	// Delete the model
	model, err := m.GetModel(ctx, modelID)
	if err != nil {
		return err
	} else if model.OwnedBy != "whisper" {
		err = errors.New("can only delete whisper models, " + modelID + " is owned by " + model.OwnedBy)
		return err
	}
	err = m.whisper.DeleteModelById(modelID)
	return err
}

// Transcribe performs a transcription request in the language of the speech
func (m *Manager) Transcribe(ctx context.Context, w schema.SegmentWriter, r io.Reader, req *schema.TranscribeRequest) (*schema.Transcription, error) {
	// OTEL request tracing
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "manager.Transcribe",
		attribute.String("request", req.String()),
	)
	var err error
	defer func() { endSpan(err) }()

	model, err := m.GetModel(ctx, req.Model)
	if err != nil {
		return nil, err
	}

	// Route based on model ownership
	switch model.OwnedBy {
	case "whisper":
		var result *schema.Transcription
		result, err = m.transcribeWhisper(ctx, w, r, req, model)
		return result, err
	case "openai":
		var result *schema.Transcription
		result, err = m.transcribeOpenAI(ctx, w, r, req, model)
		return result, err
	case "elevenlabs":
		var result *schema.Transcription
		result, err = m.transcribeElevenLabs(ctx, w, r, req, model)
		return result, err
	default:
		err = errors.New("unsupported model owner: " + model.OwnedBy)
		return nil, err
	}
}

// Translate performs a transcription request and returns the result in English
func (m *Manager) Translate(ctx context.Context, w schema.SegmentWriter, r io.Reader, req *schema.TranslateRequest) (*schema.Transcription, error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "manager.Translate",
		attribute.String("request", req.String()),
	)
	var err error
	defer func() { endSpan(err) }()

	// Find the model
	model, err := m.GetModel(ctx, req.Model)
	if err != nil {
		return nil, err
	}

	// Route based on model ownership
	switch model.OwnedBy {
	case "whisper":
		var result *schema.Transcription
		result, err = m.translateWhisper(ctx, w, r, req, model)
		return result, err
	case "openai":
		var result *schema.Transcription
		result, err = m.translateOpenAI(ctx, w, r, req, model)
		return result, err
	case "elevenlabs":
		var result *schema.Transcription
		result, err = m.translateElevenLabs(ctx, w, r, req, model)
		return result, err
	default:
		err = errors.New("unsupported model owner: " + model.OwnedBy)
		return nil, err
	}
}

// NewStreamSession creates a new streaming transcription session for the
// specified model. The caller MUST call the returned release function when
// the session is finished to return the task to the pool.
// Streaming is only supported for local whisper models.
func (m *Manager) NewStreamSession(ctx context.Context, cfg *schema.StreamConfig) (*whisper.StreamSession, func(), error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "manager.NewStreamSession",
		attribute.String("model", cfg.Model),
	)
	var err error
	defer func() { endSpan(err) }()

	// Resolve the model
	model, err := m.GetModel(ctx, cfg.Model)
	if err != nil {
		return nil, nil, err
	}

	// Streaming is only supported for local whisper models
	if model.OwnedBy != "whisper" {
		err = httpresponse.ErrBadRequest.With("streaming is only supported for local whisper models")
		return nil, nil, err
	}

	// Acquire a long-lived task from the pool
	task, release, err := m.whisper.AcquireTask(model)
	if err != nil {
		return nil, nil, err
	}

	// Configure task parameters from the stream config
	if cfg.Language != nil {
		task.SetLanguage(*cfg.Language)
	}
	if cfg.Temperature != nil {
		if err = task.SetTemperature(*cfg.Temperature); err != nil {
			release()
			return nil, nil, err
		}
	}
	if cfg.Prompt != nil {
		task.SetPrompt(*cfg.Prompt)
	}
	if cfg.Translate {
		task.SetTranslate(true)
	}

	session := whisper.NewStreamSession(task, *cfg)
	return session, release, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - WHISPER

// transcribeWhisper transcribes using the local whisper model
func (m *Manager) transcribeWhisper(ctx context.Context, w schema.SegmentWriter, r io.Reader, req *schema.TranscribeRequest, model *schema.Model) (*schema.Transcription, error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "whisper.Transcribe",
		attribute.String("model.id", model.Id),
	)
	var err error
	defer func() { endSpan(err) }()

	var result *schema.Transcription
	err = m.whisper.WithModel(model, func(task *whisper.Task) error {
		// Set transcription parameters from request
		if req.Language != nil {
			task.SetLanguage(types.PtrString(req.Language))
		}
		if req.Temperature != nil {
			if err := task.SetTemperature(types.PtrFloat64(req.Temperature)); err != nil {
				return err
			}
		}
		if req.Prompt != nil {
			task.SetPrompt(types.PtrString(req.Prompt))
		}
		if req.Diarize != nil && types.PtrBool(req.Diarize) {
			task.SetDiarize(true)
		}

		// Transcribe from reader, passing segment callback if writer provided
		var fn whisper.NewSegmentFunc
		if w != nil {
			fn = func(seg *schema.Segment) { w.Write(seg) }
		}
		if err := task.TranscribeReader(ctx, r, fn, m.segopts...); err != nil {
			return err
		}

		// Get the result
		result = task.Result()
		return nil
	})
	return result, err
}

// translateWhisper translates using the local whisper model
func (m *Manager) translateWhisper(ctx context.Context, w schema.SegmentWriter, r io.Reader, req *schema.TranslateRequest, model *schema.Model) (*schema.Transcription, error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "whisper.Translate",
		attribute.String("model.id", model.Id),
	)
	var err error
	defer func() { endSpan(err) }()

	// Execute translation with the model
	var result *schema.Transcription
	err = m.whisper.WithModel(model, func(task *whisper.Task) error {
		// Set translate flag
		task.SetTranslate(true)

		// Set translation parameters from request
		if req.Temperature != nil {
			if err := task.SetTemperature(types.PtrFloat64(req.Temperature)); err != nil {
				return err
			}
		}
		if req.Prompt != nil {
			task.SetPrompt(types.PtrString(req.Prompt))
		}

		// Translate from reader, passing segment callback if writer provided
		var fn whisper.NewSegmentFunc
		if w != nil {
			fn = func(seg *schema.Segment) { w.Write(seg) }
		}
		if err := task.TranscribeReader(ctx, r, fn, m.segopts...); err != nil {
			return err
		}

		// Get the result
		result = task.Result()
		return nil
	})

	return result, err
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - OPENAI TRANSCRIPTION

// transcribeOpenAI transcribes using the OpenAI API
func (m *Manager) transcribeOpenAI(ctx context.Context, w schema.SegmentWriter, r io.Reader, req *schema.TranscribeRequest, model *schema.Model) (*schema.Transcription, error) {
	// Determine filename with extension (default to .wav if not provided)
	filename := "audio.wav"
	if req.Filename != nil && *req.Filename != "" {
		filename = *req.Filename
	}

	// Create OpenAI transcription request
	openaiReq := openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			Model: model.Id,
			File: multipart.File{
				Body: r,
				Path: filename,
			},
			Prompt:      req.Prompt,
			Format:      types.StringPtr(openai.FormatVerboseJson), // Request verbose_json to get segments
			Temperature: req.Temperature,
		},
		Language: req.Language,
	}

	// Add diarization option
	if slices.Contains(openai.DiarizeModels, model.Id) {
		req.Diarize = types.BoolPtr(true)
	}
	if types.PtrBool(req.Diarize) {
		if !slices.Contains(openai.DiarizeModels, model.Id) {
			return nil, httpresponse.ErrBadRequest.Withf("diarization is only supported with %q", openai.DiarizeModels)
		} else {
			openaiReq.Format = types.StringPtr(openai.FormatDiarizedJson)
			openaiReq.ChunkingStrategy = &openai.ChunkingStrategy{Type: openai.ChunkingStrategyAuto}
		}
	}

	// Build stream callback if a segment writer is provided
	var segmentId int32
	var streamfn func(schema.Event)
	if w != nil {
		// Use json format for streaming (verbose_json not supported with streaming)
		if !types.PtrBool(req.Diarize) {
			openaiReq.Format = types.StringPtr(openai.FormatJson)
		}

		streamfn = func(evt schema.Event) {
			// Handle segment events for diarization or delta events for regular transcription
			if evt.Type == schema.TranscribeStreamSegmentType {
				// Diarized segment event - fields are at root level
				w.Write(&schema.Segment{
					Id:      segmentId,
					Start:   evt.Start,
					End:     evt.End,
					Text:    evt.Text,
					Speaker: evt.Speaker,
				})
				segmentId++
			} else if evt.Type == schema.TranscribeStreamDeltaType && evt.Delta != "" {
				// For non-diarized streaming, emit text deltas as partial segments
				w.Write(&schema.Segment{
					Id:   segmentId,
					Text: evt.Delta,
				})
				segmentId++
			}
		}
	}

	// Call OpenAI API
	resp, err := m.openai.Transcribe(ctx, openaiReq, streamfn)
	if err != nil {
		return nil, err
	}

	// Convert response to schema.Transcription
	return transcriptionOpenAIToSchema(resp), nil
}

// translateOpenAI translates using the OpenAI API
func (m *Manager) translateOpenAI(ctx context.Context, w schema.SegmentWriter, r io.Reader, req *schema.TranslateRequest, model *schema.Model) (*schema.Transcription, error) {
	// Determine filename with extension (default to .wav if not provided)
	filename := "audio.wav"
	if req.Filename != nil && *req.Filename != "" {
		filename = *req.Filename
	}

	// Create OpenAI translation request
	openaiReq := openai.TranslationRequest{
		Model: model.Id,
		File: multipart.File{
			Body: r,
			Path: filename,
		},
		Prompt:      req.Prompt,
		Format:      types.StringPtr(openai.FormatVerboseJson), // Request verbose_json to get segments
		Temperature: req.Temperature,
	}

	// Call OpenAI API
	resp, err := m.openai.Translate(ctx, openaiReq)
	if err != nil {
		return nil, err
	}

	// Convert response to schema.Transcription
	return transcriptionOpenAIToSchema(resp), nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - ELEVENLABS TRANSCRIPTION

// transcribeElevenLabs transcribes using the ElevenLabs API
func (m *Manager) transcribeElevenLabs(ctx context.Context, w schema.SegmentWriter, r io.Reader, req *schema.TranscribeRequest, model *schema.Model) (*schema.Transcription, error) {
	// Validate unsupported features
	if req.Prompt != nil && *req.Prompt != "" {
		return nil, httpresponse.ErrBadRequest.With("prompt is not supported by ElevenLabs")
	}

	// Create ElevenLabs transcription request
	elevenLabsReq := elevenlabs.TranscribeRequest{
		Model: model.Id,
		File: multipart.File{
			Body: r,
			Path: "audio",
		},
		Language:    req.Language,
		Temperature: req.Temperature,
		Diarize:     req.Diarize,
		Timestamps:  types.StringPtr("word"), // Request word-level timestamps for segmentation
	}

	// Call ElevenLabs API
	resp, err := m.elevenlabs.Transcribe(ctx, elevenLabsReq)
	if err != nil {
		return nil, err
	}

	// Convert response to schema.Transcription
	return transcriptionElevenLabsToSchema(resp), nil
}

// translateElevenLabs translates using the ElevenLabs API
func (m *Manager) translateElevenLabs(ctx context.Context, w schema.SegmentWriter, r io.Reader, req *schema.TranslateRequest, model *schema.Model) (*schema.Transcription, error) {
	// ElevenLabs only supports transcription, not translation
	return nil, httpresponse.ErrBadRequest.With("ElevenLabs does not support translation, only transcription")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE HELPERS

// transcriptionOpenAIToSchema converts an OpenAI transcription response to schema.Transcription
func transcriptionOpenAIToSchema(resp *openai.TranscriptionResponse) *schema.Transcription {
	if resp == nil {
		return nil
	}

	result := &schema.Transcription{
		TranscriptionSummary: schema.TranscriptionSummary{
			Language: resp.Language,
		},
		Text: resp.Text,
	}

	// Convert segments if present
	if len(resp.Segment) > 0 {
		for i, seg := range resp.Segment {
			if seg == nil {
				continue
			}
			// Use IdAsInt32() for verbose_json, fallback to index for diarized_json (string IDs)
			id := seg.IdAsInt32()
			if id == 0 && i > 0 {
				id = int32(i)
			}
			result.Segments = append(result.Segments, &schema.Segment{
				Id:      id,
				Start:   seg.Start,
				End:     seg.End,
				Text:    seg.Text,
				Speaker: seg.Speaker,
			})
		}
	}

	return result
}

// transcriptionElevenLabsToSchema converts an ElevenLabs transcription response to schema.Transcription
func transcriptionElevenLabsToSchema(resp *elevenlabs.TranscribeResponse) *schema.Transcription {
	if resp == nil {
		return nil
	}

	result := &schema.Transcription{
		TranscriptionSummary: schema.TranscriptionSummary{
			Language: resp.Language,
		},
		Text: resp.Text,
	}

	// Merge word-level timestamps into phrase-based segments
	if len(resp.Words) > 0 {
		result.Segments = mergeWordsIntoSegments(resp.Words)
	}

	return result
}

// mergeWordsIntoSegments combines word-level timestamps into larger phrase-based segments.
// Segments are split on sentence boundaries (. ! ?), speaker changes, or after 10 seconds of continuous speech.
func mergeWordsIntoSegments(words []elevenlabs.TranscribeWord) []*schema.Segment {
	if len(words) == 0 {
		return nil
	}

	const maxSegmentDuration = 10.0 // seconds

	var segments []*schema.Segment
	var currentText string
	var currentSpeaker *string
	var segmentStart float64
	var segmentId int32
	var prevWordEnd float64

	for i, word := range words {
		wordText := strings.TrimSpace(word.Text)
		if wordText == "" {
			continue // Skip empty words (e.g., spacing tokens)
		}

		// Check if we need to end the current segment before adding this word
		shouldEndBeforeWord := false

		if i > 0 {
			// End on speaker change (diarization)
			if word.Speaker != nil && currentSpeaker != nil && *word.Speaker != *currentSpeaker {
				shouldEndBeforeWord = true
			}

			// End if segment duration would exceed max
			if word.End-segmentStart > maxSegmentDuration {
				shouldEndBeforeWord = true
			}
		}

		// Finalize previous segment if needed
		if shouldEndBeforeWord && currentText != "" {
			seg := &schema.Segment{
				Id:    segmentId,
				Start: schema.SecToTimestamp(segmentStart),
				End:   schema.SecToTimestamp(prevWordEnd),
				Text:  currentText,
			}
			// Add speaker information if available
			if currentSpeaker != nil {
				seg.Speaker = *currentSpeaker
			}
			segments = append(segments, seg)
			segmentId++
			currentText = ""
			segmentStart = word.Start
			currentSpeaker = word.Speaker
		}

		// Initialize first segment
		if currentText == "" {
			segmentStart = word.Start
			currentSpeaker = word.Speaker
		}

		// Append word text
		if currentText != "" && !strings.HasSuffix(currentText, "-") {
			currentText += " "
		}
		currentText += wordText
		prevWordEnd = word.End

		// Check if we should end after this word
		isLastWord := i == len(words)-1
		shouldEndAfterWord := false

		// End on sentence boundaries
		if strings.HasSuffix(wordText, ".") || strings.HasSuffix(wordText, "!") || strings.HasSuffix(wordText, "?") {
			shouldEndAfterWord = true
		}

		// Always end on last word
		if isLastWord {
			shouldEndAfterWord = true
		}

		// Create segment if we should end
		if shouldEndAfterWord && currentText != "" {
			seg := &schema.Segment{
				Id:    segmentId,
				Start: schema.SecToTimestamp(segmentStart),
				End:   schema.SecToTimestamp(word.End),
				Text:  currentText,
			}
			// Add speaker information if available
			if currentSpeaker != nil {
				seg.Speaker = *currentSpeaker
			}
			segments = append(segments, seg)
			segmentId++
			currentText = ""
		}
	}

	return segments
}
