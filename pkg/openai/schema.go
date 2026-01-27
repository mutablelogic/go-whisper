package openai

import (
	"encoding/json"
	"io"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-client/pkg/multipart"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/mutablelogic/go-whisper/pkg/schema"
)

/////////////////////////////////////////////////////////////////////////////////
// TYPES

type TranslationRequest struct {
	Model       string         `json:"model"` // whisper-1
	File        multipart.File `json:"file"`
	Prompt      *string        `json:"prompt,omitempty"`
	Format      *string        `json:"response_format,omitempty"` // json, text, srt, verbose_json, or vtt
	Temperature *float64       `json:"temperature,omitempty"`     // 0.0 -> 1.0
}

type TranscriptionRequest struct {
	TranslationRequest
	Include                []string          `json:"include,omitempty"`                  // logprobs
	Language               *string           `json:"language,omitempty"`                 // Transcription only en, es, fr, etc.
	Stream                 *bool             `json:"stream,omitempty"`                   // If true, returns a stream of events
	Timestamps             []string          `json:"timestamp_granularities,omitempty"`  // combination of word, segment
	ChunkingStrategy       *ChunkingStrategy `json:"chunking_strategy,omitempty"`        // "auto" or server_vad object
	KnownSpeakerNames      []string          `json:"known_speaker_names,omitempty"`      // Speaker names for diarization (up to 4)
	KnownSpeakerReferences []string          `json:"known_speaker_references,omitempty"` // Audio samples as data URLs (2-10 seconds each)
}

// ChunkingStrategy controls how the audio is cut into chunks for diarization
type ChunkingStrategy struct {
	Type              string   `json:"type"`                          // "auto" or "server_vad"
	VadThreshold      *float64 `json:"threshold,omitempty"`           // VAD threshold (0.0-1.0)
	PrefixPaddingMs   *int     `json:"prefix_padding_ms,omitempty"`   // Padding before speech (ms)
	SilenceDurationMs *int     `json:"silence_duration_ms,omitempty"` // Silence duration to end segment (ms)
}

type TranscriptionResponse struct {
	Task     string                  `json:"task,omitempty"`
	Language string                  `json:"language,omitempty"`
	Duration schema.Timestamp        `json:"duration,omitempty"`
	Text     string                  `json:"text,omitempty"`
	Segment  []*TranscriptionSegment `json:"segments,omitempty" writer:",width:40,wrap"`
	Usage    *TranscriptionUsage     `json:"usage,omitempty"`
}

type TranscriptionUsage struct {
	Type    string `json:"type"`    // "duration"
	Seconds int    `json:"seconds"` // Billed duration in seconds
}

type TranscriptionSegment struct {
	Type             string           `json:"type,omitempty"` // Segment type (e.g., "transcript.text.segment" for diarized)
	Id               any              `json:"id"`             // Segment ID (int32 for verbose_json, string for diarized_json)
	Seek             uint32           `json:"seek,omitempty"`
	Start            schema.Timestamp `json:"start"`
	End              schema.Timestamp `json:"end"`
	Text             string           `json:"text"`
	Speaker          string           `json:"speaker,omitempty"`           // Speaker label for diarized transcription
	Tokens           []uint32         `json:"tokens,omitempty"`            // Array of token IDs for the text content.
	Temperature      *float64         `json:"temperature,omitempty"`       // Temperature parameter used for generating the segment.
	AvgLogProb       *float64         `json:"avg_logprob,omitempty"`       // Average logprob of the segment. If the value is lower than -1, consider the logprobs failed.
	CompressionRatio *float64         `json:"compression_ratio,omitempty"` // Compression ratio of the segment. If the value is greater than 2.4, consider the compression failed.
	NoSpeechProb     *float64         `json:"no_speech_prob,omitempty"`    // Probability of no speech in the segment. If the value is higher than 1.0 and the avg_logprob is below -1, consider this segment silent.
}

/////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	Endpoint       = "https://api.openai.com/v1/"
	TranscribePath = "audio/transcriptions" // Endpoint for transcription
	TranslatePath  = "audio/translations"   // Endpoint for translation
)

const (
	FormatJson         = "json"
	FormatVerboseJson  = "verbose_json"
	FormatDiarizedJson = "diarized_json"
	FormatText         = "text"
	FormatSrt          = "srt"
	FormatVtt          = "vtt"
)

const (
	streamDoneText = "[DONE]" // Text indicating the end of a stream
)

const (
	ChunkingStrategyAuto      = "auto"       // Auto chunking with VAD
	ChunkingStrategyServerVAD = "server_vad" // Server-side VAD chunking (required for diarization)
)

var (
	// Supported models for transcription and translation
	Models = []string{
		"whisper-1",
		"gpt-4o-mini-transcribe",
		"gpt-4o-mini-transcribe-2025-12-15",
		"gpt-4o-transcribe",
	}
	// Supported models for diarization
	DiarizeModels = []string{
		"gpt-4o-transcribe-diarize",
	}

	// Supported response formats
	Formats = []string{
		FormatText, FormatJson, FormatVerboseJson, FormatDiarizedJson, FormatSrt, FormatVtt,
	}
)

/////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s TranscriptionRequest) String() string {
	return stringify(s)
}

func (s TranslationRequest) String() string {
	return stringify(s)
}

func (s TranscriptionResponse) String() string {
	return stringify(s)
}

func (c ChunkingStrategy) String() string {
	if c.Type == "" || c.Type == ChunkingStrategyAuto {
		return ChunkingStrategyAuto
	}
	return stringify(c)
}

func stringify(s any) string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

/////////////////////////////////////////////////////////////////////////////////
// UNMARSHALL

func (s *TranscriptionResponse) Unmarshal(header http.Header, r io.Reader) error {
	mimetype, err := types.ParseContentType(header.Get(types.ContentTypeHeader))
	if err != nil {
		return err
	}
	switch mimetype {
	case types.ContentTypeJSON:
		// If the content type is JSON, we unmarshal directly
		return json.NewDecoder(r).Decode(&s)
	case types.ContentTypeTextPlain:
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		} else {
			s.Text = string(data)
		}
		return nil
	case types.ContentTypeTextStream:
		// Streaming responses are handled by the stream callback, not here
		return httpresponse.ErrNotImplemented
	}

	// Decode error
	return httpresponse.ErrBadRequest.Withf("Unsupported content type %q", mimetype)
}

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (s *TranscriptionResponse) Segments() *schema.Transcription {
	resp := &schema.Transcription{
		TranscriptionSummary: schema.TranscriptionSummary{
			Task:     s.Task,
			Language: s.Language,
			Duration: s.Duration,
		},
		Text:     s.Text,
		Segments: make([]*schema.Segment, 0, len(s.Segment)),
	}
	for _, seg := range s.Segment {
		resp.Segments = append(resp.Segments, &schema.Segment{
			Id:      seg.IdAsInt32(),
			Start:   seg.Start,
			End:     seg.End,
			Text:    seg.Text,
			Speaker: seg.Speaker,
		})
	}
	return resp
}

// IdAsInt32 returns the segment ID as int32, handling both numeric and string IDs
func (seg *TranscriptionSegment) IdAsInt32() int32 {
	switch v := seg.Id.(type) {
	case float64:
		return int32(v)
	case int:
		return int32(v)
	case int32:
		return v
	case int64:
		return int32(v)
	default:
		return 0
	}
}
