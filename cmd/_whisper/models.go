package main

import (
	"encoding/json"
	"io"
	"os"

	// Packages
	goclient "github.com/mutablelogic/go-client"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	client "github.com/mutablelogic/go-whisper/pkg/client"
)

type ModelsCmd struct {
	Remote bool `flag:"" help:"List remote (openai, gowhisper, elevenlabs) models"`
}

func (cmd ModelsCmd) Run(ctx *Globals) error {
	if cmd.Remote {
		return run_remote_models(ctx)
	} else {
		return run_local_models(ctx)
	}
}

func run_local_models(app *Globals) error {
	models := app.service.ListModels()
	if len(models) == 0 {
		return httpresponse.ErrNotFound.With("no models found")
	} else {
		return list_models(os.Stdout, models)
	}
}

func run_remote_models(app *Globals) error {
	// Create a client
	opts := []goclient.ClientOpt{}
	if app.Debug {
		opts = append(opts, goclient.OptTrace(os.Stderr, false))
	}
	remote, err := client.New(opts...)
	if err != nil {
		return err
	}

	// List models
	models, err := remote.ListModels(app.ctx)
	if err != nil {
		return err
	} else if len(models) == 0 {
		return httpresponse.ErrNotFound.With("no models found")
	} else {
		return list_models(os.Stdout, models)
	}
}

func list_models(w io.Writer, models any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(models)
}
