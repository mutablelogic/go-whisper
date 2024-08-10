package segmenter

import (
	"math"
	"time"
	// Packages
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// silence is a silence detector and audio booster for raw samples
// typical values are gain=20, threshold=0.003, timeout=2s
type silence struct {
	Gain      float64       // gain in decibels
	Threshold float64       // threshold for silence
	Timeout   time.Duration // duration of silence before stopping recording

	// When we last started recording
	t time.Time
	r bool
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Increase gain and compute energy of a frame of audio data, return true
// if the frame of data should be recorded, false if it should be ignored
func (s *silence) Process(data []float32) bool {
	energy := process(data, float32(math.Pow(10, s.Gain/20.0)))

	// Compute the gain
	if energy > s.Threshold {
		if s.t.IsZero() {
			// Transition from silence to recording
			s.r = true
		}
		s.t = time.Now()
	} else if !s.t.IsZero() {
		if time.Since(s.t) > s.Timeout {
			// Transition from recording to silence
			s.t = time.Time{}
			s.r = false
		}
	}
	return s.r
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Increase gain and compute energy of a frame of audio data, return the
// energy of the frame of data
func process(data []float32, gain float32) float64 {
	energy := float64(0)
	for i := 0; i < len(data); i++ {
		data[i] *= gain
		energy += float64(data[i]) * float64(data[i])
	}
	return energy / math.Sqrt(float64(len(data)))
}
