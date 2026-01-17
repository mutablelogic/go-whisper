package main

import (
	"fmt"
	"os"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	httpclient "github.com/mutablelogic/go-whisper/pkg/httpclient"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TranscribeCommands struct {
	Transcribe TranscribeCommand `cmd:"" name:"transcribe" help:"Transcribe audio file." group:"TRANSCRIBE & TRANSLATE"`
}

type TranscribeCommand struct {
	Model       string   `arg:"" name:"model" help:"Model ID to use for transcription"`
	File        string   `arg:"" name:"file" help:"Audio file to transcribe"`
	Language    *string  `name:"language" help:"Language code (e.g., 'en', 'es', 'fr')"`
	Prompt      *string  `name:"prompt" help:"Initial prompt to guide transcription"`
	Temperature *float64 `name:"temperature" help:"Temperature (0.0-1.0)"`
	Diarize     *bool    `name:"diarize" help:"Enable speaker diarization"`
	Format      string   `name:"format" help:"Output format: json, text, vtt, srt" default:"json"`
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
	var format httpclient.FormatType
	switch cmd.Format {
	case "text", string(httpclient.FormatText):
		format = httpclient.FormatText
	case "vtt", string(httpclient.FormatVTT):
		format = httpclient.FormatVTT
	case "srt", string(httpclient.FormatSRT):
		format = httpclient.FormatSRT
	case "json", string(httpclient.FormatJSON):
		format = httpclient.FormatJSON
	default:
		return fmt.Errorf("unsupported format: %s", cmd.Format)
	}
	opts = append(opts, httpclient.WithFormat(format))

	// Transcribe
	var result *schema.Transcription
	result, err = client.Transcribe(parent, cmd.Model, file, opts...)
	if err != nil {
		return err
	}

	// Print result - if non-JSON format, just print the text
	if format != httpclient.FormatJSON {
		fmt.Print(result.Text)
	} else {
		fmt.Println(result)
	}
	return nil
}
