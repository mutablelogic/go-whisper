package main

import (
	"context"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mutablelogic/go-media/pkg/ffmpeg"
	"github.com/mutablelogic/go-whisper/pkg/segmenter"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("Usage: rnnoise <infile> <outfile>")
	}
	r, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	defer r.Close()

	w, err := ffmpeg.Create(os.Args[2], ffmpeg.OptStream(0, ffmpeg.AudioPar("fltp", "mono", 48000)))
	if err != nil {
		log.Fatalln(err)
	}
	defer w.Close()

	// Create a denoiser
	denoiser, err := segmenter.NewDenoiseReader(r)
	if err != nil {
		log.Fatalln(err)
	}
	defer denoiser.Close()

	// Make a channel for frames
	frames := make(chan []float32)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := w.Encode(context.Background(), func(i int) (*ffmpeg.Frame, error) {
			data := <-frames
			if data != nil {
				log.Printf("[FRAME] nsamp=%d %v", len(data), w.Stream(i))
				frame, err := ffmpeg.NewFrame(w.Stream(i).Par())
				if err != nil {
					return nil, err
				}
				return frame, nil
			} else {
				log.Printf("[EOF]")
				return nil, io.EOF
			}
		}, nil); err != nil {
			log.Fatalln(err)
		}
	}()

	// Decode each frame
	if err := denoiser.Decode(context.Background(), func(ts time.Duration, p float32, data []float32) error {
		log.Printf("[SEGMENT] ts=%-5v p=%.2f nsamp=%d", ts, p, len(data))
		frames <- data
		return nil
	}); err != nil {
		log.Fatalln(err)
	}

	// Close channel
	close(frames)
}
