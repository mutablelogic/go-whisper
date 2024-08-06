package sdl

import (
	"errors"
	"fmt"
	"time"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	sdl "github.com/veandco/go-sdl2/sdl"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Audio struct {
	device sdl.AudioDeviceID
	capture bool
}

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
extern void audio_callback(void *userdata, void *data, int len);
*/
import "C"

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

/*
	SDL Supported audio formats:
	AUDIO_S8 // signed 8-bit samples
	AUDIO_U8// unsigned 8-bit samples
	AUDIO_S16LSB // signed 16-bit samples in little-endian byte order
	AUDIO_S16MSB  // signed 16-bit samples in big-endian byte order
	AUDIO_S16SYS  // signed 16-bit samples in native byte order
	AUDIO_S16     // AUDIO_S16LSB
	AUDIO_U16LSB  // unsigned 16-bit samples in little-endian byte order
	AUDIO_U16MSB // unsigned 16-bit samples in big-endian byte order
	AUDIO_U16SYS  // unsigned 16-bit samples in native byte order
	AUDIO_U16     // AUDIO_U16LSB
	AUDIO_S32LSB // 32-bit integer samples in little-endian byte order
	AUDIO_S32MSB  // 32-bit integer samples in big-endian byte order
	AUDIO_S32SYS // 32-bit integer samples in native byte order
	AUDIO_S32    // AUDIO_S32LSB
	AUDIO_F32LSB  // 32-bit floating point samples in little-endian byte order
	AUDIO_F32MSB  // 32-bit floating point samples in big-endian byte order
	AUDIO_F32SYS  // 32-bit floating point samples in native byte order
	AUDIO_F32     // AUDIO_F32LSB
*/

var (
	mapAudio = map[string]sdl.AudioFormat{
//		"u8":  sdl.AUDIO_U8,
//		"s8":  sdl.AUDIO_S8,
//		"s16": sdl.AUDIO_S16SYS,
		"flt": sdl.AUDIO_F32SYS,
	}
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new audio
func (s *Context) newAudio(par *ffmpeg.Par, buffer time.Duration) (*Audio, error) {
	if !par.Type().Is(AUDIO) {
		return nil, errors.New("invalid audio parameters")
	}

	src_format := fmt.Sprint(par.SampleFormat())
	format, exists := mapAudio[src_format]
	if !exists {
		return nil, ErrBadParameter.Withf("unsupported sample format %q", src_format)
	}

	var capture bool
	switch {
	case par.Type().Is(INPUT):
		capture = true
	case par.Type().Is(OUTPUT):
		capture = false
	default:
		return nil, errors.New("invalid audio parameters, require INPUT or OUTPUT")
	}

	// Set parameters for the audio device
	var desired, obtained sdl.AudioSpec
	desired.Freq = int32(par.Samplerate())
	desired.Format = format
	desired.Channels = uint8(par.ChannelLayout().NumChannels())
	desired.Samples = uint16(float64(par.Samplerate()) * buffer.Seconds())
	desired.Callback = sdl.AudioCallback(C.audio_callback)

	// Open the audio device
	device, err := sdl.OpenAudioDevice("", capture, &desired, &obtained, 0)
	if err != nil {
		return nil, err
	} else if desired.Freq != obtained.Freq {
		return nil, ErrBadParameter.Withf("sample rate %d not supported", desired.Freq)
	}
	if desired.Format != obtained.Format {
		return nil, ErrBadParameter.Withf("unsupported sample format %q", src_format)
	}
	if desired.Channels != obtained.Channels {
		return nil, ErrBadParameter.Withf("number of channels %d not supported", desired.Channels)
	}

	// Return the audio device
	return &Audio{
		device,
		capture,
	}, nil
}

func (a *Audio) Close() error {
	var result error

	// Close the audio device
	sdl.CloseAudioDevice(a.device)

	// Return any errors
	return result
}

// callback function to capture audio data
//
//export audio_callback
func audio_callback(_ unsafe.Pointer, data unsafe.Pointer, length C.int) {
	frame := cFloat32Slice(data, length << 2)
	if state := v.Decode(frame); state != s {
		if state {
			fmt.Println("Recording frame")
		} else {
			fmt.Println("Silence")
		}
		s = state
	}
}

// Utility methods
func cFloat32Slice(p unsafe.Pointer, sz C.int) []float32 {
	if p == nil {
		return nil
	}
	length := int(sz)
	return (*[1 << 30]float32)(p)[:length:length]
}
