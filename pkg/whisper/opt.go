package whisper

import (
	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	MaxConcurrent int
	logfn         LogFn
	debug         bool
	gpu           int
}

type Opt func(*opts) error
type LogFn func(string)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set maximum number of concurrent tasks
func OptMaxConcurrent(v int) Opt {
	return func(o *opts) error {
		if v < 1 {
			return ErrBadParameter.With("max concurrent must be greater than zero")
		}
		o.MaxConcurrent = v
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
