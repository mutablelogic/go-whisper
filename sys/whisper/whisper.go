package whisper

import (
	"encoding/json"
	"errors"
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo LDFLAGS: -lwhisper -lggml -lm -lstdc++
#cgo darwin LDFLAGS: -framework Accelerate -framework Metal -framework Foundation -framework CoreGraphics
#include <whisper.h>
#include <ggml.h>
#include <stdlib.h>

extern void callNewSegment(void* user_data, int new);
extern void callProgress(void* user_data, int progress);
extern bool callEncoderBegin(void* user_data);
extern void callLog(enum ggml_log_level level,char* text, void* user_data);

// Text segment callback
// Called on every newly generated text segment
// Use the whisper_full_...() functions to obtain the text segments
static void whisper_new_segment_cb(struct whisper_context* ctx, struct whisper_state* state, int n_new, void* user_data) {
    if(user_data != NULL && ctx != NULL) {
        callNewSegment(user_data, n_new);
    }
}

// Progress callback
// Called on every newly generated text segment
// Use the whisper_full_...() functions to obtain the text segments
static void whisper_progress_cb(struct whisper_context* ctx, struct whisper_state* state, int progress, void* user_data) {
    if(user_data != NULL && ctx != NULL) {
        callProgress(user_data, progress);
    }
}

// Encoder begin callback
// If not NULL, called before the encoder starts
// If it returns false, the computation is aborted
static bool whisper_encoder_begin_cb(struct whisper_context* ctx, struct whisper_state* state, void* user_data) {
    if(user_data != NULL && ctx != NULL) {
        return callEncoderBegin(user_data);
    }
    return false;
}

// Get default parameters and set callbacks
static struct whisper_full_params whisper_full_default_params_cb(struct whisper_context* ctx, enum whisper_sampling_strategy strategy) {
	struct whisper_full_params params = whisper_full_default_params(strategy);
	params.new_segment_callback = whisper_new_segment_cb;
	params.new_segment_callback_user_data = (void*)(ctx);
	params.encoder_begin_callback = whisper_encoder_begin_cb;
	params.encoder_begin_callback_user_data = (void*)(ctx);
	params.progress_callback = whisper_progress_cb;
	params.progress_callback_user_data = (void*)(ctx);
	return params;
}

// Logging callback
#include <stdio.h>
static void whisper_log_cb(enum ggml_log_level level, const char* text, void* user_data) {
	callLog(level, (char*)text, user_data);
}

static void whisper_log_set_ex(void* user_data) {
	whisper_log_set(whisper_log_cb, user_data);
}

*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Context          C.struct_whisper_context
	Token            C.whisper_token
	TokenData        C.struct_whisper_token_data
	SamplingStrategy C.enum_whisper_sampling_strategy
	Params           C.struct_whisper_full_params
	LogLevel         C.enum_ggml_log_level
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SAMPLING_GREEDY      SamplingStrategy = C.WHISPER_SAMPLING_GREEDY      // similar to OpenAI's GreedyDecoder
	SAMPLING_BEAM_SEARCH SamplingStrategy = C.WHISPER_SAMPLING_BEAM_SEARCH // similar to OpenAI's BeamSearchDecoder
)

const (
	SampleRate = C.WHISPER_SAMPLE_RATE                 // Expected sample rate, samples per second
	SampleBits = uint16(unsafe.Sizeof(C.float(0))) * 8 // Sample size in bits
	NumFFT     = C.WHISPER_N_FFT
	HopLength  = C.WHISPER_HOP_LENGTH
	ChunkSize  = C.WHISPER_CHUNK_SIZE
)

var (
	ErrTokenizerFailed  = errors.New("whisper_tokenize failed")
	ErrAutoDetectFailed = errors.New("whisper_lang_auto_detect failed")
	ErrConversionFailed = errors.New("whisper_convert failed")
	ErrInvalidLanguage  = errors.New("invalid language")
)

var (
	cbLog func(level LogLevel, text string)
)

const (
	LogLevelDebug LogLevel = C.GGML_LOG_LEVEL_DEBUG
	LogLevelInfo  LogLevel = C.GGML_LOG_LEVEL_INFO
	LogLevelWarn  LogLevel = C.GGML_LOG_LEVEL_WARN
	LogLevelError LogLevel = C.GGML_LOG_LEVEL_ERROR
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocates all memory needed for the model and loads the model from the given file.
// Returns NULL on failure.
func Whisper_init(path string) *Context {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	if ctx := C.whisper_init_from_file_with_params(cPath, C.whisper_context_default_params()); ctx != nil {
		return (*Context)(ctx)
	} else {
		return nil
	}
}

// Frees all memory allocated by the model.
func (ctx *Context) Whisper_free() {
	C.whisper_free((*C.struct_whisper_context)(ctx))
}

// Run the entire model: PCM -> log mel spectrogram -> encoder -> decoder -> text
// Uses the specified decoding strategy to obtain the text.
func (ctx *Context) Whisper_full(
	params *Params,
	samples []float32,
	encoderBeginCallback func() bool,
	newSegmentCallback func(int),
	progressCallback func(int),
) error {
	registerEncoderBeginCallback(ctx, encoderBeginCallback)
	registerNewSegmentCallback(ctx, newSegmentCallback)
	registerProgressCallback(ctx, progressCallback)
	defer registerEncoderBeginCallback(ctx, nil)
	defer registerNewSegmentCallback(ctx, nil)
	defer registerProgressCallback(ctx, nil)
	if C.whisper_full((*C.struct_whisper_context)(ctx), (C.struct_whisper_full_params)(*params), (*C.float)(&samples[0]), C.int(len(samples))) == 0 {
		return nil
	} else {
		return ErrConversionFailed
	}
}

// Set logging output
func Whisper_log_set(fn func(level LogLevel, text string)) {
	if fn == nil {
		cbLog = nil
		C.whisper_log_set_ex(nil)
	} else {
		cbLog = fn
		C.whisper_log_set_ex(nil)
	}
}

// Return the id of the specified language, returns -1 if not found
// Examples:
//
//	"de" -> 2
//	"german" -> 2
func (ctx *Context) Whisper_lang_id(lang string) int {
	return int(C.whisper_lang_id(C.CString(lang)))
}

// Largest language id (i.e. number of available languages - 1)
func Whisper_lang_max_id() int {
	return int(C.whisper_lang_max_id())
}

// Return the short string of the specified language id (e.g. 2 -> "de"),
// returns empty string if not found
func Whisper_lang_str(id int) string {
	return C.GoString(C.whisper_lang_str(C.int(id)))
}

// Return the id of the autodetected language, returns -1 if not found
// Added to whisper.cpp in
// https://github.com/ggerganov/whisper.cpp/commit/a1c1583cc7cd8b75222857afc936f0638c5683d6
//
// Examples:
//
//	"de" -> 2
//	"german" -> 2
func (ctx *Context) Whisper_full_lang_id() int {
	return int(C.whisper_full_lang_id((*C.struct_whisper_context)(ctx)))
}

// Number of generated text segments.
// A segment can be a few words, a sentence, or even a paragraph.
func (ctx *Context) Whisper_full_n_segments() int {
	return int(C.whisper_full_n_segments((*C.struct_whisper_context)(ctx)))
}

// Get the start and end time of the specified segment.
func (ctx *Context) Whisper_full_get_segment_t0(segment int) int64 {
	return int64(C.whisper_full_get_segment_t0((*C.struct_whisper_context)(ctx), C.int(segment)))
}

// Get the start and end time of the specified segment.
func (ctx *Context) Whisper_full_get_segment_t1(segment int) int64 {
	return int64(C.whisper_full_get_segment_t1((*C.struct_whisper_context)(ctx), C.int(segment)))
}

// Get the text of the specified segment.
func (ctx *Context) Whisper_full_get_segment_text(segment int) string {
	return C.GoString(C.whisper_full_get_segment_text((*C.struct_whisper_context)(ctx), C.int(segment)))
}

// Get number of tokens in the specified segment.
func (ctx *Context) Whisper_full_n_tokens(segment int) int {
	return int(C.whisper_full_n_tokens((*C.struct_whisper_context)(ctx), C.int(segment)))
}

// Get the token text of the specified token index in the specified segment.
func (ctx *Context) Whisper_full_get_token_text(segment int, token int) string {
	return C.GoString(C.whisper_full_get_token_text((*C.struct_whisper_context)(ctx), C.int(segment), C.int(token)))
}

// Get the token of the specified token index in the specified segment.
func (ctx *Context) Whisper_full_get_token_id(segment int, token int) Token {
	return Token(C.whisper_full_get_token_id((*C.struct_whisper_context)(ctx), C.int(segment), C.int(token)))
}

// Get token data for the specified token in the specified segment.
// This contains probabilities, timestamps, etc.
func (ctx *Context) Whisper_full_get_token_data(segment int, token int) TokenData {
	return TokenData(C.whisper_full_get_token_data((*C.struct_whisper_context)(ctx), C.int(segment), C.int(token)))
}

// Get the probability of the specified token in the specified segment.
func (ctx *Context) Whisper_full_get_token_p(segment int, token int) float32 {
	return float32(C.whisper_full_get_token_p((*C.struct_whisper_context)(ctx), C.int(segment), C.int(token)))
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v LogLevel) String() string {
	switch v {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "LOG"
	}
}

func (v SamplingStrategy) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v SamplingStrategy) String() string {
	switch v {
	case SAMPLING_GREEDY:
		return "SAMPLING_GREEDY"
	case SAMPLING_BEAM_SEARCH:
		return "SAMPLING_BEAM_SEARCH"
	default:
		return "???"
	}
}

///////////////////////////////////////////////////////////////////////////////
// CALLBACKS

var (
	cbNewSegment   = make(map[unsafe.Pointer]func(int))
	cbProgress     = make(map[unsafe.Pointer]func(int))
	cbEncoderBegin = make(map[unsafe.Pointer]func() bool)
)

func registerNewSegmentCallback(ctx *Context, fn func(int)) {
	if fn == nil {
		delete(cbNewSegment, unsafe.Pointer(ctx))
	} else {
		cbNewSegment[unsafe.Pointer(ctx)] = fn
	}
}

func registerProgressCallback(ctx *Context, fn func(int)) {
	if fn == nil {
		delete(cbProgress, unsafe.Pointer(ctx))
	} else {
		cbProgress[unsafe.Pointer(ctx)] = fn
	}
}

func registerEncoderBeginCallback(ctx *Context, fn func() bool) {
	if fn == nil {
		delete(cbEncoderBegin, unsafe.Pointer(ctx))
	} else {
		cbEncoderBegin[unsafe.Pointer(ctx)] = fn
	}
}

//export callNewSegment
func callNewSegment(user_data unsafe.Pointer, new C.int) {
	if fn, ok := cbNewSegment[user_data]; ok {
		fn(int(new))
	}
}

//export callProgress
func callProgress(user_data unsafe.Pointer, progress C.int) {
	if fn, ok := cbProgress[user_data]; ok {
		fn(int(progress))
	}
}

//export callEncoderBegin
func callEncoderBegin(user_data unsafe.Pointer) C.bool {
	if fn, ok := cbEncoderBegin[user_data]; ok {
		if fn() {
			return C.bool(true)
		} else {
			return C.bool(false)
		}
	}
	return true
}

//export callLog
func callLog(level C.enum_ggml_log_level, text *C.char, user_data unsafe.Pointer) {
	if cbLog != nil {
		cbLog(LogLevel(level), C.GoString(text))
	}
}
