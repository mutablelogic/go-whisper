package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	// Packages
	context "github.com/mutablelogic/go-server/pkg/context"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	api "github.com/mutablelogic/go-whisper/pkg/api"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
	sys "github.com/mutablelogic/go-whisper/sys/whisper"
)

func main() {
	// Parse the command line flags
	name := filepath.Base(os.Args[0])
	flags, err := NewFlags(name, os.Args[1:])
	if err != nil {
		if err != flag.ErrHelp {
			log.Println(err)
		}
		os.Exit(-1)
	}

	// Set logging
	sys.Whisper_log_set(func(level sys.LogLevel, text string) {
		if flags.Debug() && (level == sys.LogLevelDebug || level == sys.LogLevelInfo || level == sys.LogLevelWarn) {
			return
		}
		log.Println(level, strings.TrimSpace(text))
	})

	// Determine the directory for models
	dir := flags.Dir()
	if dir == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Println(err)
			os.Exit(-1)
		}
		dir = filepath.Join(cacheDir, name)
	}

	// Create the directory for models
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	// Create a whisper service
	log.Println("Storing models at", dir)
	whisper, err := whisper.New(dir)
	if err != nil {
		log.Println(err)
		os.Exit(-2)
	}

	// Display models
	var models []string
	for _, model := range whisper.ListModels() {
		models = append(models, strconv.Quote(model.Id))
	}
	if len(models) > 0 {
		log.Println("Models:", strings.Join(models, ", "))
	} else {
		log.Println("No models")
	}

	// Create a mux for serving requests, then register the endpoints with the mux
	mux := http.NewServeMux()
	api.RegisterEndpoints(flags.Endpoint(), mux, whisper)

	// Create a new HTTP server
	log.Println("List address", flags.Listen())
	server, err := httpserver.Config{
		Listen: flags.Listen(),
		Router: mux,
	}.New()
	if err != nil {
		log.Println(err)
		os.Exit(-2)
	}

	// Run the server until CTRL+C
	log.Println("Press CTRL+C to exit")
	ctx := context.ContextForSignal(os.Interrupt, syscall.SIGQUIT)
	if err := server.Run(ctx); err != nil {
		log.Println(err)
		os.Exit(-3)
	}

	// Release whisper resources
	log.Println("Terminating")
	if err := whisper.Close(); err != nil {
		log.Println(err)
		os.Exit(-4)
	}
}
