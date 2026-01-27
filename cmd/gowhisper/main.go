package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	// Packages
	kong "github.com/alecthomas/kong"
	client "github.com/mutablelogic/go-client"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	server "github.com/mutablelogic/go-server"
	logger "github.com/mutablelogic/go-server/pkg/logger"
	httpclient "github.com/mutablelogic/go-whisper/pkg/httpclient"
	trace "go.opentelemetry.io/otel/trace"
	terminal "golang.org/x/term"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Globals struct {
	// Debug option
	Debug   bool             `name:"debug" help:"Enable debug logging"`
	Version kong.VersionFlag `name:"version" help:"Print version and exit"`

	// HTTP server options
	HTTP struct {
		Prefix  string        `name:"prefix" help:"HTTP path prefix" default:"/api"`
		Addr    string        `name:"addr" env:"GOWHISPER_ADDR" help:"HTTP Listen address" default:"localhost:8081"`
		Timeout time.Duration `name:"timeout" help:"HTTP server read/write timeout" default:"10m"`
	} `embed:"" prefix:"http."`

	// Open Telemetry options
	OTel struct {
		Endpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" help:"Open Telemetry endpoint" default:""`
		Header   string `env:"OTEL_EXPORTER_OTLP_HEADERS" help:"OpenTelemetry collector headers"`
		Name     string `env:"OTEL_SERVICE_NAME" help:"OpenTelemetry service name" default:"${EXECUTABLE_NAME}"`
	} `embed:"" prefix:"otel."`

	// Private fields
	ctx    context.Context
	cancel context.CancelFunc
	logger server.Logger
	tracer trace.Tracer
}

type CLI struct {
	Globals
	ModelCommands
	TranscribeCommands
	TranslateCommands
	ServerCommands
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	// Get executable name
	execName := "gowhisper"
	if exe, err := os.Executable(); err == nil {
		execName = filepath.Base(exe)
	}

	// Parse command-line arguments
	cli := new(CLI)
	ctx := kong.Parse(cli,
		kong.Name(execName),
		kong.Description("gowhisper command line interface"),
		kong.Vars{
			"version":         VersionJSON(),
			"EXECUTABLE_NAME": execName,
		},
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	// Run the command
	os.Exit(run(ctx, &cli.Globals))
}

func run(ctx *kong.Context, globals *Globals) int {
	parent := context.Background()

	// Create Logger
	if isTerminal(os.Stderr) {
		globals.logger = logger.New(os.Stderr, logger.Term, globals.Debug)
	} else {
		globals.logger = logger.New(os.Stderr, logger.JSON, globals.Debug)
	}

	// Create the context and cancel function
	globals.ctx, globals.cancel = signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer globals.cancel()

	// Open Telemetry
	if globals.OTel.Endpoint != "" {
		provider, err := otel.NewProvider(globals.OTel.Endpoint, globals.OTel.Header, globals.OTel.Name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return -2
		}
		defer provider.Shutdown(context.Background())

		// Store tracer for creating spans
		globals.tracer = provider.Tracer(globals.OTel.Name)
	}

	// Call the Run() method of the selected parsed command.
	if err := ctx.Run(globals); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return -1
	}

	return 0
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (g *Globals) Client() (*httpclient.Client, error) {
	endpoint, opts, err := g.clientEndpoint("whisper")
	if err != nil {
		return nil, err
	}
	return httpclient.New(endpoint, opts...)
}

// clientEndpoint returns the endpoint URL and client options for the given path suffix.
func (g *Globals) clientEndpoint(suffix string) (string, []client.ClientOpt, error) {
	scheme := "http"
	host, port, err := net.SplitHostPort(g.HTTP.Addr)
	if err != nil {
		return "", nil, err
	}

	// Default host to localhost if empty (e.g., ":8080")
	if host == "" {
		host = "localhost"
	}

	// Parse port
	portn, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return "", nil, err
	}
	if portn == 443 {
		scheme = "https"
	}

	// Client options
	opts := []client.ClientOpt{}
	if g.Debug {
		opts = append(opts, client.OptTrace(os.Stderr, true))
	}
	if g.tracer != nil {
		opts = append(opts, client.OptTracer(g.tracer))
	}
	if g.HTTP.Timeout > 0 {
		opts = append(opts, client.OptTimeout(g.HTTP.Timeout))
	}

	// Set prefix
	prefix, err := url.JoinPath("/", g.HTTP.Prefix, suffix)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("%s://%s:%v%s", scheme, host, portn, prefix), opts, nil
}

func isTerminal(w io.Writer) bool {
	if fd, ok := w.(*os.File); ok {
		return terminal.IsTerminal(int(fd.Fd()))
	}
	return false
}
