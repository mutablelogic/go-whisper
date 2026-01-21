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

type TranslateCommands struct {
	Translate TranslateCommand `cmd:"" name:"translate" help:"Translate audio file to English." group:"TRANSCRIBE & TRANSLATE"`
}

type TranslateCommand struct {
	Model       string   `arg:"" name:"model" help:"Model ID to use for translation"`
	File        string   `arg:"" name:"file" help:"Audio file to translate"`
	Prompt      *string  `name:"prompt" help:"Initial prompt to guide translation"`
	Temperature *float64 `name:"temperature" help:"Temperature (0.0-1.0)"`
	Format      string   `name:"format" help:"Output format: json, text, vtt, srt" default:"json"`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *TranslateCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// OTEL
	parent, endSpan := otel.StartSpan(ctx.tracer, ctx.ctx, "TranslateCommand")
	defer func() { endSpan(err) }()

	// Open audio file
	file, err := os.Open(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Build options
	opts := []httpclient.Opt{}
	if cmd.Prompt != nil {
		opts = append(opts, httpclient.WithPrompt(*cmd.Prompt))
	}
	if cmd.Temperature != nil {
		opts = append(opts, httpclient.WithTemperature(*cmd.Temperature))
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

	// Translate
	var result *schema.Transcription
	result, err = client.Translate(parent, cmd.Model, file, opts...)
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
