package segmenter

import (
	"context"
	"errors"
	"fmt"
	"io"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

type Segmenter struct {
	reader *ffmpeg.Reader
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new segmenter for "NumSamples" with a reader r
// If NumSamples is zero then no segmenting is performed
func NewSegmenter(r io.Reader) (*Segmenter, error) {
	segmenter := new(Segmenter)

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

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// TODO: segments are output through a callback, with the samples and a timestamp
// TODO: we could do some basic silence and voice detection to segment to ensure
// we don't overtax the CPU/GPU with silence and non-speech
// TODO: We have hard-coded the sample format, sample rate and number of channels
// here. We should make this configurable
func (s *Segmenter) Decode(ctx context.Context) error {
	mapFunc := func(stream int, params *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == s.reader.BestStream(media.AUDIO) {
			return ffmpeg.NewAudioPar("flt", "mono", 16000)
		}
		// Ignore no-audio streams
		return nil, nil
	}
	return s.reader.Decode(ctx, mapFunc, func(stream int, frame *ffmpeg.Frame) error {
		// Append float32 samples to buffer
		fmt.Println("TODO: Implement Decode", frame)
		return nil
	})
}
