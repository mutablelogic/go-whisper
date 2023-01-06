package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	// Packages
	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

func main() {
	// Flags
	flags, err := NewFlags(filepath.Base(os.Args[0]), os.Args[1:])
	if err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	// Initialize audio streaming
	streamer, err := NewStreamer()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer streamer.Close()

	// Print out capture devices
	if !flags.HasDevice() {
		PrintDevices(streamer)
		os.Exit(0)
	}

	// Check for file argument
	if flags.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Missing file argument")
		os.Exit(1)
	}

	// Create sample processor
	process, err := NewProcess(flags.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer process.Close()

	// Open capture device
	context, err := streamer.Open(flags.GetDevice(), whisper.SampleRate, 1, 2048)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer context.Close()

	// Repeat until cancelled
	ctx := ContextWithCancel([]os.Signal{os.Interrupt})
	fmt.Println("[speak now]")
	context.Start()

	t := process.T()
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		default:
			if samples, err := context.Samples(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			} else if len(samples) > 0 {
				process.C() <- samples
			}
			if t2 := process.T(); t2.Truncate(time.Second) != t.Truncate(time.Second) {
				fmt.Println("T=", t2)
				t = t2
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	context.Stop()
	fmt.Println("[stop speaking]")
}

func PrintDevices(streamer *Streamer) {
	devices := streamer.AudioDevices()
	if len(devices) == 0 {
		fmt.Fprintln(os.Stderr, "No audio devices found")
		os.Exit(1)
	}
	// List devices
	fmt.Println("Audio Devices:")
	for i, name := range devices {
		fmt.Printf("  %d: %s\n", i, name)
	}
	fmt.Println("\nUse flag -device <n> to select device to capture from")
}

func ContextWithCancel(sigs []os.Signal) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, sigs...)
		<-c
		cancel()
	}()
	return ctx
}
