package whisper

import (
	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	MaxConcurrent int
}

type Opt func(*opts) error

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func OptMaxConcurrent(v int) Opt {
	return func(o *opts) error {
		if v < 1 {
			return ErrBadParameter.With("max concurrent must be greater than zero")
		}
		o.MaxConcurrent = v
		return nil
	}
}
