package httpclient

import (
	"fmt"
	"path/filepath"

	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	schema.TranscribeMultipartRequest
	format FormatType
}

// Opt is an option to set on the client request.
type Opt func(*opt) error

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opts ...Opt) (*opt, error) {
	o := new(opt)
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

///////////////////////////////////////////////////////////////////////////////
// OPTIONS - COMMON

// WithPrompt sets an optional text prompt to guide the model's style.
func WithPrompt(prompt string) Opt {
	return func(o *opt) error {
		if prompt != "" {
			o.Prompt = &prompt
		}
		return nil
	}
}

// WithTemperature sets the sampling temperature for transcription.
// Valid range is [0, 1] inclusive, where 0 is deterministic and 1 is most random.
func WithTemperature(temperature float64) Opt {
	return func(o *opt) error {
		if temperature < 0 || temperature > 1 {
			return fmt.Errorf("temperature must be between 0 and 1 (inclusive)")
		}
		o.Temperature = &temperature
		return nil
	}
}

// WithDiarize enables speaker diarization (identifying and separating speakers).
func WithDiarize(diarize bool) Opt {
	return func(o *opt) error {
		o.Diarize = &diarize
		return nil
	}
}

// WithStream enables streaming response.
func WithStream(stream bool) Opt {
	return func(o *opt) error {
		o.Stream = &stream
		return nil
	}
}

// WithFilename sets an optional filename with extension for the audio file.
// Only the base filename (not the full path) will be used.
func WithFilename(filename string) Opt {
	return func(o *opt) error {
		if filename != "" {
			o.Audio.Path = filepath.Base(filename)
		}
		return nil
	}
}

// WithFormat sets the Accept header to request a specific response format.
// Use the Format* constants (FormatJSON, FormatText, FormatVTT, FormatSRT).
func WithFormat(format FormatType) Opt {
	return func(o *opt) error {
		if format != "" {
			o.format = format
		}
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// OPTIONS - TRANSCRIBE ONLY

// WithLanguage sets the language of the audio (two-letter code) for transcription.
// Use "auto" or empty string for automatic language detection.
func WithLanguage(language string) Opt {
	return func(o *opt) error {
		if language != "" {
			o.Language = &language
		}
		return nil
	}
}
