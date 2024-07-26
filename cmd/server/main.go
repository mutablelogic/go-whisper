package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Packages
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
		if flags.Debug() && level == sys.LogLevelDebug || level == sys.LogLevelInfo || level == sys.LogLevelWarn {
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

	// Register the endpoints
	api.RegisterEndpoints(flags.Endpoint(), http.DefaultServeMux, whisper)

	// Start the server
	log.Println("Listening on", flags.Listen())
	if err := http.ListenAndServe(flags.Listen(), nil); err != nil {
		log.Println(err)
		os.Exit(-2)
	}
}
