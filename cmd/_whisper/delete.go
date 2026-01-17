package main

import (
	"os"
	"time"

	// Packages
	goclient "github.com/mutablelogic/go-client"
	client "github.com/mutablelogic/go-whisper/pkg/client"
)

type DeleteCmd struct {
	Model  string `arg:"" help:"Model to delete"`
	Remote bool   `flag:"" help:"Delete remote (gowhisper) model"`
}

func (cmd *DeleteCmd) Run(ctx *Globals) error {
	if cmd.Remote {
		return cmd.run_remote_delete(ctx)
	} else {
		return cmd.run_local_delete(ctx)
	}
}

func (cmd *DeleteCmd) run_local_delete(app *Globals) error {
	if err := app.service.DeleteModelById(cmd.Model); err != nil {
		return err
	}
	return ModelsCmd{}.Run(app)
}

func (cmd *DeleteCmd) run_remote_delete(app *Globals) error {
	// Create a client for the whisper service
	opts := []goclient.ClientOpt{
		goclient.OptTimeout(5 * time.Minute), // Set a timeout for the request
	}
	if app.Debug {
		opts = append(opts, goclient.OptTrace(os.Stderr, true))
	}
	remote, err := client.New(opts...)
	if err != nil {
		return err
	}

	if err := remote.DeleteModel(app.ctx, cmd.Model); err != nil {
		return err
	}

	return ModelsCmd{Remote: true}.Run(app)
}
