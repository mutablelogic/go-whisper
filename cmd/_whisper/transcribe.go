package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	// Packages
	goclient "github.com/mutablelogic/go-client"
	segmenterPkg "github.com/mutablelogic/go-media/pkg/segmenter"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
	whisper "github.com/mutablelogic/go-whisper"
	client "github.com/mutablelogic/go-whisper/pkg/client"
	openai "github.com/mutablelogic/go-whisper/pkg/client/openai"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisperPkg "github.com/mutablelogic/go-whisper/pkg/whisper"
	wav "github.com/mutablelogic/go-whisper/pkg/wav"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type TranslateCmd struct {
	Model       string        `arg:"" help:"Model to use"`
	Path        string        `arg:"" help:"Path to audio file"`
	Format      string        `flag:"" help:"Output format" default:"text" enum:"json,verbose_json,text,vtt,srt"`
	Segments    time.Duration `flag:"" help:"Segment size for reading audio file"`
	Silence     time.Duration `flag:"" help:"Segment silence threshold"`
	Remote      bool          `flag:"" help:"Use remote service (gowhisper, openai, elevenlabs) for translation or transcription"`
	Temperature *float64      `flag:"" help:"Temperature"`
	Diarize     bool          `flag:"" help:"Diarize the transcription"`
	Stream      bool          `flag:"" help:"Stream the transcription results"`
	Language    string        `flag:"language" help:"Language to transcribe"`
	Prompt      *string       `flag:"prompt" help:"Prompt to guide the model's style or continue a previous audio segment"`
}

type TranscribeCmd struct {
	TranslateCmd
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *TranscribeCmd) Run(app *Globals) error {
	if cmd.Remote {
		return cmd.TranslateCmd.run_remote(app, false)
	} else {
		return cmd.TranslateCmd.run_local(app, false)
	}
}

func (cmd *TranslateCmd) Run(app *Globals) error {
	if cmd.Remote {
		return cmd.run_remote(app, true)
	} else {
		return cmd.run_local(app, true)
	}
}

func (cmd *TranslateCmd) run_local(app *Globals, translate bool) error {
	// Get the model
	model_ := app.service.GetModelById(cmd.Model)
	if model_ == nil {
		return httpresponse.ErrNotFound.With(cmd.Model)
	}

	// Open the audio file
	f, err := os.Open(cmd.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a segmenter - read segments based on requested segment size
	opts := []segmenterPkg.Opt{}
	if cmd.Segments > 0 {
		opts = append(opts, segmenterPkg.WithSegmentSize(cmd.Segments))
	}
	segmenter, err := segmenterPkg.NewFromReader(f, whisper.SampleRate, opts...)
	if err != nil {
		return err
	}
	defer segmenterPkg.Close()

	// Perform the transcription
	return app.service.WithModel(model_, func(task *whisperPkg.Task) error {
		// Transcribe or Translate
		task.SetTranslate(translate)
		task.SetDiarize(cmd.Diarize)

		// Set language
		if cmd.Language != "" {
			if err := task.SetLanguage(cmd.Language); err != nil {
				return err
			}
		}
		// Set temperature
		if cmd.Temperature != nil {
			if err := task.SetTemperature(*cmd.Temperature); err != nil {
				return err
			}
		}
		// Set prompt
		if cmd.Prompt != nil {
			if err := task.SetPrompt(*cmd.Prompt); err != nil {
				return err
			}
		}

		// Read samples and transcribe them
		if err := segmenterPkg.DecodeFloat32(app.ctx, func(ts time.Duration, buf []float32) error {
			// Perform the transcription, return any errors
			return task.Transcribe(app.ctx, ts, buf, func(segment *schema.Segment) {
				var buf bytes.Buffer
				switch cmd.Format {
				case "json", "verbose_json":
					fmt.Println(segment)
				case "srt":
					whisperPkg.WriteSegmentSrt(&buf, segment)
					fmt.Println(buf.String())
				case "vtt":
					if segment.Id == 0 {
						fmt.Println("WEBVTT" + "\n")
					}
					whisperPkg.WriteSegmentVtt(&buf, segment)
					fmt.Println(buf.String())
				case "text":
					whisperPkg.WriteSegmentText(&buf, segment)
					fmt.Println(buf.String())
				}
			})
		}); err != nil {
			return err
		}

		return nil
	})
}

func (cmd *TranslateCmd) run_remote(app *Globals, translate bool) error {
	// Open the audio file
	f, err := os.Open(cmd.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a client for the whisper service
	opts := []goclient.ClientOpt{
		goclient.OptTimeout(5 * time.Minute), // Set a timeout for the request
	}
	if app.Debug {
		opts = append(opts, goclient.OptTrace(os.Stderr, true))
	}
	remote, err := client.New(opts...)
	if err != nil {
		return err
	}

	// Create an array of parameters for the transcription
	params := []client.Opt{
		client.OptPath("audio.wav"), client.OptFormat(openai.FormatJson), client.OptLanguage(cmd.Language),
	}
	if cmd.Diarize {
		params = append(params, client.OptDiarize())
	}
	if cmd.Stream {
		params = append(params, client.OptStream(func(evt schema.Event) {
			// TODO: Delta will be text or json depending on the format
			fmt.Println("event:", evt)
		}))
	}
	if cmd.Temperature != nil {
		params = append(params, client.OptTemperature(types.PtrFloat64(cmd.Temperature)))
	}
	if cmd.Prompt != nil {
		params = append(params, client.OptPrompt(types.PtrString(cmd.Prompt)))
	}

	// Create a segmenter - read segments based on requested segment size
	sopts := []segmenterPkg.Opt{}
	if cmd.Segments > 0 {
		sopts = append(sopts, segmenterPkg.WithSegmentSize(cmd.Segments))
	}
	if cmd.Silence > 0 {
		sopts = append(sopts, segmenterPkg.WithDefaultSilence())
		sopts = append(sopts, segmenterPkg.WithSilenceSize(cmd.Silence))
	}
	splitter, err := segmenterPkg.NewFromReader(f, whisper.SampleRate, sopts...)
	if err != nil {
		return err
	}
	defer splitter.Close()

	// Read samples and transcribe or translate them
	return splitter.DecodeInt16(app.ctx, func(ts time.Duration, data []int16) error {
		// Make a mono WAV file from the float32 samples
		r, err := wav.NewInt16(data, whisper.SampleRate, 1)
		if err != nil {
			return err
		}

		var segments []*schema.Segment
		if translate {
			translation, err := remote.Translate(app.ctx, cmd.Model, r, params...)
			if err != nil {
				return err
			} else {
				segments = translation.Segments
			}
		} else {
			transcription, err := remote.Transcribe(app.ctx, cmd.Model, r, params...)
			if err != nil {
				return err
			} else {
				segments = transcription.Segments
			}
		}

		// Write the segments to the writer
		for _, segment := range segments {
			var buf bytes.Buffer
			switch cmd.Format {
			case "json", "verbose_json":
				fmt.Println(segment)
			case "srt":
				segment.WriteSRT(&buf, ts)
				fmt.Println(buf.String())
			case "vtt":
				segment.WriteVTT(&buf, ts)
				fmt.Println(buf.String())
			case "text":
				segment.WriteText(&buf)
				fmt.Println(buf.String())
			}
		}

		return nil
	})
}
