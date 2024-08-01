package main

import (
	"flag"
	"os"
)

type Flags struct {
	*flag.FlagSet

	// Flag parameters
	endpoint *string
	listen   *string
	dir      *string
	debug    *bool
}

func NewFlags(name string, args []string) (*Flags, error) {
	flags := &Flags{
		FlagSet: flag.NewFlagSet(name, flag.ContinueOnError),
	}
	flags.endpoint = flags.String("endpoint", "/v1", "HTTP endpoint")
	flags.listen = flags.String("listen", "127.0.0.1:8080", "HTTP Listen address")
	flags.dir = flags.String("dir", "${WHISPER_DATA}", "Model data directory")
	flags.debug = flags.Bool("debug", false, "Output additional debug information")

	// Parse flags and return any error
	return flags, flags.Parse(args)
}

func (f *Flags) Listen() string {
	return *f.listen
}

func (f *Flags) Dir() string {
	return os.ExpandEnv(*f.dir)
}

func (f *Flags) Endpoint() string {
	return *f.endpoint
}

func (f *Flags) Debug() bool {
	return *f.debug
}
