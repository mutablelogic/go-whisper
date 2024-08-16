package main

import (
	"context"
	"log"
	"os"
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

	// Decode each frame
	if err := denoiser.Decode(context.Background(), func(ts time.Duration, p float32, data []float32) error {
		log.Printf("[SEGMENT] ts=%-5v p=%.2f nsamp=%d", ts, p, len(data))
		return nil
	}); err != nil {
		log.Fatalln(err)
	}
}
