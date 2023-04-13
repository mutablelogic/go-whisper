package main

import (
	"reflect"
	"unsafe"

	sdl "github.com/veandco/go-sdl2/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type StreamerContext struct {
	Dev  sdl.AudioDeviceID
	Spec sdl.AudioSpec
	u8   []byte
	f32  []float32
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewStreamerContext(dev sdl.AudioDeviceID, spec sdl.AudioSpec) (*StreamerContext, error) {
	streamer := new(StreamerContext)
	streamer.Dev = dev
	streamer.Spec = spec
	streamer.u8 = make([]byte, spec.Size)

	// f32 is a fake slice
	header := (*reflect.SliceHeader)(unsafe.Pointer(&streamer.f32))
	header.Data = uintptr(unsafe.Pointer(&streamer.u8[0]))
	header.Len = int(spec.Size >> 2)
	header.Cap = int(spec.Size >> 2)

	// Return success
	return streamer, nil
}

func (streamer *StreamerContext) Close() error {
	sdl.CloseAudioDevice(streamer.Dev)
	streamer.Dev = 0
	streamer.u8 = nil
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Start capturing audio
func (streamer *StreamerContext) Start() {
	sdl.PauseAudioDevice(streamer.Dev, false)
}

// Stop capturing audio
func (streamer *StreamerContext) Stop() {
	sdl.PauseAudioDevice(streamer.Dev, true)
}

// Return the number of samples in the audio queue
func (streamer *StreamerContext) NumSamples() uint32 {
	// Number of samples is the number of bytes divided by the number of bytes per sample (float32)
	return sdl.GetQueuedAudioSize(streamer.Dev) >> 2
}

// Dequeue audio from the queue, return nil if not enough audio to fill the buffer
func (streamer *StreamerContext) Samples() ([]float32, error) {
	// Check for audio data
	n := sdl.GetQueuedAudioSize(streamer.Dev)
	if n < uint32(len(streamer.u8)) {
		return nil, nil
	}

	// Return the samples
	if _, err := sdl.DequeueAudio(streamer.Dev, streamer.u8); err != nil {
		return nil, err
	} else {
		return streamer.f32[:], nil
	}
}

// Clear any queued audio
func (streamer *StreamerContext) Clear() {
	sdl.ClearQueuedAudio(streamer.Dev)
}
