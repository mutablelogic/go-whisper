package main

import (
	"fmt"

	"github.com/djthorpe/go-tablewriter"
)

type DownloadCmd struct {
	Model string `arg:"" name:"model" help:"Model to download (must end in .bin)"`
}

func (cmd *DownloadCmd) Run(ctx *Globals) error {
	model, err := ctx.client.DownloadModel(ctx.ctx, cmd.Model, func(status string, cur, total int64) {
		pct := fmt.Sprintf("%02d%%", int(100*float64(cur)/float64(total)))
		ctx.writer.Writeln(pct, status)
	})
	if err != nil {
		return err
	}
	return ctx.writer.Write(model, tablewriter.OptHeader())
}
