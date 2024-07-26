package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mutablelogic/go-whisper/pkg/api"
	"github.com/mutablelogic/go-whisper/pkg/whisper"
)

func main() {
	// Parse the command line flags
	name := filepath.Base(os.Args[0])
	flags, err := NewFlags(name, os.Args[1:])
	if err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(-1)
	}

	// Create a whisper service
	whisper, err := whisper.New(flags.Dir())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

	// Register the endpoints
	api.RegisterEndpoints(flags.Endpoint(), http.DefaultServeMux, whisper)

	// Start the server
	fmt.Println("Listening on", flags.Listen())
	if err := http.ListenAndServe(flags.Listen(), nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
}
