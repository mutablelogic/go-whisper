package main

import (
	sdl "github.com/veandco/go-sdl2/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Streamer struct{}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewStreamer() (*Streamer, error) {
	if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
		return nil, err
	}
	sdl.LogSetPriority(sdl.LOG_CATEGORY_APPLICATION, sdl.LOG_PRIORITY_INFO)
	sdl.SetHintWithPriority(sdl.HINT_AUDIO_RESAMPLING_MODE, "medium", sdl.HINT_OVERRIDE)

	return &Streamer{}, nil
}

func (streamer *Streamer) Close() error {
	sdl.Quit()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AudioDevices returns a map of audio devices
func (Streamer) AudioDevices() map[int]string {
	num := sdl.GetNumAudioDevices(true)
	result := make(map[int]string, num)
	for i := 0; i < num; i++ {
		result[i] = sdl.GetAudioDeviceName(i, true)
	}
	return result
}

// Open returns a capture context
func (Streamer) Open(dev int, rate int32, ch uint8, samples uint16) (*StreamerContext, error) {
	var requested, obtained sdl.AudioSpec
	requested.Freq = rate
	requested.Format = sdl.AUDIO_F32
	requested.Channels = ch
	requested.Samples = samples
	if id, err := sdl.OpenAudioDevice(sdl.GetAudioDeviceName(dev, true), true, &requested, &obtained, 0); err != nil {
		return nil, err
	} else {
		return NewStreamerContext(id, obtained)
	}
}
