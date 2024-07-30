package main

import (
	"os"
	"time"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-whisper/pkg/whisper/client"
)

type TranscribeCmd struct {
	Model       string         `arg:"" required:"" help:"Model Identifier" type:"string"`
	Path        string         `arg:"" required:"" help:"Audio File Path" type:"string"`
	Language    string         `flag:"language" help:"Source Language" type:"string"`
	SegmentSize *time.Duration `flag:"segment-size" help:"Segment Size" type:"duration"`
	ResponseFmt *string        `flag:"format" help:"Response Format" enum:"json,verbose_json,text,vtt,srt"`
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
	if cmd.SegmentSize != nil {
		opts = append(opts, client.OptSegmentSize(*cmd.SegmentSize))
	}
	if cmd.ResponseFmt != nil {
		opts = append(opts, client.OptResponseFormat(*cmd.ResponseFmt))
	}

	transcription, err := ctx.api.Transcribe(ctx.ctx, cmd.Model, r, opts...)
	if err != nil {
		return err
	}
	return ctx.writer.Write(transcription, tablewriter.OptHeader())
}
