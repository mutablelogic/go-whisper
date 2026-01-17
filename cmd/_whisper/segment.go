package main

import (
	"fmt"
	"io"
	"os"
	"time"

	// Packages
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
	whisper "github.com/mutablelogic/go-whisper"
	wav "github.com/mutablelogic/go-whisper/pkg/wav"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type SegmentCmd struct {
	Path     string        `arg:"" help:"Path to audio file"`
	Segments time.Duration `flag:"" help:"Segment size for reading audio file"`
	Silence  time.Duration `flag:"" help:"Segment silence threshold"`
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *SegmentCmd) Run(app *Globals) error {
	// Open the audio file
	f, err := os.Open(cmd.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a segmenter - read segments based on requested segment size
	opts := []segmenter.Opt{}
	if cmd.Segments > 0 {
		opts = append(opts, segmenter.WithSegmentSize(cmd.Segments))
	}
	if cmd.Silence > 0 {
		opts = append(opts, segmenter.WithDefaultSilenceThreshold())
		opts = append(opts, segmenter.WithSilenceSize(cmd.Silence))
	}
	segmenter, err := segmenter.NewReader(f, whisper.SampleRate, opts...)
	if err != nil {
		return err
	}
	defer segmenter.Close()

	// Read samples and transcribe them
	return segmenter.DecodeInt16(app.ctx, func(ts time.Duration, data []int16) error {
		// Make a WAV file from the float32 samples
		r, err := wav.NewInt16(data, whisper.SampleRate, 1)
		if err != nil {
			return err
		}

		fmt.Println("Writing segment at", ts)

		w, err := os.Create("segment-" + fmt.Sprint(ts.Seconds()) + ".wav")
		if err != nil {
			return err
		}
		defer w.Close()

		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
		return nil
	})
}
