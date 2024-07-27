package main

import tablewriter "github.com/djthorpe/go-tablewriter"

type ModelsCmd struct{}

func (_ *ModelsCmd) Run(ctx *Globals) error {
	models, err := ctx.api.ListModels(ctx.ctx)
	if err != nil {
		return err
	}
	return ctx.writer.Write(models, tablewriter.OptHeader())
}
