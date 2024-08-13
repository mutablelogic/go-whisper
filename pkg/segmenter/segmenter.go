package segmenter

import (
	"context"
	"errors"
	"io"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

// A segmenter reads audio samples from a reader and segments them into
// fixed-size chunks. The segmenter can be used to process audio samples
type Segmenter struct {
	ts          time.Duration
	sample_rate int
	n           int
	buf         []float32
	reader      *ffmpeg.Reader
}

// SegmentFunc is a callback function which is called when a segment is ready
// to be processed. The first argument is the timestamp of the segment.
type SegmentFunc func(time.Duration, []float32) error

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new segmenter with a reader r which segments raw audio of 'dur'
// length. If dur is zero then no segmenting is performed, the whole
// audio file is read, which could cause some memory issues.
//
// The sample rate is the number of samples per second.
//
// At the moment, the audio format is auto-detected, but there should be
// a way to specify the audio format.
func NewReader(r io.Reader, dur time.Duration, sample_rate int) (*Segmenter, error) {
	segmenter := new(Segmenter)

	// Check arguments
	if dur < 0 || sample_rate <= 0 {
		return nil, ErrBadParameter.With("invalid duration or sample rate arguments")
	} else {
		segmenter.sample_rate = sample_rate
	}

	// Sample buffer is duration * sample rate
	if dur > 0 {
		segmenter.n = int(dur.Seconds() * float64(sample_rate))
		segmenter.buf = make([]float32, 0, segmenter.n)
	}

	// Open the file
	media, err := ffmpeg.NewReader(r)
	if err != nil {
		return nil, err
	} else {
		segmenter.reader = media
	}

	return segmenter, nil
}

// Close the segmenter
func (s *Segmenter) Close() error {
	var result error

	if s.reader != nil {
		result = errors.Join(result, s.reader.Close())
	}
	s.reader = nil
	s.buf = nil

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Segments are output through a callback, with the samples and a timestamp
// TODO: we could do some basic silence and voice detection to segment to ensure
// we don't overtax the CPU/GPU with silence and non-speech
// TODO: We whould be able to select the audio stream to use. At the moment
// the "best" audio stream is used, based on ffmpeg heuristic.
func (s *Segmenter) Decode(ctx context.Context, fn SegmentFunc) error {
	// Check input parameters
	if fn == nil {
		return ErrBadParameter.With("SegmentFunc is nil")
	}

	// Map function chooses the best audio stream
	mapFunc := func(stream int, params *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == s.reader.BestStream(media.AUDIO) {
			return ffmpeg.NewAudioPar("flt", "mono", s.sample_rate)
		}
		// Ignore no-audio streams
		return nil, nil
	}

	// Decode samples and segment
	if err := s.reader.Decode(ctx, mapFunc, func(stream int, frame *ffmpeg.Frame) error {
		// We get null frames sometimes, ignore them
		if frame == nil {
			return nil
		}

		// Append float32 samples from plane 0 to buffer
		s.buf = append(s.buf, frame.Float32(0)...)

		// n != 0 and len(buf) >= n we have a segment to process
		if s.n != 0 && len(s.buf) >= s.n {
			if err := s.segment(fn); err != nil {
				return err
			}
			// Increment the timestamp
			s.ts += time.Duration(len(s.buf)) * time.Second / time.Duration(s.sample_rate)
			// Clear the buffer
			s.buf = s.buf[:0]
		}

		// Continue processing
		return nil
	}); err != nil {
		return err
	}

	// Output any remaining samples
	if len(s.buf) > 0 {
		if err := s.segment(fn); err != nil {
			return err
		}
	}

	// Return success
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (s *Segmenter) segment(fn SegmentFunc) error {
	// Not segmenting
	if s.n == 0 {
		return fn(s.ts, s.buf)
	}

	// Split into n-sized segments
	bufLength := len(s.buf)
	ts := s.ts
	tsinc := time.Duration(s.n) * time.Second / time.Duration(s.sample_rate)
	for i := 0; i < bufLength; i += s.n {
		end := i + s.n
		var segment []float32
		if end <= bufLength {
			// If the segment fits exactly or there are enough items
			segment = s.buf[i:end]
		} else {
			// If the segment is smaller than segmentSize, pad with zeros
			segment = make([]float32, s.n)
			copy(segment, s.buf[i:bufLength])
		}
		if err := fn(ts, segment); err != nil {
			return err
		} else {
			ts += tsinc
		}
	}

	// Return success
	return nil
}
