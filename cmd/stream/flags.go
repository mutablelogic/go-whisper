package main

import "flag"

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
