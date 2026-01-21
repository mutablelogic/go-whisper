//go:build !client

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	"github.com/mutablelogic/go-media/pkg/segmenter"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	pkg "github.com/mutablelogic/go-whisper/pkg"
	httphandler "github.com/mutablelogic/go-whisper/pkg/httphandler"
	version "github.com/mutablelogic/go-whisper/pkg/version"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServerCommands struct {
	RunServer RunServer `cmd:"" name:"run" help:"Run server." group:"SERVER"`
}

type RunServer struct {
	Models string `name:"models" env:"GOWHISPER_DIR" help:"Models directory path" default:""`

	// API keys for external services
	OpenAIKey     string `name:"openai-api-key" env:"OPENAI_API_KEY" help:"OpenAI API key"`
	ElevenLabsKey string `name:"elevenlabs-api-key" env:"ELEVENLABS_API_KEY" help:"ElevenLabs API key"`

	// TLS server options
	TLS struct {
		ServerName string `name:"name" help:"TLS server name"`
		CertFile   string `name:"cert" help:"TLS certificate file"`
		KeyFile    string `name:"key" help:"TLS key file"`
	} `embed:"" prefix:"tls."`

	// Whisper options
	Whisper struct {
		MaxContexts uint `name:"max-contexts" help:"Maximum number of concurrent contexts" default:"0"`
		GPU         bool `name:"gpu" help:"Use GPU if available" default:"true"`
	} `embed:"" prefix:"whisper."`

	// Segmenter options
	Segmenter struct {
		MinSilenceSize time.Duration `name:"min-silence-size" help:"Minimum silence segment size"`
		MaxSegmentSize time.Duration `name:"max-segment-size" help:"Maximum segment size"`
	} `embed:"" prefix:"segmenter."`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *RunServer) Run(ctx *Globals) error {
	// Set models path - use default if not specified
	modelsPath := cmd.Models
	if modelsPath == "" {
		// Use default from environment or cache dir
		if dir, err := os.UserCacheDir(); err == nil {
			modelsPath = filepath.Join(dir, "gowhisper")
		} else {
			modelsPath = filepath.Join(os.TempDir(), "gowhisper")
		}
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(modelsPath, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	// Report models path
	ctx.logger.With("models", modelsPath).Print(ctx.ctx, "using models directory")

	// Build options
	managerOpts := []pkg.Opt{}
	if cmd.Whisper.MaxContexts > 0 {
		managerOpts = append(managerOpts, pkg.WithWhisperOpt(whisper.OptMaxConcurrent(int(cmd.Whisper.MaxContexts))))
	}
	if !cmd.Whisper.GPU {
		managerOpts = append(managerOpts, pkg.WithWhisperOpt(whisper.OptNoGPU()))
	}
	if ctx.Debug {
		managerOpts = append(managerOpts, pkg.WithWhisperOpt(whisper.OptDebug()))

		// Provide a log function so debug output is actually shown
		managerOpts = append(managerOpts, pkg.WithWhisperOpt(whisper.OptLog(func(s string) {
			ctx.logger.Print(ctx.ctx, s)
		})))
	}
	if ctx.tracer != nil {
		managerOpts = append(managerOpts, pkg.WithTracer(ctx.tracer))
	}
	if ctx.Debug {
		// Enable HTTP tracing for OpenAI and ElevenLabs clients
		managerOpts = append(managerOpts, pkg.WithClientOpts(client.OptTrace(os.Stderr, false)))
	}
	if ctx.HTTP.Timeout > 0 {
		// Set HTTP client timeout for OpenAI and ElevenLabs clients
		managerOpts = append(managerOpts, pkg.WithClientOpts(client.OptTimeout(ctx.HTTP.Timeout)))
	}
	if cmd.OpenAIKey != "" {
		managerOpts = append(managerOpts, pkg.OptOpenAIKey(cmd.OpenAIKey))
	}
	if cmd.ElevenLabsKey != "" {
		managerOpts = append(managerOpts, pkg.OptElevenLabsKey(cmd.ElevenLabsKey))
	}
	if cmd.Segmenter.MinSilenceSize > 0 {
		managerOpts = append(managerOpts, pkg.WithSegmenterOpt(segmenter.WithSilenceSize(cmd.Segmenter.MinSilenceSize)))
	}
	if cmd.Segmenter.MaxSegmentSize > 0 {
		managerOpts = append(managerOpts, pkg.WithSegmenterOpt(segmenter.WithSegmentSize(cmd.Segmenter.MaxSegmentSize)))
	}

	// Create the whisper manager
	manager, err := pkg.New(modelsPath, managerOpts...)
	if err != nil {
		return err
	}
	defer func() {
		if err := manager.Close(); err != nil {
			ctx.logger.Print(ctx.ctx, "error closing manager: ", err)
		}
	}()

	// Set logging middleware
	middleware := httphandler.HTTPMiddlewareFuncs{
		ctx.logger.HandleFunc,
	}

	// If we have an OTEL tracer, add tracing middleware
	if ctx.tracer != nil {
		middleware = append(middleware, otel.HTTPHandlerFunc(ctx.tracer))
	}

	// Register HTTP handlers
	router := http.NewServeMux()

	// Register whisper handlers
	whisper_prefix, err := url.JoinPath(ctx.HTTP.Prefix, "whisper")
	if err != nil {
		return err
	}
	httphandler.RegisterHandlers(router, whisper_prefix, manager, middleware)

	// Create a TLS config
	var tlsconfig *tls.Config
	if cmd.TLS.CertFile != "" || cmd.TLS.KeyFile != "" {
		tlsconfig, err = httpserver.TLSConfig(cmd.TLS.ServerName, true, cmd.TLS.CertFile, cmd.TLS.KeyFile)
		if err != nil {
			return err
		}
	}

	// Create a HTTP server with timeouts
	httpopts := []httpserver.Opt{}
	if ctx.HTTP.Timeout > 0 {
		httpopts = append(httpopts, httpserver.WithReadTimeout(ctx.HTTP.Timeout))
		httpopts = append(httpopts, httpserver.WithWriteTimeout(ctx.HTTP.Timeout))
	}
	server, err := httpserver.New(ctx.HTTP.Addr, router, tlsconfig, httpopts...)
	if err != nil {
		return err
	}

	// We run the server
	var wg sync.WaitGroup
	var result error

	// Output the version
	if version.GitTag != "" {
		ctx.logger.Printf(ctx.ctx, "gowhisper@%s", version.GitTag)
	} else if version.GitHash != "" {
		ctx.logger.Printf(ctx.ctx, "gowhisper@%s", version.GitHash[:8])
	} else {
		ctx.logger.Printf(ctx.ctx, "gowhisper")
	}

	// Run the HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Output listening information
		ctx.logger.With("addr", ctx.HTTP.Addr, "prefix", ctx.HTTP.Prefix).Print(ctx.ctx, "http server starting")

		// Run the server
		if err := server.Run(ctx.ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				result = errors.Join(result, fmt.Errorf("http server error: %w", err))
			}
			ctx.cancel()
		}
	}()

	// Wait for goroutine to finish
	wg.Wait()

	// Terminated message
	if result == nil {
		ctx.logger.With("addr", ctx.HTTP.Addr, "prefix", ctx.HTTP.Prefix).Print(ctx.ctx, "terminated gracefully")
	} else {
		ctx.logger.With("addr", ctx.HTTP.Addr, "prefix", ctx.HTTP.Prefix).Print(ctx.ctx, result)
	}

	// Return any error
	return result
}
