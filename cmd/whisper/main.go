package main

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	// Packages
	kong "github.com/alecthomas/kong"
	tablewriter "github.com/djthorpe/go-tablewriter"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	whisper "github.com/mutablelogic/go-whisper"
)

type Globals struct {
	NoGPU bool   `name:"nogpu" help:"Disable GPU acceleration"`
	Debug bool   `name:"debug" help:"Enable debug output"`
	Dir   string `name:"dir" help:"Path to model store, uses ${WHISPER_DIR} " default:"${WHISPER_DIR}"`

	// Writer, service and context
	writer  *tablewriter.Writer
	service *whisper.Whisper
	ctx     context.Context
}

type CLI struct {
	Globals
	Models   ModelsCmd   `cmd:"models" help:"List models"`
	Download DownloadCmd `cmd:"download" help:"Download a model"`
	Server   ServerCmd   `cmd:"server" help:"Run the whisper server"`
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
			"WHISPER_DIR": dirEnvOrDefault(name),
		},
	)

	// Create a whisper server - set options
	opts := []whisper.Opt{}
	if cli.Globals.Debug {
		opts = append(opts, whisper.OptDebug())
	}
	if cli.Globals.NoGPU {
		opts = append(opts, whisper.OptNoGPU())
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(cli.Globals.Dir, 0755); err != nil {
		cmd.FatalIfErrorf(err)
		return
	}

	// Create a whisper server - create
	service, err := whisper.New(cli.Globals.Dir, opts...)
	if err != nil {
		cmd.FatalIfErrorf(err)
		return
	} else {
		cli.Globals.service = service
	}
	defer service.Close()

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

func dirEnvOrDefault(name string) string {
	if dir := os.Getenv("WHISPER_DIR"); dir != "" {
		return dir
	}
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, name)
	}
	return os.TempDir()
}
