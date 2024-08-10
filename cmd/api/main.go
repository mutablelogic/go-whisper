package main

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	// Packages
	kong "github.com/alecthomas/kong"
	tablewriter "github.com/djthorpe/go-tablewriter"
	opt "github.com/mutablelogic/go-client"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	client "github.com/mutablelogic/go-whisper/pkg/client"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Globals struct {
	Url   string `name:"url" help:"URL of whisper service (can be set from WHISPER_URL env)" default:"${WHISPER_URL}"`
	Debug bool   `name:"debug" help:"Enable debug output"`

	// Writer, service and context
	writer *tablewriter.Writer
	client *client.Client
	ctx    context.Context
}

type CLI struct {
	Globals

	Ping     PingCmd     `cmd:"ping" help:"Ping the whisper service"`
	Models   ModelsCmd   `cmd:"models" help:"List models"`
	Download DownloadCmd `cmd:"download" help:"Download a model"`
	Delete   DeleteCmd   `cmd:"delete" help:"Delete a model"`
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultEndpoint = "http://localhost:8080/api/v1"
)

////////////////////////////////////////////////////////////////////////////////
// MAIN

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
		kong.Description("speech transcription and translation service client"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{
			"WHISPER_URL": envOrDefault("WHISPER_URL", defaultEndpoint),
		},
	)

	// Set whisper client options
	opts := []opt.ClientOpt{}
	if cli.Globals.Debug {
		opts = append(opts, opt.OptTrace(os.Stderr, true))
	}

	// Create a whisper client
	client, err := client.New(cli.Globals.Url, opts...)
	if err != nil {
		cmd.FatalIfErrorf(err)
		return
	} else {
		cli.Globals.client = client
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

func envOrDefault(name, def string) string {
	if value := os.Getenv(name); value != "" {
		return value
	} else {
		return def
	}
}
