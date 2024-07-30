package whisper

import (
	"encoding/json"
	"strings"
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libwhisper
#include <whisper.h>
#include <stdio.h>
#include <stdbool.h>

extern void whisper_progress_cb_ex(struct whisper_context * ctx, struct whisper_state * state, int progress, void * user_data);
extern void whisper_segment_cb_ex(struct whisper_context * ctx, struct whisper_state * state, int n, void * user_data);
extern bool whisper_abort_cb_ex(void * user_data);

// Set callbacks
static void set_callbacks(struct whisper_full_params* params,  bool enabled) {
	if (enabled) {
		params->progress_callback = whisper_progress_cb_ex;
		params->abort_callback = whisper_abort_cb_ex;
		params->new_segment_callback = whisper_segment_cb_ex;
	} else {
		params->progress_callback = NULL;
		params->abort_callback = NULL;
		params->new_segment_callback = NULL;
	}
}
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	FullParams       C.struct_whisper_full_params
	SamplingStrategy C.enum_whisper_sampling_strategy
)

// Returns the new segment number
type ProgressCallback func(progress int)

// Returns the new segment number
type SegmentCallback func(segment int)

// If it returns true, the computation is aborted
type AbortCallback func() bool

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SAMPLING_GREEDY      SamplingStrategy = C.WHISPER_SAMPLING_GREEDY      // similar to OpenAI's GreedyDecoder
	SAMPLING_BEAM_SEARCH SamplingStrategy = C.WHISPER_SAMPLING_BEAM_SEARCH // similar to OpenAI's BeamSearchDecoder
)

var (
	// Map a uintptr context to a callback
	progressCb = map[uint]ProgressCallback{}
	segmentCb  = map[uint]SegmentCallback{}
	abortCb    = map[uint]AbortCallback{}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func DefaultFullParams(strategy SamplingStrategy) FullParams {
	params := (FullParams)(C.whisper_full_default_params((C.enum_whisper_sampling_strategy)(strategy)))
	C.set_callbacks((*C.struct_whisper_full_params)(&params), C.bool(true))
	return params
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx FullParams) MarshalJSON() ([]byte, error) {
	type j struct {
		Strategy                SamplingStrategy `json:"strategy"`
		NumThreads              int              `json:"n_threads,omitempty"`
		MaxTextCtx              int              `json:"n_max_text_ctx,omitempty"` // max tokens to use from past text as prompt for the decoder
		OffsetMS                int              `json:"offset_ms,omitempty"`      // start offset in ms
		DurationMS              int              `json:"duration_ms,omitempty"`    // audio duration to process in ms
		Translate               bool             `json:"translate,omitempty"`
		NoContext               bool             `json:"no_context,omitempty"`       // do not use past transcription (if any) as initial prompt for the decoder
		NoTimestamps            bool             `json:"no_timestamps,omitempty"`    // do not generate timestamps
		SingleSegment           bool             `json:"single_segment,omitempty"`   // force single segment output (useful for streaming)
		PrintSpecial            bool             `json:"print_special,omitempty"`    // print special tokens (e.g. <SOT>, <EOT>, <BEG>, etc.)
		PrintProgress           bool             `json:"print_progress,omitempty"`   // print progress information
		PrintRealtime           bool             `json:"print_realtime,omitempty"`   // print results from within whisper.cpp (avoid it, use callback instead)
		PrintTimestamps         bool             `json:"print_timestamps,omitempty"` // print timestamps for each text segment when printing realtime
		TokenTimestamps         bool             `json:"token_timestamps,omitempty"` // enable token-level timestamps
		TholdPt                 int              `json:"thold_pt,omitempty"`         // timestamp token probability threshold (~0.01)
		TholdPtsum              int              `json:"thold_ptsum,omitempty"`      // timestamp token sum probability threshold (~0.01)
		MaxLen                  int              `json:"max_len,omitempty"`          // max segment length in characters
		SplitOnWord             bool             `json:"split_on_word,omitempty"`    // split on word rather than on token (when used with max_len)
		MaxTokens               int              `json:"max_tokens,omitempty"`       // max tokens per segment (0 = no limit)
		DebugMode               bool             `json:"debug_mode,omitempty"`       // enable debug_mode provides extra info (eg. Dump log_mel)
		AudioCtx                int              `json:"audio_ctx,omitempty"`        // overwrite the audio context size (0 = use default)
		DiarizeEnable           bool             `json:"tdrz_enable,omitempty"`      // enable tinydiarize speaker turn detection
		SuppressRegex           string           `json:"suppress_regex,omitempty"`   // A regular expression that matches tokens to suppress
		InitialPrompt           string           `json:"initial_prompt,omitempty"`   // tokens to provide to the whisper decoder as initial prompt
		PromptTokens            []string         `json:"prompt_tokens,omitempty"`    // use whisper_tokenize() to convert text to tokens
		Language                string           `json:"language,omitempty"`         // for auto-detection, set to "" or "auto"
		DetectLanguage          bool             `json:"detect_language,omitempty"`
		SuppressBlank           bool             `json:"suppress_blank,omitempty"`             // ref: https://github.com/openai/whisper/blob/f82bc59f5ea234d4b97fb2860842ed38519f7e65/whisper/decoding.py#L89
		SuppressNonSpeechTokens bool             `json:"suppress_non_speech_tokens,omitempty"` // ref: https://github.com/openai/whisper/blob/7858aa9c08d98f75575035ecd6481f462d66ca27/whisper/tokenizer.py#L224-L253
		Temperature             float32          `json:"temperature,omitempty"`                // initial decoding temperature, ref: https://ai.stackexchange.com/a/32478
		MaxInitialTs            float32          `json:"max_initial_ts,omitempty"`             // ref: https://github.com/openai/whisper/blob/f82bc59f5ea234d4b97fb2860842ed38519f7e65/whisper/decoding.py#L97
		LengthPenalty           float32          `json:"length_penalty,omitempty"`             // ref: https://github.com/openai/whisper/blob/f82bc59f5ea234d4b97fb2860842ed38519f7e65/whisper/transcribe.py#L267
		TemperatureInc          float32          `json:"temperature_inc,omitempty"`            // ref: https://github.com/openai/whisper/blob/f82bc59f5ea234d4b97fb2860842ed38519f7e65/whisper/transcribe.py#L274-L278
		EntropyThreshold        float32          `json:"entropy_thold,omitempty"`              // similar to OpenAI's "compression_ratio_threshold"
		LogProbThreshold        float32          `json:"logprob_thold,omitempty"`
		ProgressCallback        uintptr          `json:"progress_callback,omitempty"`
		AbortCallback           uintptr          `json:"abort_callback,omitempty"`
		SegmentCallback         uintptr          `json:"segment_callback,omitempty"`
	}
	return json.Marshal(j{
		Strategy:                SamplingStrategy(ctx.strategy),
		NumThreads:              int(ctx.n_threads),
		MaxTextCtx:              int(ctx.n_max_text_ctx),
		OffsetMS:                int(ctx.offset_ms),
		DurationMS:              int(ctx.duration_ms),
		Translate:               bool(ctx.translate),
		NoContext:               bool(ctx.no_context),
		NoTimestamps:            bool(ctx.no_timestamps),
		SingleSegment:           bool(ctx.single_segment),
		PrintSpecial:            bool(ctx.print_special),
		PrintProgress:           bool(ctx.print_progress),
		PrintRealtime:           bool(ctx.print_realtime),
		PrintTimestamps:         bool(ctx.print_timestamps),
		TholdPt:                 int(ctx.thold_pt),
		TholdPtsum:              int(ctx.thold_ptsum),
		MaxLen:                  int(ctx.max_len),
		SplitOnWord:             bool(ctx.split_on_word),
		MaxTokens:               int(ctx.max_tokens),
		DebugMode:               bool(ctx.debug_mode),
		AudioCtx:                int(ctx.audio_ctx),
		DiarizeEnable:           bool(ctx.tdrz_enable),
		SuppressRegex:           C.GoString(ctx.suppress_regex),
		InitialPrompt:           C.GoString(ctx.initial_prompt),
		PromptTokens:            nil, // TODO
		Language:                C.GoString(ctx.language),
		DetectLanguage:          bool(ctx.detect_language),
		SuppressBlank:           bool(ctx.suppress_blank),
		SuppressNonSpeechTokens: bool(ctx.suppress_non_speech_tokens),
		Temperature:             float32(ctx.temperature),
		MaxInitialTs:            float32(ctx.max_initial_ts),
		LengthPenalty:           float32(ctx.length_penalty),
		TemperatureInc:          float32(ctx.temperature_inc),
		EntropyThreshold:        float32(ctx.entropy_thold),
		LogProbThreshold:        float32(ctx.logprob_thold),
		ProgressCallback:        uintptr(ctx.progress_callback_user_data),
		AbortCallback:           uintptr(ctx.abort_callback_user_data),
		SegmentCallback:         uintptr(ctx.new_segment_callback_user_data),
	})
}

func (ctx FullParams) String() string {
	str, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(str)
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
// PUBLIC METHODS

func (c *FullParams) SetNumThreads(v int) {
	c.n_threads = (C.int)(v)
}

func (c *FullParams) SetMaxTextCtx(v int) {
	c.n_max_text_ctx = (C.int)(v)
}

func (c *FullParams) SetOffsetMS(v int) {
	c.offset_ms = (C.int)(v)
}

func (c *FullParams) SetDurationMS(v int) {
	c.duration_ms = (C.int)(v)
}

func (c *FullParams) SetTranslate(v bool) {
	c.translate = (C.bool)(v)
}

func (c *FullParams) SetNoContext(v bool) {
	c.no_context = (C.bool)(v)
}

func (c *FullParams) SetNoTimestamps(v bool) {
	c.no_timestamps = (C.bool)(v)
}

func (c *FullParams) SetSingleSegment(v bool) {
	c.single_segment = (C.bool)(v)
}

func (c *FullParams) SetPrintSpecial(v bool) {
	c.print_special = (C.bool)(v)
}

func (c *FullParams) SetPrintProgress(v bool) {
	c.print_progress = (C.bool)(v)
}

func (c *FullParams) SetPrintRealtime(v bool) {
	c.print_realtime = (C.bool)(v)
}

func (c *FullParams) SetPrintTimestamps(v bool) {
	c.print_timestamps = (C.bool)(v)
}

func (c *FullParams) SetTokenTimestamps(v bool) {
	c.token_timestamps = (C.bool)(v)
}

func (c *FullParams) SetDiarizeEnable(v bool) {
	c.tdrz_enable = (C.bool)(v)
}

func (c *FullParams) SetLanguage(v string) {
	v = strings.ToLower(v)
	if v == "" || v == "auto" {
		c.language = (*C.char)(unsafe.Pointer(uintptr(0)))
	} else if id := Whisper_lang_id(v); id == -1 {
		c.language = (*C.char)(unsafe.Pointer(uintptr(0)))
	} else {
		c.language = C.whisper_lang_str(C.int(id))
	}
}

func (c *FullParams) SetProgressCallback(ctx *Context, cb ProgressCallback) {
	key := cbkey(unsafe.Pointer(ctx))
	if cb == nil {
		c.progress_callback_user_data = nil
		delete(progressCb, key)
	} else {
		c.progress_callback_user_data = unsafe.Pointer(uintptr(key))
		progressCb[key] = cb
	}
}

func (c *FullParams) SetSegmentCallback(ctx *Context, cb SegmentCallback) {
	key := cbkey(unsafe.Pointer(ctx))
	if cb == nil {
		c.new_segment_callback_user_data = nil
		delete(segmentCb, key)
	} else {
		c.new_segment_callback_user_data = unsafe.Pointer(uintptr(key))
		segmentCb[key] = cb
	}
}

func (c *FullParams) SetAbortCallback(ctx *Context, cb AbortCallback) {
	key := cbkey(unsafe.Pointer(ctx))
	if cb == nil {
		c.abort_callback_user_data = nil
		delete(abortCb, key)
	} else {
		c.abort_callback_user_data = unsafe.Pointer(uintptr(key))
		abortCb[key] = cb
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func cbkey(ptr unsafe.Pointer) uint {
	return uint(uintptr(ptr))
}

//export whisper_progress_cb_ex
func whisper_progress_cb_ex(ctx *C.struct_whisper_context, state *C.struct_whisper_state, progress C.int, user_data unsafe.Pointer) {
	key := cbkey(user_data)
	if cb, ok := progressCb[key]; ok {
		cb(int(progress))
	}
}

//export whisper_segment_cb_ex
func whisper_segment_cb_ex(ctx *C.struct_whisper_context, state *C.struct_whisper_state, n C.int, user_data unsafe.Pointer) {
	key := cbkey(user_data)
	if cb, ok := segmentCb[key]; ok {
		cb(int(n))
	}
}

//export whisper_abort_cb_ex
func whisper_abort_cb_ex(user_data unsafe.Pointer) C.bool {
	key := cbkey(user_data)
	if cb, ok := abortCb[key]; ok {
		return C.bool(cb())
	}
	return C.bool(false)
}
