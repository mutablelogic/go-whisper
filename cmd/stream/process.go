package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	multierror "github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Process struct {
	wg     sync.WaitGroup
	cancel context.CancelFunc
	ch     chan []float32
	t      time.Duration
	model  whisper.Model
	buf    []float32
	n      int
}

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	defaultWindow = 2 * time.Second
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProcess(model string, window time.Duration) (*Process, error) {
	process := new(Process)
	process.ch = make(chan []float32)

	// Set default window
	if window == 0 {
		window = defaultWindow
	}

	// Create a buffer for the samples. The idea is to make a circular buffer eventually
	process.n = int(whisper.SampleRate * defaultWindow / time.Second)
	process.buf = make([]float32, 0, process.n)

	// Load the model
	if model, err := whisper.New(model); err != nil {
		return nil, err
	} else {
		process.model = model
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

	// Close channel, release buffer
	close(process.ch)
	process.ch = nil
	process.buf = nil

	// Close model
	if err := process.model.Close(); err != nil {
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
				process.buf = append(process.buf, samples...)
				if len(process.buf) >= process.n {
					samples := make([]float32, len(process.buf))
					copy(samples, process.buf)
					go func() {
						if err := process.do(ctx, process.t, samples); err != nil {
							fmt.Println(err)
						}
					}()
					process.n = len(process.buf)
					process.buf = process.buf[:0]
				}
				process.t += time.Second * time.Duration(len(samples)) / whisper.SampleRate
			}
		}
	}
}

func (process *Process) do(ctx context.Context, offset time.Duration, samples []float32) error {
	context, err := process.model.NewContext()
	if err != nil {
		return err
	}

	// Process samples
	if err := context.Process(samples, nil); err != nil {
		return err
	}
	// Print out the results
	for {
		segment, err := context.NextSegment()
		if err != nil {
			break
		}
		fmt.Printf("[%6s->%6s] %s\n", segment.Start+offset, segment.End+offset, segment.Text)
	}

	// Return success
	return nil
}
