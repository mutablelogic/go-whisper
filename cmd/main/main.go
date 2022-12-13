package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	// Packages
	whisper "github.com/djthorpe/go-whisper/pkg/whisper"
	audio "github.com/go-audio/audio"
	wav "github.com/go-audio/wav"
)

const (
	BufferSize = 64 * 1024
)

func main() {
	flagset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	RegisterFlags(flagset)
	if err := flagset.Parse(os.Args[1:]); errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}
	if flagset.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Missing model and audio file arguments, use --help for usage")
		os.Exit(1)
	}

	// Open the model
	model, err := whisper.New(flagset.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer model.Close()

	// Process each WAV file
	for _, path := range flagset.Args()[1:] {
		fh, err := os.Open(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		defer fh.Close()
		decoder := wav.NewDecoder(fh)
		if decoder == nil || !decoder.IsValidFile() {
			fmt.Fprintln(os.Stderr, "Invalid WAV file:", path)
			continue
		}
		if decoder.NumChans != 1 && decoder.NumChans != 2 {
			fmt.Fprintln(os.Stderr, "WAV file must be mono or stereo:", path)
			continue
		}
		if decoder.SampleRate != whisper.SAMPLE_RATE {
			fmt.Fprintln(os.Stderr, "WAV file must be 16kHz", path)
			continue
		}
		if decoder.BitDepth != 16 {
			fmt.Fprintln(os.Stderr, "WAV file must be 16-bit samples", path)
			continue
		}

		// Set up float32 samples
		duration, err := decoder.Duration()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		samples := make([]float32, 0, int(duration.Seconds()*whisper.SAMPLE_RATE))

		// Decode the WAV file
		buf := &audio.IntBuffer{Data: make([]int, BufferSize), SourceBitDepth: int(decoder.BitDepth), Format: &audio.Format{
			NumChannels: int(decoder.NumChans),
			SampleRate:  int(decoder.SampleRate),
		}}
		for {
			n, err := decoder.PCMBuffer(buf)
			if err != nil || n == 0 {
				break
			} else if n != len(buf.Data) {
				buf.Data = buf.Data[:n]
			}
			if buf.Format.NumChannels == 2 {
				for i := 0; i < len(buf.Data); i += 2 {
					samples = append(samples, float32(buf.Data[i])+float32(buf.Data[i+1])/32768.0)
				}
			} else {
				for i := 0; i < len(buf.Data); i++ {
					samples = append(samples, float32(buf.Data[i])/65536.0)
				}
			}
		}

		// Process the samples
		model.SetProcessors(FlagProc(flagset))
		if err := model.Process(samples); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}
}
