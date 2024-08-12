package rnnoise

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: rnnoise
#include <rnnoise.h>
*/
import "C"
import "errors"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	DenoiseState C.DenoiseState
	RNNModel     C.RNNModel
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// SampleRate expected by rnnoise
	SampleRate = 48000
)

var (
	ErrInvalidModel = errors.New("invalid model")
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return the size of DenoiseState
func Rnnoise_get_size() int {
	return int(C.rnnoise_get_size())
}

// Return the number of samples processed by rnnoise_process_frame at a time
func Rnnoise_get_frame_size() int {
	return int(C.rnnoise_get_frame_size())
}

// Initializes a pre-allocated DenoiseState
// If model is NULL the default model is used.
func Rnnoise_init(state *DenoiseState, model *RNNModel) error {
	if C.rnnoise_init((*C.DenoiseState)(state), (*C.RNNModel)(model)) != 0 {
		return ErrInvalidModel
	} else {
		return nil
	}
}

// Allocate and initialize a DenoiseState
// If model is NULL the default model is used.
// The returned pointer MUST be freed with rnnoise_destroy().
func Rnnoise_create(model *RNNModel) (*DenoiseState, error) {
	if st := (*DenoiseState)(C.rnnoise_create((*C.RNNModel)(model))); st == nil {
		return nil, ErrInvalidModel
	} else {
		return st, nil
	}
}

// Free a DenoiseState produced by rnnoise_create.
// The optional custom model must be freed by rnnoise_model_free() after.
func Rnnoise_destroy(state *DenoiseState) {
	C.rnnoise_destroy((*C.DenoiseState)(state))
}

// Denoise a frame of samples
// in and out must be at least rnnoise_get_frame_size() large (in samples, not bytes)
// Returns the speech probability of the frame (VAD probability)
func Rnnoise_process_frame(state *DenoiseState, out, in []float32) float32 {
	return float32(C.rnnoise_process_frame((*C.DenoiseState)(state), (*C.float)(&out[0]), (*C.float)(&in[0])))
}
