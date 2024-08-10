package main

import (
	"errors"

	// Packages
	"github.com/djthorpe/go-tablewriter"
)

type ModelsCmd struct{}

func (ModelsCmd) Run(ctx *Globals) error {
	models := ctx.service.ListModels()
	if len(models) == 0 {
		return errors.New("no models found")
	} else {
		return ctx.writer.Write(ctx.service.ListModels(), tablewriter.OptHeader())
	}
}
