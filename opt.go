package whisper

import (
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Re-export types from pkg/whisper
type Opt = whisper.Opt
type LogFn = whisper.LogFn

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Re-export option functions from pkg/whisper
var (
	OptMaxConcurrent = whisper.OptMaxConcurrent
	OptLog           = whisper.OptLog
	OptDebug         = whisper.OptDebug
	OptNoGPU         = whisper.OptNoGPU
)
