package main

import (
	"flag"
	"os"
)

type Flags struct {
	*flag.FlagSet

	// Flag parameters
	endpoint *string
	debug    *bool
}

func NewFlags(name string, args []string) (*Flags, error) {
	flags := &Flags{
		FlagSet: flag.NewFlagSet(name, flag.ContinueOnError),
	}
	flags.endpoint = flags.String("endpoint", "${WHISPER_URL}", "HTTP endpoint")
	flags.debug = flags.Bool("debug", false, "Display debug information")

	// Parse flags and return any error
	return flags, flags.Parse(args)
}

func (f *Flags) Endpoint() string {
	return os.ExpandEnv(*f.endpoint)
}

func (f *Flags) Debug() bool {
	return *f.debug
}
