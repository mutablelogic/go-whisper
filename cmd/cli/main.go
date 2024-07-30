package main

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	// Packages
	kong "github.com/alecthomas/kong"
	tablewriter "github.com/djthorpe/go-tablewriter"
	client "github.com/mutablelogic/go-client"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	api "github.com/mutablelogic/go-whisper/pkg/whisper/client"
)

type Globals struct {
	Debug    bool   `name:"debug" help:"Enable debug output"`
	Endpoint string `name:"endpoint" help:"HTTP endpoint for whisper service (set WHISPER_URL environment variable to use as default)" default:"${WHISPER_URL}"`

	// Writer, client and context
	writer *tablewriter.Writer
	api    *api.Client
	ctx    context.Context
}

type CLI struct {
	Globals
	Models     ModelsCmd     `cmd:"models" help:"List available models"`
	Delete     DeleteCmd     `cmd:"delete" help:"Delete a model"`
	Download   DownloadCmd   `cmd:"download" help:"Download a model"`
	Transcribe TranscribeCmd `cmd:"transcribe" help:"Transcribe a file"`
	Translate  TranslateCmd  `cmd:"translate" help:"Translate a file"`
}

func main() {
	// The name of the executable
	name, err := os.Executable()
	if err != nil {
		panic(err)
	} else {
		name = filepath.Base(name)
	}

	// Create a cli parser
	cli := CLI{}
	cmd := kong.Parse(&cli,
		kong.Name(name),
		kong.Description("speech transcription and translation service"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{
			"WHISPER_URL": endpointEnvOrDefault(),
		},
	)

	// Create a whisper client
	opts := []client.ClientOpt{}
	if cli.Globals.Debug {
		opts = append(opts, client.OptTrace(os.Stderr, true))
	}
	client, err := api.New(cli.Globals.Endpoint, opts...)
	if err != nil {
		cmd.FatalIfErrorf(err)
	} else {
		cli.Globals.api = client
	}

	// Create a tablewriter object with text output
	writer := tablewriter.New(os.Stdout, tablewriter.OptOutputText())
	cli.Globals.writer = writer

	// Create a context
	cli.Globals.ctx = ctx.ContextForSignal(os.Interrupt, syscall.SIGQUIT)

	// Run the command
	if err := cmd.Run(&cli.Globals); err != nil {
		cmd.FatalIfErrorf(err)
	}
}

func endpointEnvOrDefault() string {
	if endpoint := os.Getenv("WHISPER_URL"); endpoint != "" {
		return endpoint
	}
	return "http://localhost:8080/v1"
}
