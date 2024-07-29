package whisper

import "unsafe"

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libwhisper
#include <whisper.h>
#include <stdlib.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Context C.struct_whisper_context
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new context with path to model and context parameters. Returns nil on error.
func Whisper_init_from_file_with_params(path string, params ContextParams) *Context {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	return (*Context)(C.whisper_init_from_file_with_params(cPath, (C.struct_whisper_context_params)(params)))
}

// Create a new context with model data and context parameters. Returns nil on error.
func Whisper_init_from_buffer_with_params(data []byte, params ContextParams) *Context {
	return (*Context)(C.whisper_init_from_buffer_with_params(unsafe.Pointer(&data[0]), C.size_t(len(data)), (C.struct_whisper_context_params)(params)))
}

// Frees all memory allocated by the model.
func Whisper_free(ctx *Context) {
	C.whisper_free((*C.struct_whisper_context)(ctx))
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Return largest language id (i.e. number of available languages - 1)
func Whisper_lang_max_id() int {
	return int(C.whisper_lang_max_id())
}

// Return the language id of the specified short or full string (e.g. "de" -> 2),
// or -1 if not found
func Whisper_lang_id(str string) int {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return int(C.whisper_lang_id(cStr))
}

// Return the short string of the specified language id (e.g. 2 -> "de"),
// or empty string if not found
func Whisper_lang_str(id int) string {
	return C.GoString(C.whisper_lang_str(C.int(id)))
}

// Return the long string of the specified language name (e.g. 2 -> "german"),
// or empty string if not found
func Whisper_lang_str_full(id int) string {
	return C.GoString(C.whisper_lang_str_full(C.int(id)))
}

// Return model capabilities
func Whisper_is_multilingual(ctx *Context) bool {
	return C.whisper_is_multilingual((*C.struct_whisper_context)(ctx)) != 0
}

// Run the entire model: PCM -> log mel spectrogram -> encoder -> decoder -> text
// Not thread safe for same context
// Uses the specified decoding strategy to obtain the text.
func Whisper_full(ctx *Context, params FullParams, samples []float32) error {
	if C.whisper_full((*C.struct_whisper_context)(ctx), (C.struct_whisper_full_params)(params), (*C.float)(&samples[0]), C.int(len(samples))) != 0 {
		return ErrTranscriptionFailed
	}
	return nil
}

// Number of generated text segments
// A segment can be a few words, a sentence, or even a paragraph.
func (ctx *Context) NumSegments() int {
	return int(C.whisper_full_n_segments((*C.struct_whisper_context)(ctx)))
}

// Language id associated with the context's default state
func (ctx *Context) DefaultLangId() int {
	return int(C.whisper_full_lang_id((*C.struct_whisper_context)(ctx)))
}

// Get the start time of the specified segment
func (ctx *Context) SegmentT0(n int) int64 {
	return int64(C.whisper_full_get_segment_t0((*C.struct_whisper_context)(ctx), C.int(n)))
}

// Get the end time of the specified segment
func (ctx *Context) SegmentT1(n int) int64 {
	return int64(C.whisper_full_get_segment_t1((*C.struct_whisper_context)(ctx), C.int(n)))
}

// Get whether the next segment is predicted as a speaker turn
func (ctx *Context) SegmentSpeakerTurnNext(n int) bool {
	return (bool)(C.whisper_full_get_segment_speaker_turn_next((*C.struct_whisper_context)(ctx), C.int(n)))
}

// Get the text of the specified segment
func (ctx *Context) SegmentText(n int) string {
	return C.GoString(C.whisper_full_get_segment_text((*C.struct_whisper_context)(ctx), C.int(n)))
}

// Get number of tokens in the specified segment
func (ctx *Context) SegmentNumTokens(n int) int {
	return int(C.whisper_full_n_tokens((*C.struct_whisper_context)(ctx), C.int(n)))
}

// Get the token text in the specified segment
func (ctx *Context) SegmentTokenText(n, i int) string {
	return C.GoString(C.whisper_full_get_token_text((*C.struct_whisper_context)(ctx), C.int(n), C.int(i)))
}

// Get the token id in the specified segment
func (ctx *Context) SegmentTokenId(n, i int) int32 {
	return int32(C.whisper_full_get_token_id((*C.struct_whisper_context)(ctx), C.int(n), C.int(i)))
}

// Get the token probability of the specified token in the specified segment
func (ctx *Context) SegmentTokenProb(n, i int) float32 {
	return float32(C.whisper_full_get_token_p((*C.struct_whisper_context)(ctx), C.int(n), C.int(i)))
}

// Get token data for the specified token in the specified segment
// This contains probabilities, timestamps, etc.
func (ctx *Context) SegmentTokenData(n, i int) TokenData {
	return (TokenData)(C.whisper_full_get_token_data((*C.struct_whisper_context)(ctx), C.int(n), C.int(i)))
}
