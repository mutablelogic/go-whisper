package whisper

import (
	"fmt"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#include <stdbool.h>
*/
import "C"

/*
   typedef int whisper_token;

   typedef struct whisper_token_data {
       whisper_token id;  // token id
       whisper_token tid; // forced timestamp token id

       float p;           // probability of the token
       float pt;          // probability of the timestamp token
       float ptsum;       // sum of probabilities of all timestamp tokens

       // token-level timestamp data
       // do not use if you haven't computed token-level timestamps
       int64_t t0;        // start time of the token
       int64_t t1;        //   end time of the token

       float vlen;        // voice length of the token
   } whisper_token_data;
*/

/*
   // Parameters for the whisper_full() function
   // If you chnage the order or add new parameters, make sure to update the default values in whisper.cpp:
   // whisper_full_default_params()
   struct whisper_full_params {
       enum whisper_sampling_strategy strategy;

       int n_threads;
       int n_max_text_ctx;
       int offset_ms;          // start offset in ms
       int duration_ms;        // audio duration to process in ms

       bool translate;
       bool no_context;
       bool single_segment;    // force single segment output (useful for streaming)
       bool print_special;
       bool print_progress;
       bool print_realtime;
       bool print_timestamps;

       // [EXPERIMENTAL] token-level timestamps
       bool  token_timestamps; // enable token-level timestamps
       float thold_pt;         // timestamp token probability threshold (~0.01)
       float thold_ptsum;      // timestamp token sum probability threshold (~0.01)
       int   max_len;          // max segment length in characters
       int   max_tokens;       // max tokens per segment (0 = no limit)

       // [EXPERIMENTAL] speed-up techniques
       bool speed_up;          // speed-up the audio by 2x using Phase Vocoder
       int  audio_ctx;         // overwrite the audio context size (0 = use default)

       // tokens to provide the whisper model as initial prompt
       // these are prepended to any existing text context from a previous call
       const whisper_token * prompt_tokens;
       int prompt_n_tokens;

       const char * language;

       struct {
           int n_past;
       } greedy;

       struct {
           int n_past;
           int beam_width;
           int n_best;
       } beam_search;

       whisper_new_segment_callback new_segment_callback;
       void * new_segment_callback_user_data;

       whisper_encoder_begin_callback encoder_begin_callback;
       void * encoder_begin_callback_user_data;
   };

*/

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *Whisper_full_params) SetTranslate(v bool) {
	p.translate = toBool(v)
}

func (p *Whisper_full_params) SetNoContext(v bool) {
	p.no_context = toBool(v)
}

func (p *Whisper_full_params) SetSingleSegment(v bool) {
	p.single_segment = toBool(v)
}

func (p *Whisper_full_params) SetPrintSpecial(v bool) {
	p.print_special = toBool(v)
}

func (p *Whisper_full_params) SetPrintProgress(v bool) {
	p.print_progress = toBool(v)
}

func (p *Whisper_full_params) SetPrintRealtime(v bool) {
	p.print_realtime = toBool(v)
}

func (p *Whisper_full_params) SetPrintTimestamps(v bool) {
	p.print_timestamps = toBool(v)
}

func (p *Whisper_full_params) SetSpeedup(v bool) {
	p.speed_up = toBool(v)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toBool(v bool) C.bool {
	if v {
		return C.bool(true)
	}
	return C.bool(false)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *Whisper_full_params) String() string {
	str := "<whisper.params"
	str += fmt.Sprintf(" strategy=%v", p.strategy)
	str += fmt.Sprintf(" n_threads=%d", p.n_threads)
	if p.language != nil {
		str += fmt.Sprintf(" language=%s", C.GoString(p.language))
	}
	str += fmt.Sprintf(" n_max_text_ctx=%d", p.n_max_text_ctx)
	str += fmt.Sprintf(" offset_ms=%d", p.offset_ms)
	str += fmt.Sprintf(" duration_ms=%d", p.duration_ms)
	if p.translate {
		str += " translate"
	}
	if p.no_context {
		str += " no_context"
	}
	if p.single_segment {
		str += " single_segment"
	}
	if p.print_special {
		str += " print_special"
	}
	if p.print_progress {
		str += " print_progress"
	}
	if p.print_realtime {
		str += " print_realtime"
	}
	if p.print_timestamps {
		str += " print_timestamps"
	}
	if p.token_timestamps {
		str += " token_timestamps"
	}
	if p.speed_up {
		str += " speed_up"
	}

	return str + ">"
}
