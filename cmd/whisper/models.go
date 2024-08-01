package main

import (
	// Packages
	"github.com/djthorpe/go-tablewriter"
)

type ModelsCmd struct{}

func (*ModelsCmd) Run(ctx *Globals) error {
	return ctx.writer.Write(ctx.service.ListModels(), tablewriter.OptHeader())
}
