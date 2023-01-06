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
	this := new(Flags)
	this.FlagSet = flag.NewFlagSet(name, flag.ContinueOnError)

	// Register flags
	this.Int("device", 0, "Audio device to use")
	this.String("model", "models/ggml-base.bin", "Whisper model path")
	this.Duration("window", 0, "Window size for processing")

	// Parse flags
	if err := this.Parse(args); err != nil {
		return nil, err
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *Flags) HasDevice() bool {
	var found bool
	this.FlagSet.Visit(func(f *flag.Flag) {
		if f.Name == "device" {
			found = true
		}
	})
	return found
}

func (this *Flags) GetDevice() int {
	return this.Lookup("device").Value.(flag.Getter).Get().(int)
}

func (this *Flags) GetModel() string {
	return this.Lookup("model").Value.String()
}

func (this *Flags) GetWindow() time.Duration {
	return this.Lookup("window").Value.(flag.Getter).Get().(time.Duration)
}
