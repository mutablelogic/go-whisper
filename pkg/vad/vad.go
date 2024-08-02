package vad

import (
	"math"
	"time"
	// Packages
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type VAD struct {
	Gain      float64
	Threshold float64
	Timeout   time.Duration

	// When we last started recording
	t time.Time
	r bool
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Typical values are gain=20, threshold=0.003, timeout=2s
func New(gain, threshold float64, timeout time.Duration) *VAD {
	return &VAD{Gain: gain, Threshold: threshold, Timeout: timeout}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Compute the energy of a frame of audio data
// and return true if the frame of data should be recorded
func (v *VAD) Decode(data []float32) bool {
	// Compute the gain
	gain := float32(math.Pow(10, v.Gain/20.0))

	// Compute frame energy
	energy := float64(0)
	for i := 0; i < len(data); i++ {
		data[i] *= gain
		energy += float64(data[i]) * float64(data[i])
	}
	energy /= math.Sqrt(float64(len(data)))
	if energy > v.Threshold {
		if v.t.IsZero() {
			v.r = true
		}
		v.t = time.Now()
	} else if !v.t.IsZero() {
		if time.Since(v.t) > v.Timeout {
			v.t = time.Time{}
			v.r = false
		}
	}

	// Return the computed result
	return v.r
}
