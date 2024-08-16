package segmenter

import (
	"context"
	"errors"
	"io"
	"time"

	// Packages
	"github.com/mutablelogic/go-whisper/sys/rnnoise"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Denoiser struct {
	denoiser  *rnnoise.DenoiseState
	segmenter *Segmenter
}

// DenoiseFunc is a callback function which is called when a segment has been denoised.
// The first argument is the timestamp of the segment, the second argument is the
// probability of speech in the segment, and the third argument is the denoised
// mono audio samples at 48kHz
type DenoiseFunc func(time.Duration, float32, []float32) error

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new denoiser with an audio reader. You should release resources
// by calling Close() on the denoiser.
func NewDenoiseReader(r io.Reader) (*Denoiser, error) {
	d := new(Denoiser)
	dur := float64(rnnoise.Rnnoise_get_frame_size()) * float64(time.Second) / float64(rnnoise.SampleRate)
	if ctx, err := rnnoise.Rnnoise_create(nil); err != nil {
		return nil, err
	} else if segmenter, err := NewReader(r, time.Duration(dur), rnnoise.SampleRate); err != nil {
		rnnoise.Rnnoise_destroy(ctx)
		return nil, err
	} else {
		d.denoiser = ctx
		d.segmenter = segmenter
	}

	// Return success
	return d, nil
}

// Release resources associated with the denoiser
func (d *Denoiser) Close() error {
	var result error

	if d.denoiser != nil {
		rnnoise.Rnnoise_destroy(d.denoiser)
	}
	if d.segmenter != nil {
		result = errors.Join(result, d.segmenter.Close())
	}

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Decode the audio stream, segment and denoise it, then call fn for each segment.
func (d *Denoiser) Decode(ctx context.Context, fn DenoiseFunc) error {
	if fn == nil {
		return errors.New("no function specified")
	}
	return d.segmenter.Decode(ctx, func(ts time.Duration, data []float32) error {
		return fn(ts, rnnoise.Rnnoise_process_frame(d.denoiser, data), data)
	})
}
