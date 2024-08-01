package main

import (
	"log"

	// Packages
	"github.com/djthorpe/go-tablewriter"
)

type DownloadCmd struct {
	Model string `arg:"" help:"Model to download"`
}

func (cmd *DownloadCmd) Run(ctx *Globals) error {
	model, err := ctx.service.DownloadModel(ctx.ctx, cmd.Model, func(curBytes, totalBytes uint64) {
		log.Printf("Downloaded %d of %d bytes", curBytes, totalBytes)
	})
	if err != nil {
		return err
	}
	return ctx.writer.Write(model, tablewriter.OptHeader())
}
