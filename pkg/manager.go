package pkg

import (
	"context"
	"errors"
	"io"

	// Packages
	goclient "github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/multipart"
	"github.com/mutablelogic/go-client/pkg/otel"
	types "github.com/mutablelogic/go-server/pkg/types"
	elevenlabs "github.com/mutablelogic/go-whisper/pkg/elevenlabs"
	openai "github.com/mutablelogic/go-whisper/pkg/openai"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
	attribute "go.opentelemetry.io/otel/attribute"
	trace "go.opentelemetry.io/otel/trace"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	whisper    *whisper.Manager
	elevenlabs *elevenlabs.Client
	openai     *openai.Client
	tracer     trace.Tracer
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new manager with a whisper client and optionally
// elevenlabs and openai clients if API keys are provided via Opts
func New(modelsPath string, whisperOpts []whisper.Opt, opt ...Opt) (*Manager, error) {
	var o opts
	self := new(Manager)

	// Apply options
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}

	// Create whisper client (required - never nil)
	whisperManager, err := whisper.New(modelsPath, whisperOpts...)
	if err != nil {
		return nil, err
	}
	self.whisper = whisperManager

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
	err = ErrNotFound.With(modelID)
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
func (m *Manager) Transcribe(ctx context.Context, r io.Reader, req *schema.TranscribeRequest) (*schema.Transcription, error) {
	// OTEL request tracing
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "manager.Transcribe",
		attribute.String("model.id", req.Model),
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
		result, err = m.transcribeWhisper(ctx, r, req, model)
		return result, err
	case "openai":
		var result *schema.Transcription
		result, err = m.transcribeOpenAI(ctx, r, req, model)
		return result, err
	case "elevenlabs":
		var result *schema.Transcription
		result, err = m.transcribeElevenLabs(ctx, r, req, model)
		return result, err
	default:
		err = errors.New("unsupported model owner: " + model.OwnedBy)
		return nil, err
	}
}

// Translate performs a transcription request and returns the result in English
func (m *Manager) Translate(ctx context.Context, r io.Reader, req *schema.TranslateRequest) (*schema.Transcription, error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "manager.Translate",
		attribute.String("model.id", req.Model),
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
		result, err = m.translateWhisper(ctx, r, req, model)
		return result, err
	case "openai":
		var result *schema.Transcription
		result, err = m.translateOpenAI(ctx, r, req, model)
		return result, err
	case "elevenlabs":
		var result *schema.Transcription
		result, err = m.translateElevenLabs(ctx, r, req, model)
		return result, err
	default:
		err = errors.New("unsupported model owner: " + model.OwnedBy)
		return nil, err
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - WHISPER

// transcribeWhisper transcribes using the local whisper model
func (m *Manager) transcribeWhisper(ctx context.Context, r io.Reader, req *schema.TranscribeRequest, model *schema.Model) (*schema.Transcription, error) {
	var result *schema.Transcription
	err := m.whisper.WithModel(model, func(task *whisper.Task) error {
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

		// Transcribe from reader
		if err := task.TranscribeReader(ctx, r, nil); err != nil {
			return err
		}

		// Get the result
		result = task.Result()
		return nil
	})
	return result, err
}

// translateWhisper translates using the local whisper model
func (m *Manager) translateWhisper(ctx context.Context, r io.Reader, req *schema.TranslateRequest, model *schema.Model) (*schema.Transcription, error) {
	// Execute translation with the model
	var result *schema.Transcription
	err := m.whisper.WithModel(model, func(task *whisper.Task) error {
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
		if req.Diarize != nil && types.PtrBool(req.Diarize) {
			task.SetDiarize(true)
		}

		// Transcribe from reader
		if err := task.TranscribeReader(ctx, r, nil); err != nil {
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
func (m *Manager) transcribeOpenAI(ctx context.Context, r io.Reader, req *schema.TranscribeRequest, model *schema.Model) (*schema.Transcription, error) {
	// Validate unsupported features
	if types.PtrBool(req.Stream) {
		return nil, ErrBadParameter.With("stream is not supported by OpenAI")
	}
	if types.PtrBool(req.Diarize) {
		return nil, ErrBadParameter.With("diarize is not supported by OpenAI")
	}

	// Create OpenAI transcription request
	openaiReq := openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			Model: model.Id,
			File: multipart.File{
				Body: r,
				Path: "audio",
			},
			Prompt:      req.Prompt,
			Format:      nil,
			Temperature: req.Temperature,
		},
		Language: req.Language,
		Stream:   nil,
	}

	// Call OpenAI API
	resp, err := m.openai.Transcribe(ctx, openaiReq)
	if err != nil {
		return nil, err
	}

	// Convert response to schema.Transcription
	return transcriptionOpenAIToSchema(resp), nil
}

// translateOpenAI translates using the OpenAI API
func (m *Manager) translateOpenAI(ctx context.Context, r io.Reader, req *schema.TranslateRequest, model *schema.Model) (*schema.Transcription, error) {
	// Validate unsupported features
	if types.PtrBool(req.Diarize) {
		return nil, ErrBadParameter.With("diarize is not supported by OpenAI")
	}

	// Create OpenAI translation request
	openaiReq := openai.TranslationRequest{
		Model: model.Id,
		File: multipart.File{
			Body: r,
			Path: "audio",
		},
		Prompt:      req.Prompt,
		Format:      nil,
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
func (m *Manager) transcribeElevenLabs(ctx context.Context, r io.Reader, req *schema.TranscribeRequest, model *schema.Model) (*schema.Transcription, error) {
	// Validate unsupported features
	if types.PtrBool(req.Stream) {
		return nil, ErrBadParameter.With("stream is not supported by ElevenLabs")
	}
	if req.Prompt != nil && *req.Prompt != "" {
		return nil, ErrBadParameter.With("prompt is not supported by ElevenLabs")
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
func (m *Manager) translateElevenLabs(ctx context.Context, r io.Reader, req *schema.TranslateRequest, model *schema.Model) (*schema.Transcription, error) {
	// ElevenLabs only supports transcription, not translation
	return nil, ErrBadParameter.With("ElevenLabs does not support translation, only transcription")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE HELPERS

// transcriptionOpenAIToSchema converts an OpenAI transcription response to schema.Transcription
func transcriptionOpenAIToSchema(resp *openai.TranscriptionResponse) *schema.Transcription {
	if resp == nil {
		return nil
	}

	result := &schema.Transcription{
		Text:     resp.Text,
		Language: resp.Language,
	}

	// Convert segments if present
	if len(resp.Segment) > 0 {
		for _, seg := range resp.Segment {
			if seg == nil {
				continue
			}
			result.Segments = append(result.Segments, &schema.Segment{
				Id:    seg.Id,
				Start: seg.Start,
				End:   seg.End,
				Text:  seg.Text,
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
		Text:     resp.Text,
		Language: resp.Language,
	}

	// Convert word-level timestamps to segments if present
	if len(resp.Words) > 0 {
		for _, word := range resp.Words {
			result.Segments = append(result.Segments, &schema.Segment{
				Start: schema.SecToTimestamp(word.Start),
				End:   schema.SecToTimestamp(word.End),
				Text:  word.Text,
			})
		}
	}

	return result
}
