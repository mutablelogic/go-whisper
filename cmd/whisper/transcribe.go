package main

import (
	"os"
	"time"

	// Packages
	whisper "github.com/mutablelogic/go-whisper"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	segmenter "github.com/mutablelogic/go-whisper/pkg/segmenter"
	task "github.com/mutablelogic/go-whisper/pkg/task"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

type TranscribeCmd struct {
	Model    string `arg:"" help:"Model to use"`
	Path     string `arg:"" help:"Path to audio file"`
	Language string `flag:"language" help:"Language to transcribe" default:"auto"`
	Format   string `flag:"format" help:"Output format" default:"text" enum:"json,verbose_json,text,vtt,srt"`
}

func (cmd *TranscribeCmd) Run(ctx *Globals) error {
	// Get the model
	model := ctx.service.GetModelById(cmd.Model)
	if model == nil {
		return ErrNotFound.With(cmd.Model)
	}

	// Open the audio file
	f, err := os.Open(cmd.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a segmenter - read segments based on requested segment size
	// TODO
	segmenter, err := segmenter.NewReader(f, 0, whisper.SampleRate)
	if err != nil {
		return err
	}
	defer segmenter.Close()

	// Perform the transcription
	return ctx.service.WithModel(model, func(taskctx *task.Context) error {
		// Transcribe
		taskctx.SetTranslate(false)
		taskctx.SetDiarize(false)

		// Set language
		if cmd.Language != "" {
			if err := taskctx.SetLanguage(cmd.Language); err != nil {
				return err
			}
		}

		// Read samples and transcribe them
		if err := segmenter.Decode(ctx.ctx, func(ts time.Duration, buf []float32) error {
			// Perform the transcription, return any errors
			return taskctx.Transcribe(ctx.ctx, ts, buf, func(segment *schema.Segment) {
				ctx.writer.Write(segment)
			})
		}); err != nil {
			return err
		}

		return nil
	})
}
