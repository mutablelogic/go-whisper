package whisper

import (
	"encoding/json"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libwhisper
#include <whisper.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	TokenData C.struct_whisper_token_data
	Token     struct {
		Id   int32         `json:"id"`
		Text string        `json:"text,omitempty"`
		P    float32       `json:"p,omitempty"`
		T0   time.Duration `json:"t0,omitempty"`
		T1   time.Duration `json:"t1,omitempty"`
		Type TokenType     `json:"type,omitempty"`
	}
	Segment struct {
		Id          int32         `json:"id"`
		Text        string        `json:"text,omitempty"`
		T0          time.Duration `json:"t0,omitempty"`
		T1          time.Duration `json:"t1,omitempty"`
		SpeakerTurn bool          `json:"speaker_turn,omitempty"`
		Tokens      []Token       `json:"tokens,omitempty"`
	}
	TokenType int
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	_    TokenType = iota
	EOT            // end of text
	SOT            // start of transcript
	SOLM           // start of language model
	PREV           // start of previous segment
	NOSP           // no speaker
	NOT            // no timestamps
	BEG            // begin
	LANG           // language (TODO)
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Token) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (s Segment) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t TokenType) String() string {
	switch t {
	case EOT:
		return "[EOT]"
	case SOT:
		return "[SOT]"
	case SOLM:
		return "[SOLM]"
	case PREV:
		return "[PREV]"
	case NOSP:
		return "[NOSP]"
	case NOT:
		return "[NOT]"
	case BEG:
		return "[BEG]"
	case LANG:
		return "[LANG]"
	default:
		return "[TEXT]"
	}
}

func (t TokenType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a segment, or nil
func (ctx *Context) Segment(n int) *Segment {
	if n < 0 || n >= ctx.NumSegments() {
		return nil
	}
	return &Segment{
		Id:          int32(n),
		Text:        C.GoString(C.whisper_full_get_segment_text((*C.struct_whisper_context)(ctx), C.int(n))),
		SpeakerTurn: (bool)(C.whisper_full_get_segment_speaker_turn_next((*C.struct_whisper_context)(ctx), C.int(n))),
		Tokens:      ctx.Tokens(n),
		T0:          tsToDuration(C.whisper_full_get_segment_t0((*C.struct_whisper_context)(ctx), C.int(n))),
		T1:          tsToDuration(C.whisper_full_get_segment_t1((*C.struct_whisper_context)(ctx), C.int(n))),
	}
}

// Return a token from TokenData
func (ctx *Context) Token(data TokenData) Token {
	return Token{
		Id:   int32(data.id),
		Text: C.GoString(C.whisper_token_to_str((*C.struct_whisper_context)(ctx), C.whisper_token(data.id))),
		P:    float32(data.p),
		T0:   tsToDuration(data.t0),
		T1:   tsToDuration(data.t1),
		Type: tokenToType(ctx, data.id),
	}
}

// Return tokens for a segment
func (ctx *Context) Tokens(n int) []Token {
	if n >= ctx.NumSegments() {
		return nil
	}
	t := int(C.whisper_full_n_tokens((*C.struct_whisper_context)(ctx), C.int(n)))
	if t < 0 {
		return nil
	}
	result := make([]Token, t)
	for i := 0; i < t; i++ {
		data := (TokenData)(C.whisper_full_get_token_data((*C.struct_whisper_context)(ctx), C.int(n), C.int(i)))
		result[i] = ctx.Token(data)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE  METHODS

// Convert a int64 timestamp to a time.Duration
func tsToDuration(ts C.int64_t) time.Duration {
	if ts == -1 {
		return 0
	}
	return time.Duration(ts) * time.Millisecond
}

// return a token type from a token
// TODO: Return language tokens
func tokenToType(ctx *Context, token C.int32_t) TokenType {
	switch {
	case token == C.whisper_token_eot((*C.struct_whisper_context)(ctx)):
		return EOT
	case token == C.whisper_token_sot((*C.struct_whisper_context)(ctx)):
		return SOT
	case token == C.whisper_token_solm((*C.struct_whisper_context)(ctx)):
		return SOLM
	case token == C.whisper_token_prev((*C.struct_whisper_context)(ctx)):
		return PREV
	case token == C.whisper_token_nosp((*C.struct_whisper_context)(ctx)):
		return NOSP
	case token == C.whisper_token_not((*C.struct_whisper_context)(ctx)):
		return NOT
	case token == C.whisper_token_beg((*C.struct_whisper_context)(ctx)):
		return BEG
	default:
		return 0
	}
}
