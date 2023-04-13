package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
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
	t      int64
	model  whisper.Model
	buf    []float32
	n      int
	pool   *sync.Pool
}

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	defaultWindow = 2 * time.Second
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProcess(model string, window time.Duration) (*Process, error) {
	// Create process
	var num = int(whisper.SampleRate * defaultWindow / time.Second)
	process := &Process{
		ch: make(chan []float32),
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]float32, 0, num)
			},
		},
		n:   num,
		buf: make([]float32, 0, num),
	}

	// Set default window
	if window == 0 {
		window = defaultWindow
	}

	// Load the model
	if model, err := whisper.New(model); err != nil {
		return nil, err
	} else {
		process.model = model
	}

	// Create background goroutine
	ctx, cancel := context.WithCancel(context.Background())
	process.cancel = cancel
	process.wg.Add(1)
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
	return time.Duration(atomic.LoadInt64(&process.t)) * time.Second / whisper.SampleRate
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
					samples := process.pool.Get().([]float32)
					samples = append(samples[:0], process.buf...)
					process.pool.Put(process.buf[:0])
					go func() {
						if err := process.do(ctx, atomic.LoadInt64(&process.t), samples); err != nil {
							fmt.Println(err)
						}
					}()
					process.n = len(process.buf)
					process.buf = process.buf[:0]
				}
				atomic.AddInt64(&process.t, int64(len(samples)))
			}
		}
	}
}

func (process *Process) do(ctx context.Context, offset int64, samples []float32) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

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
		select {
		case <-ctx.Done():
			return nil
		default:
			segment, err := context.NextSegment()
			if err != nil {
				return nil
			}
			fmt.Printf("[%6s->%6s] %s\n", segment.Start+time.Duration(offset)*time.Second, segment.End+time.Duration(offset)*time.Second, segment.Text)
		}
	}
}
