package main

import (
	"fmt"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ModelCommands struct {
	ListModels    ListModelsCommand    `cmd:"" name:"models" help:"List models." group:"MODEL"`
	GetModel      GetModelCommand      `cmd:"" name:"model" help:"Get model." group:"MODEL"`
	DownloadModel DownloadModelCommand `cmd:"" name:"download-model" help:"Download model." group:"MODEL"`
	DeleteModel   DeleteModelCommand   `cmd:"" name:"delete-model" help:"Delete model." group:"MODEL"`
}

type ListModelsCommand struct{}

type GetModelCommand struct {
	ID string `arg:"" name:"id" help:"Model ID"`
}

type DownloadModelCommand struct {
	ID string `arg:"" name:"id" help:"Model ID or path"`
}

type DeleteModelCommand struct {
	ID string `arg:"" name:"id" help:"Model ID"`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *ListModelsCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// OTEL
	parent, endSpan := otel.StartSpan(ctx.tracer, ctx.ctx, "ListModelsCommand")
	defer func() { endSpan(err) }()

	// List models
	models, err := client.ListModels(parent)
	if err != nil {
		return err
	}

	// Print
	fmt.Println(models)
	return nil
}

func (cmd *GetModelCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// OTEL
	parent, endSpan := otel.StartSpan(ctx.tracer, ctx.ctx, "GetModelCommand")
	defer func() { endSpan(err) }()

	// Get model
	model, err := client.GetModel(parent, cmd.ID)
	if err != nil {
		return err
	}

	// Print
	fmt.Println(model)
	return nil
}

func (cmd *DownloadModelCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// OTEL
	parent, endSpan := otel.StartSpan(ctx.tracer, ctx.ctx, "DownloadModelCommand")
	defer func() { endSpan(err) }()

	// Download model with progress reporting
	var lastPercent int = -1
	model, err := client.DownloadModel(parent, cmd.ID, func(cur, total uint64) {
		if total > 0 {
			percent := int(float64(cur) / float64(total) * 100)
			if percent != lastPercent {
				lastPercent = percent
				fmt.Printf("\rDownloading: %d%% (%d/%d bytes)", percent, cur, total)
			}
		} else {
			fmt.Printf("\rDownloading: %d bytes", cur)
		}
	})
	if err != nil {
		fmt.Println() // New line after progress
		return err
	}

	// Print result
	fmt.Println() // New line after progress
	fmt.Println(model)
	return nil
}

func (cmd *DeleteModelCommand) Run(ctx *Globals) (err error) {
	client, err := ctx.Client()
	if err != nil {
		return err
	}

	// OTEL
	parent, endSpan := otel.StartSpan(ctx.tracer, ctx.ctx, "DeleteModelCommand")
	defer func() { endSpan(err) }()

	// Delete model
	err = client.DeleteModel(parent, cmd.ID)
	if err != nil {
		return err
	}

	// Print success message
	fmt.Printf("Model %s deleted successfully\n", cmd.ID)
	return nil
}
