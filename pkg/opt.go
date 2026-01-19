package pkg

import (
	"time"

	// Packages
	goclient "github.com/mutablelogic/go-client"
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
	trace "go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	elevenLabsKey string
	openAIKey     string
	clientOpts    []goclient.ClientOpt
	segOpts       []segmenter.Opt
	whisperopts   []whisper.Opt
	tracer        trace.Tracer
}

type Opt func(*opts) error

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// OptElevenLabsKey sets the ElevenLabs API key
func OptElevenLabsKey(key string) Opt {
	return func(o *opts) error {
		o.elevenLabsKey = key
		return nil
	}
}

// OptOpenAIKey sets the OpenAI API key
func OptOpenAIKey(key string) Opt {
	return func(o *opts) error {
		o.openAIKey = key
		return nil
	}
}

// WithClientOpts appends client options applied to HTTP-based providers
// (ElevenLabs and OpenAI). This is useful for setting timeouts, tracing,
// headers, and other transport-level behaviors shared across providers.
func WithClientOpts(clientOpts ...goclient.ClientOpt) Opt {
	return func(o *opts) error {
		o.clientOpts = append(o.clientOpts, clientOpts...)
		return nil
	}
}

// WithSegmenterOpt appends segmenter options applied to the audio segmenter
// used during transcription and translation.
func WithSegmenterOpt(segmenterOpts ...segmenter.Opt) Opt {
	return func(o *opts) error {
		o.segOpts = append(o.segOpts, segmenterOpts...)
		return nil
	}
}

// WithWhisperOpt appends whisper options applied to the whisper manager
// used during transcription and translation.
func WithWhisperOpt(whisperOpts ...whisper.Opt) Opt {
	return func(o *opts) error {
		o.whisperopts = append(o.whisperopts, whisperOpts...)
		return nil
	}
}

// WithClientTimeout is a convenience wrapper to set the HTTP client timeout for
// ElevenLabs and OpenAI requests.
func WithClientTimeout(d time.Duration) Opt {
	return WithClientOpts(goclient.OptTimeout(d))
}

// WithTracer sets the OpenTelemetry tracer for distributed tracing of
// transcription, translation, and model operations across all providers.
func WithTracer(tracer trace.Tracer) Opt {
	return func(o *opts) error {
		o.tracer = tracer
		o.whisperopts = append(o.whisperopts, whisper.OptTracer(tracer))
		return nil
	}
}
