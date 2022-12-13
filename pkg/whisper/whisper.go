package whisper

import (
	"os"
	"runtime"
	"time"

	// Packages
	whisper "github.com/djthorpe/go-whisper/sys/whisper"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Model struct {
	ctx        *whisper.Whisper_context
	params     *whisper.Whisper_full_params
	processors int
}

type SegmentCallback func(num int, begin, end time.Duration, tokens []Token)

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	SAMPLE_RATE = whisper.WHISPER_SAMPLE_RATE
	SAMPLE_SIZE = 4 // four bytes per sample
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(path string) (*Model, error) {
	this := new(Model)
	if stat, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrNotFound.With(path)
	} else if err != nil {
		return nil, err
	} else if stat.IsDir() {
		return nil, ErrBadParameter.With(path)
	}
	if ctx := whisper.Whisper_init(path); ctx != nil {
		this.ctx = ctx
	} else {
		return nil, ErrBadParameter.With(path)
	}

	// Set up parameters
	p := this.ctx.Whisper_full_default_params(whisper.WHISPER_SAMPLING_GREEDY)
	this.params = &p
	this.processors = runtime.NumCPU()

	// Return success
	return this, nil
}

func (m *Model) Close() error {
	// Free context
	if m.ctx != nil {
		m.ctx.Whisper_free()
		m.ctx = nil
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set the number of processors to use for parallel processing. If less than zero
// then runtime.NumCPU() is used, if 0 then no parallel processing is used
func (m *Model) SetProcessors(n int) {
	if n < 0 {
		m.processors = runtime.NumCPU()
	} else {
		m.processors = n
	}
}

// Set translate flag
func (m *Model) SetTranslate(v bool) {
	m.params.SetTranslate(v)
}

// Set no context flag
func (m *Model) SetNoContext(v bool) {
	m.params.SetNoContext(v)
}

// Set single segment flag
func (m *Model) SetSingleSegment(v bool) {
	m.params.SetSingleSegment(v)
}

// Set print special flag
func (m *Model) SetPrintSpecial(v bool) {
	m.params.SetPrintSpecial(v)
}

// Set print progress flag
func (m *Model) SetPrintProgress(v bool) {
	m.params.SetPrintProgress(v)
}

// Set print realtime flag
func (m *Model) SetPrintRealtime(v bool) {
	m.params.SetPrintRealtime(v)
}

// Set print timestamps flag
func (m *Model) SetPrintTimestamps(v bool) {
	m.params.SetPrintTimestamps(v)
}

// Set speedup flag
func (m *Model) SetSpeedup(v bool) {
	m.params.SetSpeedup(v)
}

// Transcribe audio from samples
func (m *Model) Process(samples []float32, fn SegmentCallback) error {
	var ret int

	// Run the sequential or parallel version
	if m.processors == 0 {
		ret = m.ctx.Whisper_full(*m.params, samples, m.beginCallback, func(new int) {
			m.segmentCallback(new, fn)
		})
	} else {
		ret = m.ctx.Whisper_full_parallel(*m.params, samples, m.processors, m.beginCallback, func(new int) {
			m.segmentCallback(new, fn)
		})
	}
	// Check for errors
	if ret != 0 {
		return ErrInternalAppError.With("Process")
	}
	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (m *Model) beginCallback() bool {
	return true
}

func (m *Model) segmentCallback(new int, fn SegmentCallback) {
	num_segments := m.ctx.Whisper_full_n_segments()
	s0 := num_segments - new
	for i := s0; i < num_segments; i++ {
		tokens := make([]Token, m.ctx.Whisper_full_n_tokens(i))
		begin := time.Duration(m.ctx.Whisper_full_get_segment_t0(i)) * time.Millisecond * 10.0
		end := time.Duration(m.ctx.Whisper_full_get_segment_t1(i)) * time.Millisecond * 10.0
		for j := 0; j < len(tokens); j++ {
			tokens[j] = Token{
				id:   m.ctx.Whisper_full_get_token_id(i, j),
				text: m.ctx.Whisper_full_get_token_text(i, j),
				p:    m.ctx.Whisper_full_get_token_p(i, j),
			}
		}
		if fn != nil {
			fn(i, begin, end, tokens)
		}
	}
}

/*
    const int n_segments = whisper_full_n_segments(ctx);

    // print the last n_new segments
    const int s0 = n_segments - n_new;
    if (s0 == 0) {
        printf("\n");
    }

    for (int i = s0; i < n_segments; i++) {
        if (params.no_timestamps) {
            if (params.print_colors) {
                for (int j = 0; j < whisper_full_n_tokens(ctx, i); ++j) {
                    if (params.print_special == false) {
                        const whisper_token id = whisper_full_get_token_id(ctx, i, j);
                        if (id >= whisper_token_eot(ctx)) {
                            continue;
                        }
                    }

                    const char * text = whisper_full_get_token_text(ctx, i, j);
                    const float  p    = whisper_full_get_token_p   (ctx, i, j);

                    const int col = std::max(0, std::min((int) k_colors.size(), (int) (std::pow(p, 3)*float(k_colors.size()))));

                    printf("%s%s%s", k_colors[col].c_str(), text, "\033[0m");
                }
            } else {
                const char * text = whisper_full_get_segment_text(ctx, i);
                printf("%s", text);
            }
            fflush(stdout);
        } else {
            const int64_t t0 = whisper_full_get_segment_t0(ctx, i);
            const int64_t t1 = whisper_full_get_segment_t1(ctx, i);

            std::string speaker = "";

            if (params.diarize && pcmf32s.size() == 2) {
                const int64_t n_samples = pcmf32s[0].size();

                const int64_t is0 = timestamp_to_sample(t0, n_samples);
                const int64_t is1 = timestamp_to_sample(t1, n_samples);

                double energy0 = 0.0f;
                double energy1 = 0.0f;

                for (int64_t j = is0; j < is1; j++) {
                    energy0 += fabs(pcmf32s[0][j]);
                    energy1 += fabs(pcmf32s[1][j]);
                }

                if (energy0 > 1.1*energy1) {
                    speaker = "(speaker 0)";
                } else if (energy1 > 1.1*energy0) {
                    speaker = "(speaker 1)";
                } else {
                    speaker = "(speaker ?)";
                }

                //printf("is0 = %lld, is1 = %lld, energy0 = %f, energy1 = %f, %s\n", is0, is1, energy0, energy1, speaker.c_str());
            }

            if (params.print_colors) {
                printf("[%s --> %s]  ", to_timestamp(t0).c_str(), to_timestamp(t1).c_str());
                for (int j = 0; j < whisper_full_n_tokens(ctx, i); ++j) {
                    if (params.print_special == false) {
                        const whisper_token id = whisper_full_get_token_id(ctx, i, j);
                        if (id >= whisper_token_eot(ctx)) {
                            continue;
                        }
                    }

                    const char * text = whisper_full_get_token_text(ctx, i, j);
                    const float  p    = whisper_full_get_token_p   (ctx, i, j);

                    const int col = std::max(0, std::min((int) k_colors.size(), (int) (std::pow(p, 3)*float(k_colors.size()))));

                    printf("%s%s%s%s", speaker.c_str(), k_colors[col].c_str(), text, "\033[0m");
                }
                printf("\n");
            } else {
                const char * text = whisper_full_get_segment_text(ctx, i);

                printf("[%s --> %s]  %s%s\n", to_timestamp(t0).c_str(), to_timestamp(t1).c_str(), speaker.c_str(), text);
            }
        }
    }
}
*/
