package main

import (
	"os"

	"github.com/djthorpe/go-tablewriter"
)

type TranscribeCmd struct {
	Model string `arg:"" required:"" help:"Model Identifier" type:"string"`
	Path  string `arg:"" required:"" help:"Audio File Path" type:"string"`
}

func (cmd *TranscribeCmd) Run(ctx *Globals) error {
	r, err := os.Open(cmd.Path)
	if err != nil {
		return err
	}
	defer r.Close()

	transcription, err := ctx.api.Transcribe(ctx.ctx, cmd.Model, r)
	if err != nil {
		return err
	}
	return ctx.writer.Write(transcription, tablewriter.OptHeader())
}
