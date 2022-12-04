package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	// Packages
	"github.com/djthorpe/go-whisper/pkg/whisper"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	RegisterFlags(flagset)

	// Check for help
	if err := flagset.Parse(os.Args[1:]); errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Init audio
	if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init SDL: %s\n", err)
		os.Exit(1)
	}
	defer sdl.Quit()

	sdl.LogSetPriority(sdl.LOG_CATEGORY_APPLICATION, sdl.LOG_PRIORITY_INFO)
	sdl.SetHintWithPriority(sdl.HINT_AUDIO_RESAMPLING_MODE, "medium", sdl.HINT_OVERRIDE)

	// Check for device, if -1 then list devices and exit
	device_num := FlagDevice(flagset)
	if device_num < 0 {
		num := sdl.GetNumAudioDevices(true)
		fmt.Println("Use -device flag to use a specific capture device:")
		for i := 0; i < num; i++ {
			fmt.Printf("  -device %d: '%s'\n", i, sdl.GetAudioDeviceName(i, true))
		}
		os.Exit(0)
	}

	// Open model
	if flagset.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Missing model argument, use --help for usage")
		os.Exit(1)
	}
	model, err := whisper.New(flagset.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer model.Close()

	// Open device
	device, err := OpenCaptureDevice(device_num)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open capture device: %s\n", err)
		os.Exit(1)
	}
	defer sdl.CloseAudioDevice(device)

	// Set buffer, parameters
	buf := NewBuffer(FlagLength(flagset))
	model.SetProcessors(FlagProc(flagset))
	model.SetTranslate(false)
	model.SetSingleSegment(true)
	model.SetPrintProgress(false)
	model.SetPrintRealtime(false)
	model.SetPrintTimestamps(false)
	model.SetPrintSpecial(false)

	// Repeat until cancelled
	ctx := ContextWithCancel([]os.Signal{os.Interrupt})
	fmt.Println("[speak now]")
	sdl.PauseAudioDevice(device, false)

FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		default:
			// Read and process audio
			if err := sdl.DequeueAudio(device, buf.Bytes()); err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			if err := model.Process(buf.Samples(), func(num int, t0, t1 time.Duration, tokens []whisper.Token) {
				fmt.Printf("n=%02d t0=%v t1=%v ", num, t0, t1)
				for _, token := range tokens {
					fmt.Printf("%v ", token.Text())
				}
				fmt.Println("")
			}); err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			/*
				if size := sdl.GetQueuedAudioSize(device); size > 2*n_samples*whisper.SAMPLE_SIZE {
					fmt.Println("WARNING: cannot process audio fast enough, dropping audio ...")
					sdl.ClearQueuedAudio(sdl.AudioDeviceID(device))
				} else {
					fmt.Println("queued audio size=", size)
				}
				// Wait for enough audio
				for {
					if size := sdl.GetQueuedAudioSize(device); size >= n_samples*whisper.SAMPLE_SIZE {
						break
					} else {
						fmt.Println("waiting for audio, size=", size)
					}
					sdl.Delay(1)
				}
				// Read audio
				n_samples_new := sdl.GetQueuedAudioSize(device) / whisper.SAMPLE_SIZE
				fmt.Println("samples=", n_samples_new)
				//sdl.DequeueAudio(device,data)
				//SDL_DequeueAudio(g_dev_id_in, pcmf32.data()+n_samples_take, n_samples_new*sizeof(float))
			*/
		}
	}

	// Set parameters from step and length
	//	const int n_samples_len = (params.length_ms / 1000.0) * WHISPER_SAMPLE_RATE
	//	const int n_samples_30s = 30 * WHISPER_SAMPLE_RATE
	//	const int n_samples_keep = 0.2 * WHISPER_SAMPLE_RATE
}

func OpenCaptureDevice(num int) (sdl.AudioDeviceID, error) {
	var requested, obtained sdl.AudioSpec
	requested.Freq = whisper.SAMPLE_RATE
	requested.Format = sdl.AUDIO_F32
	requested.Channels = 1
	requested.Samples = 1024
	return sdl.OpenAudioDevice(sdl.GetAudioDeviceName(num, true), true, &requested, &obtained, 0)
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
