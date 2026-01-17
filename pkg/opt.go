package pkg

import (
	"time"

	// Packages
	goclient "github.com/mutablelogic/go-client"
	"go.opentelemetry.io/otel/trace"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	elevenLabsKey string
	openAIKey     string
	clientOpts    []goclient.ClientOpt
	tracer        trace.Tracer
}

type Opt func(*opts) error

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// OptElevenLabsKey sets the ElevenLabs API key
func OptElevenLabsKey(key string) Opt {
	return func(o *opts) error {
		if key == "" {
			return ErrBadParameter.With("empty elevenlabs API key")
		}
		o.elevenLabsKey = key
		return nil
	}
}

// OptOpenAIKey sets the OpenAI API key
func OptOpenAIKey(key string) Opt {
	return func(o *opts) error {
		if key == "" {
			return ErrBadParameter.With("empty openai API key")
		}
		o.openAIKey = key
		return nil
	}
}

// OptClientOpts appends client options applied to HTTP-based providers
// (ElevenLabs and OpenAI). This is useful for setting timeouts, tracing,
// headers, and other transport-level behaviors shared across providers.
func OptClientOpts(clientOpts ...goclient.ClientOpt) Opt {
	return func(o *opts) error {
		if len(clientOpts) == 0 {
			return ErrBadParameter.With("client options are empty")
		}
		o.clientOpts = append(o.clientOpts, clientOpts...)
		return nil
	}
}

// OptClientTimeout is a convenience wrapper to set the HTTP client timeout for
// ElevenLabs and OpenAI requests.
func OptClientTimeout(d time.Duration) Opt {
	return OptClientOpts(goclient.OptTimeout(d))
}

// OptTracer sets the OpenTelemetry tracer for distributed tracing of
// transcription, translation, and model operations across all providers.
func OptTracer(tracer trace.Tracer) Opt {
	return func(o *opts) error {
		if tracer == nil {
			return ErrBadParameter.With("tracer is nil")
		}
		o.tracer = tracer
		return nil
	}
}
