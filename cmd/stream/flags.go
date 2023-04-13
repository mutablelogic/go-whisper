package main

import (
	"flag"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Flags struct {
	*flag.FlagSet
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFlags(name string, args []string) (*Flags, error) {
	// Create flags
	tmpFlags := new(Flags)
	tmpFlags.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)

	// Register flags
	tmpFlags.Int("device", 0, "Audio device to use")
	tmpFlags.String("model", "models/ggml-base.bin", "Whisper model path")
	tmpFlags.Duration("window", 0, "Window size for processing")

	// Parse flags
	if err := tmpFlags.Parse(args); err != nil {
		return nil, err
	}

	// Return success
	return tmpFlags, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *Flags) HasDevice() bool {
	var found bool
	t.FlagSet.Visit(func(f *flag.Flag) {
		if f.Name == "device" {
			found = true
		}
	})
	return found
}

func (t *Flags) GetDevice() int {
	return t.Lookup("device").Value.(flag.Getter).Get().(int)
}

func (t *Flags) GetModel() string {
	return t.Lookup("model").Value.String()
}

func (t *Flags) GetWindow() time.Duration {
	return t.Lookup("window").Value.(flag.Getter).Get().(time.Duration)
}
