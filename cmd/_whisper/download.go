package main

import (
	"fmt"
	"log"
	"os"
	"time"

	goclient "github.com/mutablelogic/go-client"
	client "github.com/mutablelogic/go-whisper/pkg/client"
	// Packages
)

type DownloadCmd struct {
	Path   string `arg:"" help:"Model to download"`
	Remote bool   `flag:"" help:"Download remote (gowhisper) models"`
}

func (cmd DownloadCmd) Run(ctx *Globals) error {
	if cmd.Remote {
		return cmd.run_remote_download(ctx)
	} else {
		return cmd.run_local_download(ctx)
	}
}

func (cmd *DownloadCmd) run_local_download(app *Globals) error {
	t := time.Now()
	model, err := app.service.DownloadModel(app.ctx, cmd.Path, func(curBytes, totalBytes uint64) {
		if time.Since(t) > time.Second {
			pct := float64(curBytes) / float64(totalBytes) * 100
			log.Printf("Downloaded %.0f%%", pct)
			t = time.Now()
		}
	})
	if err != nil {
		return err
	}
	fmt.Println(model)
	return nil
}

func (cmd *DownloadCmd) run_remote_download(app *Globals) error {
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

	model, err := remote.DownloadModel(app.ctx, cmd.Path, func(curBytes, totalBytes uint64) {
		pct := float64(curBytes) / float64(totalBytes) * 100
		log.Printf("Downloaded %.0f%%", pct)
	})
	if err != nil {
		return err
	}

	// Print the model details
	fmt.Println(model)
	return nil
}
