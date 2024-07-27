package main

import (
	"os"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-whisper/pkg/client"
)

type TranscribeCmd struct {
	Model       string   `arg:"" required:"" help:"Model Identifier" type:"string"`
	Path        string   `arg:"" required:"" help:"Audio File Path" type:"string"`
	Language    string   `flag:"language" help:"Source Language" type:"string"`
	Prompt      string   `flag:"prompt" help:"Initial Prompt Identifier" type:"string"`
	Temperature *float32 `flag:"temperature" help:"Temperature" type:"float32"`
}

func (cmd *TranscribeCmd) Run(ctx *Globals) error {
	r, err := os.Open(cmd.Path)
	if err != nil {
		return err
	}
	defer r.Close()

	opts := []client.Opt{}
	if cmd.Language != "" {
		opts = append(opts, client.OptLanguage(cmd.Language))
	}
	if cmd.Prompt != "" {
		opts = append(opts, client.OptPrompt(cmd.Prompt))
	}
	if cmd.Temperature != nil {
		opts = append(opts, client.OptTemperature(*cmd.Temperature))
	}

	transcription, err := ctx.api.Transcribe(ctx.ctx, cmd.Model, r, opts...)
	if err != nil {
		return err
	}
	return ctx.writer.Write(transcription, tablewriter.OptHeader())
}
