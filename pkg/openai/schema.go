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
	Include    []string `json:"include,omitempty"`                 // logprobs
	Language   *string  `json:"language,omitempty"`                // Transcription only en, es, fr, etc.
	Stream     *bool    `json:"stream,omitempty"`                  // If true, returns a stream of events
	Timestamps []string `json:"timestamp_granularities,omitempty"` // combination of word, segment
}

type TranscriptionResponse struct {
	Task     string                  `json:"task,omitempty"`
	Language string                  `json:"language,omitempty"`
	Duration schema.Timestamp        `json:"duration,omitempty"`
	Text     string                  `json:"text,omitempty"`
	Segment  []*TranscriptionSegment `json:"segments,omitempty" writer:",width:40,wrap"`
}

type TranscriptionSegment struct {
	Id               int32            `json:"id"`
	Seek             uint32           `json:"seek"`
	Start            schema.Timestamp `json:"start"`
	End              schema.Timestamp `json:"end"`
	Text             string           `json:"text"`
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
	FormatJson        = "json"
	FormatVerboseJson = "verbose_json"
	FormatText        = "text"
	FormatSrt         = "srt"
	FormatVtt         = "vtt"
)

const (
	streamDoneText = "[DONE]" // Text indicating the end of a stream
)

var (
	Models  = []string{"whisper-1", "gpt-4o-mini-transcribe", "gpt-4o-transcribe"} // Supported models for transcription and translation
	Formats = []string{
		FormatText, FormatJson, FormatVerboseJson, FormatSrt, FormatVtt,
	}
)

/////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s TranscriptionRequest) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (s TranslationRequest) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (s TranscriptionResponse) String() string {
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
			Id:    seg.Id,
			Start: seg.Start,
			End:   seg.End,
			Text:  seg.Text,
		})
	}
	return resp
}
