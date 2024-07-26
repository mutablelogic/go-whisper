package whisper

import (
	"encoding/json"
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo LDFLAGS: -lwhisper -lggml -lm -lstdc++
#cgo darwin LDFLAGS: -framework Accelerate -framework Metal -framework Foundation -framework CoreGraphics
#include <whisper.h>
#include <stdlib.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Returns new default parameters. Call Close() to free memory associated with the parameters.
func NewParams(strategy SamplingStrategy) *Params {
	return (*Params)(C.whisper_full_default_params_by_ref(C.enum_whisper_sampling_strategy(strategy)))
}

func (p *Params) Close() {
	C.free(unsafe.Pointer(p))
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *Params) MarshalJSON() ([]byte, error) {
	type j struct {
		Strategy        SamplingStrategy `json:"strategy"`
		NThreads        int              `json:"n_threads,omitempty"`
		NMaxTextCtx     int              `json:"n_max_text_ctx,omitempty"`
		OffsetMs        int              `json:"offset_ms,omitempty"`
		DurationMs      int              `json:"duration_ms,omitempty"`
		Translate       bool             `json:"translate,omitempty"`
		NoContext       bool             `json:"no_context,omitempty"`
		NoTimestamps    bool             `json:"no_timestamps,omitempty"`
		SingleSegment   bool             `json:"single_segment,omitempty"`
		PrintSpecial    bool             `json:"print_special,omitempty"`
		PrintProgress   bool             `json:"print_progress,omitempty"`
		PrintRealtime   bool             `json:"print_realtime,omitempty"`
		PrintTimestamps bool             `json:"print_timestamps,omitempty"`
	}
	return json.Marshal(j{
		Strategy:        SamplingStrategy(p.strategy),
		NThreads:        int(p.n_threads),
		NMaxTextCtx:     int(p.n_max_text_ctx),
		OffsetMs:        int(p.offset_ms),
		DurationMs:      int(p.duration_ms),
		Translate:       bool(p.translate),
		NoContext:       bool(p.no_context),
		NoTimestamps:    bool(p.no_timestamps),
		SingleSegment:   bool(p.single_segment),
		PrintSpecial:    bool(p.print_special),
		PrintProgress:   bool(p.print_progress),
		PrintRealtime:   bool(p.print_realtime),
		PrintTimestamps: bool(p.print_timestamps),
	})
}

func (p *Params) String() string {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
