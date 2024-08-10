package main

import "github.com/djthorpe/go-tablewriter"

type ModelsCmd struct{}

func (cmd *ModelsCmd) Run(ctx *Globals) error {
	if models, err := ctx.client.ListModels(ctx.ctx); err != nil {
		return err
	} else {
		return ctx.writer.Write(models, tablewriter.OptHeader())
	}
}
