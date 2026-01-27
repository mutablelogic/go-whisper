package whisper

import (
	"runtime"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	trace "go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	max    int // Max number of concurrent tasks
	gpu    int // GPU index, -1 for no GPU
	logfn  LogFn
	debug  bool
	tracer trace.Tracer
}

type Opt func(*opts) error
type LogFn func(string)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opt ...Opt) (opts, error) {
	var o opts

	// Set defaults
	o.max = runtime.NumCPU()
	o.gpu = 0

	// Apply options
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return o, err
		}
	}
	return o, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set maximum number of concurrent tasks
func OptMaxConcurrent(v int) Opt {
	return func(o *opts) error {
		if v < 1 {
			return httpresponse.ErrBadRequest.With("max concurrent must be greater than zero")
		}
		o.max = v
		return nil
	}
}

// Set logging function
func OptLog(fn LogFn) Opt {
	return func(o *opts) error {
		o.logfn = fn
		return nil
	}
}

// Set debugging
func OptDebug() Opt {
	return func(o *opts) error {
		o.debug = true
		return nil
	}
}

// Disable GPU acceleration
func OptNoGPU() Opt {
	return func(o *opts) error {
		o.gpu = -1
		return nil
	}
}

// Add OTEL Tracer
func OptTracer(tracer trace.Tracer) Opt {
	return func(o *opts) error {
		o.tracer = tracer
		return nil
	}
}
