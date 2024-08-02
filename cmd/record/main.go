package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

	// Packages
	context "github.com/mutablelogic/go-server/pkg/context"
	whisper "github.com/mutablelogic/go-whisper"
	vad "github.com/mutablelogic/go-whisper/pkg/vad"
	sdl "github.com/veandco/go-sdl2/sdl"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
extern void audio_callback(void *userdata, void *stream, int len);
*/
import "C"

var (
	flagListDevices = flag.Bool("devices", false, "List available audio devices")
	flagDevice      = flag.Int("device", -1, "Audio device to use")
)

func deviceNameForIndex(i int) string {
	numDevices := sdl.GetNumAudioDevices(true)
	if i < 0 || i >= numDevices {
		return ""
	}
	return sdl.GetAudioDeviceName(i, true)
}

var (
	v = vad.New(20, 0.01, 2*time.Second)
	s bool
)

// callback function to capture audio data
//
//export audio_callback
func audio_callback(_ unsafe.Pointer, stream unsafe.Pointer, length C.int) {
	frame := cFloat32Slice(stream, length)

	if state := v.Decode(frame); state != s {
		if state {
			fmt.Println("Recording frame")
		} else {
			fmt.Println("Silence")
		}
		s = state
	}
}

func cFloat32Slice(p unsafe.Pointer, sz C.int) []float32 {
	if p == nil {
		return nil
	}
	length := int(sz) / 4 // size of float32 is 4 bytes
	return (*[1 << 30]float32)(p)[:length:length]
}

func main() {
	flag.Parse()
	if sdl.Init(sdl.INIT_AUDIO) != nil {
		fmt.Fprintln(os.Stderr, "Failed to initialize SDL:", sdl.GetError())
		os.Exit(-1)
	}
	defer sdl.Quit()

	if *flagListDevices {
		numDevices := sdl.GetNumAudioDevices(true)
		for i := 0; i < numDevices; i++ {
			deviceName := sdl.GetAudioDeviceName(i, true)
			fmt.Fprintf(os.Stderr, "-device %d: %s\n", i, deviceName)
		}
		os.Exit(0)
		return
	}

	frameSize := 50 * time.Millisecond
	want := &sdl.AudioSpec{
		Freq:     whisper.SampleRate,
		Format:   sdl.AUDIO_F32,
		Channels: 1,
		Samples:  uint16(float64(whisper.SampleRate) * frameSize.Seconds()),
	}
	want.Callback = sdl.AudioCallback(C.audio_callback)

	// Open up the audio device
	var have sdl.AudioSpec
	dev, err := sdl.OpenAudioDevice(deviceNameForIndex(*flagDevice), true, want, &have, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, sdl.GetError())
		os.Exit(-1)
		return
	}
	defer sdl.CloseAudioDevice(dev)

	// Start capturing audio
	sdl.PauseAudioDevice(dev, false)
	fmt.Println("Capturing audio, press Ctrl+C to stop")
	ctx := context.ContextForSignal(os.Interrupt, syscall.SIGQUIT)

	// Main loop
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		}
	}

	// Stop capturing audio
	sdl.PauseAudioDevice(dev, true)
}
