package segmenter

import (
	"context"
	"errors"
	"io"
	"time"

	// Packages

	"github.com/mutablelogic/go-media/pkg/ffmpeg"
	"github.com/mutablelogic/go-whisper/sys/rnnoise"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Denoiser struct {
	denoiser  *rnnoise.DenoiseState
	re        *ffmpeg.Re
	frame     *ffmpeg.Frame
	segmenter *Segmenter
}

// DenoiseFunc is a callback function which is called when a segment has been denoised.
// The first argument is the timestamp of the segment, the second argument is the
// probability of speech in the segment, and the third argument is the denoised
// audio frame
type DenoiseFunc func(time.Duration, float32, *ffmpeg.Frame) error

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new denoiser with an audio reader. You should release resources
// by calling Close() on the denoiser.
func NewDenoiseReader(r io.Reader, sample_format string, sample_rate int) (*Denoiser, error) {
	d := new(Denoiser)
	dur := float64(rnnoise.Rnnoise_get_frame_size()) * float64(time.Second) / float64(rnnoise.SampleRate)
	inpar := ffmpeg.AudioPar("fltp", "mono", rnnoise.SampleRate)
	if ctx, err := rnnoise.Rnnoise_create(nil); err != nil {
		return nil, err
	} else if segmenter, err := NewReader(r, time.Duration(dur), rnnoise.SampleRate); err != nil {
		rnnoise.Rnnoise_destroy(ctx)
		return nil, err
	} else if outpar, err := ffmpeg.NewAudioPar(sample_format, "mono", sample_rate); err != nil {
		rnnoise.Rnnoise_destroy(ctx)
		return nil, errors.Join(err, segmenter.Close())
	} else if frame, err := ffmpeg.NewFrame(inpar); err != nil {
		rnnoise.Rnnoise_destroy(ctx)
		return nil, errors.Join(err, segmenter.Close())
	} else if re, err := ffmpeg.NewRe(outpar, false); err != nil {
		rnnoise.Rnnoise_destroy(ctx)
		return nil, errors.Join(err, segmenter.Close(), frame.Close())
	} else {
		d.frame = frame
		d.re = re
		d.segmenter = segmenter
		d.denoiser = ctx
	}

	// Return success
	return d, nil
}

// Release resources associated with the denoiser
func (d *Denoiser) Close() error {
	var result error

	// Free resources
	if d.frame != nil {
		result = errors.Join(result, d.frame.Close())
	}
	if d.re != nil {
		result = errors.Join(result, d.re.Close())
	}
	if d.denoiser != nil {
		rnnoise.Rnnoise_destroy(d.denoiser)
	}
	if d.segmenter != nil {
		result = errors.Join(result, d.segmenter.Close())
	}

	// Clear resources
	d.frame = nil
	d.re = nil
	d.denoiser = nil
	d.segmenter = nil

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
		// Process the frame
		p := rnnoise.Rnnoise_process_frame(d.denoiser, data)

		// Resample the frame
		d.frame.SetTs(ts.Seconds())
		d.frame.SetFloat32(0, data)
		if dst, err := d.re.Frame(d.frame); err != nil {
			return err
		} else if dst != nil {
			return fn(ts, p, dst)
		} else {
			return nil
		}
	})
}
