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
#include <stdbool.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

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
		Strategy                SamplingStrategy `json:"strategy"`
		NThreads                int              `json:"n_threads,omitempty"`
		NMaxTextCtx             int              `json:"n_max_text_ctx,omitempty"`
		OffsetMs                int              `json:"offset_ms,omitempty"`
		DurationMs              int              `json:"duration_ms,omitempty"`
		Translate               bool             `json:"translate,omitempty"`
		NoContext               bool             `json:"no_context,omitempty"`
		NoTimestamps            bool             `json:"no_timestamps,omitempty"`
		SingleSegment           bool             `json:"single_segment,omitempty"`
		PrintSpecial            bool             `json:"print_special,omitempty"`
		PrintProgress           bool             `json:"print_progress,omitempty"`
		PrintRealtime           bool             `json:"print_realtime,omitempty"`
		PrintTimestamps         bool             `json:"print_timestamps,omitempty"`
		TokenTimestamps         bool             `json:"token_timestamps,omitempty"`
		TholdPt                 int              `json:"thold_pt,omitempty"`
		TholdPtsum              int              `json:"thold_ptsum,omitempty"`
		MaxLen                  int              `json:"max_len,omitempty"`
		SplitOnWord             bool             `json:"split_on_word,omitempty"`
		MaxTokens               int              `json:"max_tokens,omitempty"`
		DebugMode               bool             `json:"debug_mode,omitempty"`
		AudioCtx                int              `json:"audio_ctx,omitempty"`
		DiarizeEnable           bool             `json:"tdrz_enable,omitempty"`
		SuppressRegex           string           `json:"suppress_regex,omitempty"`
		InitialPrompt           string           `json:"initial_prompt,omitempty"`
		PromptTokens            []string         `json:"prompt_tokens,omitempty"`
		Language                string           `json:"language,omitempty"`
		DetectLanguage          bool             `json:"detect_language,omitempty"`
		SuppressBlank           bool             `json:"suppress_blank,omitempty"`
		SuppressNonSpeechTokens bool             `json:"suppress_non_speech_tokens,omitempty"`
		Temperature             float32          `json:"temperature,omitempty"`
		MaxInitialTs            float32          `json:"max_initial_ts,omitempty"`
		LengthPenalty           float32          `json:"length_penalty,omitempty"`
		TemperatureInc          float32          `json:"temperature_inc,omitempty"`
		EntropyThreshold        float32          `json:"entropy_thold,omitempty"`
		LogProbThreshold        float32          `json:"logprob_thold,omitempty"`
	}
	return json.Marshal(j{
		Strategy:                SamplingStrategy(p.strategy),
		NThreads:                int(p.n_threads),
		NMaxTextCtx:             int(p.n_max_text_ctx),
		OffsetMs:                int(p.offset_ms),
		DurationMs:              int(p.duration_ms),
		Translate:               bool(p.translate),
		NoContext:               bool(p.no_context),
		NoTimestamps:            bool(p.no_timestamps),
		SingleSegment:           bool(p.single_segment),
		PrintSpecial:            bool(p.print_special),
		PrintProgress:           bool(p.print_progress),
		PrintRealtime:           bool(p.print_realtime),
		PrintTimestamps:         bool(p.print_timestamps),
		TholdPt:                 int(p.thold_pt),
		TholdPtsum:              int(p.thold_ptsum),
		MaxLen:                  int(p.max_len),
		SplitOnWord:             bool(p.split_on_word),
		MaxTokens:               int(p.max_tokens),
		DebugMode:               bool(p.debug_mode),
		AudioCtx:                int(p.audio_ctx),
		DiarizeEnable:           bool(p.tdrz_enable),
		SuppressRegex:           C.GoString(p.suppress_regex),
		InitialPrompt:           C.GoString(p.initial_prompt),
		PromptTokens:            nil, // TODO
		Language:                C.GoString(p.language),
		DetectLanguage:          bool(p.detect_language),
		SuppressBlank:           bool(p.suppress_blank),
		SuppressNonSpeechTokens: bool(p.suppress_non_speech_tokens),
		Temperature:             float32(p.temperature),
		MaxInitialTs:            float32(p.max_initial_ts),
		LengthPenalty:           float32(p.length_penalty),
		TemperatureInc:          float32(p.temperature_inc),
		EntropyThreshold:        float32(p.entropy_thold),
		LogProbThreshold:        float32(p.logprob_thold),
	})
}

func (p *Params) String() string {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *Params) SetTranslate(v bool) {
	p.translate = toBool(v)
}

func (p *Params) SetSplitOnWord(v bool) {
	p.split_on_word = toBool(v)
}

func (p *Params) SetNoContext(v bool) {
	p.no_context = toBool(v)
}

func (p *Params) SetSingleSegment(v bool) {
	p.single_segment = toBool(v)
}

func (p *Params) SetPrintSpecial(v bool) {
	p.print_special = toBool(v)
}

func (p *Params) SetPrintProgress(v bool) {
	p.print_progress = toBool(v)
}

func (p *Params) SetPrintRealtime(v bool) {
	p.print_realtime = toBool(v)
}

func (p *Params) SetPrintTimestamps(v bool) {
	p.print_timestamps = toBool(v)
}

// Set language id
func (p *Params) SetLanguage(lang int) error {
	if lang == -1 {
		p.language = nil
		return nil
	}
	str := C.whisper_lang_str(C.int(lang))
	if str == nil {
		return ErrInvalidLanguage
	} else {
		p.language = str
	}
	return nil
}

// Get language id
func (p *Params) Language() int {
	if p.language == nil {
		return -1
	}
	return int(C.whisper_lang_id(p.language))
}

// Threads available
func (p *Params) Threads() int {
	return int(p.n_threads)
}

// Set number of threads to use
func (p *Params) SetThreads(threads int) {
	p.n_threads = C.int(threads)
}

// Set start offset in ms
func (p *Params) SetOffset(offset_ms int) {
	p.offset_ms = C.int(offset_ms)
}

// Set audio duration to process in ms
func (p *Params) SetDuration(duration_ms int) {
	p.duration_ms = C.int(duration_ms)
}

// Set timestamp token probability threshold (~0.01)
func (p *Params) SetTokenThreshold(t float32) {
	p.thold_pt = C.float(t)
}

// Set timestamp token sum probability threshold (~0.01)
func (p *Params) SetTokenSumThreshold(t float32) {
	p.thold_ptsum = C.float(t)
}

// Set max segment length in characters
func (p *Params) SetMaxSegmentLength(n int) {
	p.max_len = C.int(n)
}

func (p *Params) SetTokenTimestamps(b bool) {
	p.token_timestamps = toBool(b)
}

// Set max tokens per segment (0 = no limit)
func (p *Params) SetMaxTokensPerSegment(n int) {
	p.max_tokens = C.int(n)
}

// Set audio encoder context
func (p *Params) SetAudioCtx(n int) {
	p.audio_ctx = C.int(n)
}

// Set initial prompt
func (p *Params) SetInitialPrompt(prompt string) {
	// TODO: This results in a memory leak
	p.initial_prompt = C.CString(prompt)
}

// Set temperature
func (p *Params) SetTemperature(t float32) {
	p.temperature = C.float(t)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toBool(v bool) C.bool {
	if v {
		return C.bool(true)
	}
	return C.bool(false)
}
