package main

import (
	"context"
	"encoding/binary"
	"os"
	"sync"
	"time"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Process struct {
	wg     sync.WaitGroup
	cancel context.CancelFunc
	w      *os.File
	ch     chan []float32
	t      time.Duration
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProcess(path string) (*Process, error) {
	process := new(Process)
	process.ch = make(chan []float32)

	// Create file
	fh, err := os.Create(path)
	if err != nil {
		return nil, err
	} else {
		process.w = fh
	}

	// Create background goroutine
	ctx, cancel := context.WithCancel(context.Background())
	process.wg.Add(1)
	process.cancel = cancel
	go func() {
		defer process.wg.Done()
		process.process(ctx)
	}()

	// Return success
	return process, nil
}

func (process *Process) Close() error {
	var result error

	// Cancel context and wait for end of context
	process.cancel()
	process.wg.Wait()

	// Close channel
	close(process.ch)
	process.ch = nil

	// Close output file
	if err := process.w.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	// Return success or errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return channel to send samples to
func (process *Process) C() chan<- []float32 {
	return process.ch
}

// Return the timestamp of the queued samples
func (process *Process) T() time.Duration {
	return process.t
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// process incoming samples
func (process *Process) process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case samples := <-process.ch:
			if len(samples) > 0 {
				if err := binary.Write(process.w, binary.LittleEndian, samples); err != nil {
					return err
				} else {
					process.t += time.Second * time.Duration(len(samples)) / whisper.SampleRate
				}
			}
		}
	}
}
