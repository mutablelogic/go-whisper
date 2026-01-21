package main

import (
	"fmt"
	"io"
	"os"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httpclient "github.com/mutablelogic/go-whisper/pkg/httpclient"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TranscribeCommands struct {
	Transcribe TranscribeCommand `cmd:"" name:"transcribe" help:"Transcribe audio file." group:"TRANSCRIBE & TRANSLATE"`
}

type TranscribeCommand struct {
	TranslateCommand
	Diarize  *bool   `name:"diarize" help:"Enable speaker diarization"`
	Language *string `name:"language" help:"Language code (e.g., 'en', 'es', 'fr')"`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *TranscribeCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// OTEL
	parent, endSpan := otel.StartSpan(ctx.tracer, ctx.ctx, "TranscribeCommand")
	defer func() { endSpan(err) }()

	// Open audio file
	file, err := os.Open(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Build options
	opts := []httpclient.Opt{}
	if cmd.Language != nil {
		opts = append(opts, httpclient.WithLanguage(*cmd.Language))
	}
	if cmd.Prompt != nil {
		opts = append(opts, httpclient.WithPrompt(*cmd.Prompt))
	}
	if cmd.Temperature != nil {
		opts = append(opts, httpclient.WithTemperature(*cmd.Temperature))
	}
	if cmd.Diarize != nil {
		opts = append(opts, httpclient.WithDiarize(*cmd.Diarize))
	}

	// Set format
	format, err := formatFromString(cmd.Format)
	if err != nil {
		return err
	}

	// Add real-time segment printing callback
	opts = append(opts, httpclient.WithSegmentCallback(func(seg *schema.Segment) error {
		writeSegment(os.Stdout, seg, format)
		return nil
	}))

	// Transcribe
	var result *schema.Transcription
	result, err = client.Transcribe(parent, cmd.Model, file, opts...)
	if err != nil {
		return err
	}

	// If segments were not printed via streaming, print from result
	for _, seg := range result.Segments {
		writeSegment(os.Stdout, seg, format)
	}
	writeTrailer(os.Stdout, format)

	// Return success
	return nil
}

// Method to write segment in specified format
func writeSegment(w io.Writer, seg *schema.Segment, format httpclient.FormatType) {
	switch format {
	case httpclient.FormatVTT:
		seg.WriteVTT(w, 0)
	case httpclient.FormatSRT:
		seg.WriteSRT(w, 0)
	case httpclient.FormatJSON:
		seg.WriteJSON(w)
	default:
		seg.WriteText(w)
	}
}

// Method to write a trailer in specified format
func writeTrailer(w io.Writer, format httpclient.FormatType) {
	switch format {
	case httpclient.FormatJSON:
		schema.WriteJSONTrailer(w)
	default:
		schema.WriteTextTrailer(w)
	}
}

func formatFromString(format string) (httpclient.FormatType, error) {
	switch format {
	case "text", string(httpclient.FormatText):
		return httpclient.FormatText, nil
	case "vtt", string(httpclient.FormatVTT):
		return httpclient.FormatVTT, nil
	case "srt", string(httpclient.FormatSRT):
		return httpclient.FormatSRT, nil
	case "json", string(httpclient.FormatJSON):
		return httpclient.FormatJSON, nil
	default:
		return "", httpresponse.ErrBadRequest.Withf("unsupported format: %q", format)
	}
}
