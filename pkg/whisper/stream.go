package whisper

import (
	"context"
	"math"
	"time"

	// Packages
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/sys/whisper"
)

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Default sliding-window parameters (milliseconds)
	defaultStepMs   = 3000
	defaultLengthMs = 10000
	defaultKeepMs   = 200
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// StreamSession implements a sliding-window streaming transcription processor
// that mirrors the approach used in whisper.cpp's examples/stream/stream.cpp.
//
// Audio samples are accumulated in audioBuffer. When enough new samples have
// arrived (>= stepSamples), Process() is called to run whisper_full() on a
// window of up to lengthSamples. After processing, the buffer is trimmed to
// keepSamples of overlap so that speech crossing a window boundary is not lost.
// Segment timestamps are adjusted by windowOffset so they are always expressed
// in absolute time, and duplicate segments (whose end time was already emitted
// in a previous window) are suppressed.
type StreamSession struct {
	// task is held for the lifetime of the session. The caller is responsible
	// for returning it to the pool (e.g. via Manager.WithModel) once the
	// session is finished.
	task *Task

	// audioBuffer accumulates raw 16 kHz mono PCM samples.
	audioBuffer []float32

	// stepSamples is the minimum number of new samples that must be present
	// before Process() will trigger a whisper_full() run.
	stepSamples int

	// lengthSamples is the maximum window size fed to whisper_full().
	lengthSamples int

	// keepSamples is the number of samples retained as overlap when the
	// buffer is trimmed after each processing step.
	keepSamples int

	// newSamples counts how many samples have been added since the last
	// Process() call.
	newSamples int

	// lastSegmentEnd is the absolute end timestamp (as time.Duration) of the
	// last segment that was emitted. Used to suppress duplicates when the same
	// speech appears in consecutive overlapping windows.
	lastSegmentEnd time.Duration

	// vadThreshold, if > 0, enables a simple energy gate: Process() will not
	// run unless the RMS energy of the newly-arrived samples exceeds this value.
	vadThreshold float32

	// windowOffset is the absolute time position (as time.Duration) of the
	// first sample currently in audioBuffer. It is advanced whenever samples
	// are trimmed from the front of the buffer.
	windowOffset time.Duration
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewStreamSession creates a StreamSession that wraps the given Task.
// The task must already be initialised (via Task.Init / Manager.WithModel)
// and have had CopyParams() called on it so that its params are ready.
// cfg carries the sliding-window parameters; zero values trigger defaults.
func NewStreamSession(task *Task, cfg schema.StreamConfig) *StreamSession {
	// Apply defaults for zero values
	stepMs := cfg.StepMs
	if stepMs <= 0 {
		stepMs = defaultStepMs
	}
	lengthMs := cfg.LengthMs
	if lengthMs <= 0 {
		lengthMs = defaultLengthMs
	}
	keepMs := cfg.KeepMs
	if keepMs <= 0 {
		keepMs = defaultKeepMs
	}

	// Clamp: keep must be strictly less than length, and step must be <= length.
	if keepMs >= lengthMs {
		keepMs = lengthMs / 2
	}
	if stepMs > lengthMs {
		stepMs = lengthMs
	}

	return &StreamSession{
		task:          task,
		stepSamples:   msToSamples(stepMs),
		lengthSamples: msToSamples(lengthMs),
		keepSamples:   msToSamples(keepMs),
		vadThreshold:  cfg.VadThreshold,
	}
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AddSamples appends raw 16 kHz mono PCM samples to the internal buffer and
// increments the new-sample counter used by ShouldProcess.
func (s *StreamSession) AddSamples(samples []float32) {
	s.audioBuffer = append(s.audioBuffer, samples...)
	s.newSamples += len(samples)
}

// ShouldProcess returns true when enough new audio has accumulated to justify
// running a transcription pass. If vadThreshold > 0, the new samples must also
// exceed that RMS energy level (simple voice-activity gate).
func (s *StreamSession) ShouldProcess() bool {
	if s.newSamples < s.stepSamples {
		return false
	}
	if s.vadThreshold > 0 {
		// TODO: This is a simple RMS energy gate. whisper.cpp has a model-based
		// VAD (Silero) via whisper_vad_context / whisper_vad_detect_speech that
		// would be far more accurate (distinguishes speech from noise, typing,
		// music, etc.), but the Go bindings for the VAD API don't exist yet.
		start := len(s.audioBuffer) - s.newSamples
		if start < 0 {
			start = 0
		}
		if rmsEnergy(s.audioBuffer[start:]) < s.vadThreshold {
			return false
		}
	}
	return true
}

// Process runs whisper_full() on the current audio window and returns any new
// segments whose absolute end time is later than the last emitted segment.
// Call this when ShouldProcess() returns true. The audio buffer is trimmed
// and windowOffset is advanced after each call.
func (s *StreamSession) Process(ctx context.Context) ([]*schema.Segment, error) {
	return s.process(ctx)
}

// ProcessFinal is like Process but runs unconditionally on whatever audio
// remains in the buffer. Use this to flush the last few words when the audio
// stream ends.
func (s *StreamSession) ProcessFinal(ctx context.Context) ([]*schema.Segment, error) {
	if len(s.audioBuffer) == 0 {
		return nil, nil
	}
	return s.process(ctx)
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// process is the shared implementation for Process and ProcessFinal.
func (s *StreamSession) process(ctx context.Context) ([]*schema.Segment, error) {
	// Choose the window: up to the last lengthSamples of the buffer.
	window := s.audioBuffer
	windowStartOffset := s.windowOffset // absolute time of window[0]
	if len(window) > s.lengthSamples {
		trimmed := len(window) - s.lengthSamples
		window = window[trimmed:]
		windowStartOffset = s.windowOffset + samplesToDuration(trimmed)
	}

	if len(window) == 0 {
		s.newSamples = 0
		return nil, nil
	}

	// Run transcription on the window. We pass windowStartOffset as the ts
	// argument so that task.Transcribe adds it to the raw whisper timestamps,
	// giving us absolute times directly in the returned segments.
	var segments []*schema.Segment
	err := s.task.Transcribe(ctx, windowStartOffset, window, func(seg *schema.Segment) {
		// This callback fires for each new segment as whisper_full progresses.
		// We collect them here; deduplication happens below after the full run.
		segments = append(segments, seg)
	})
	if err != nil {
		return nil, err
	}

	// If the callback produced no segments, fall back to whatever appendResult
	// wrote into task.result (the callback path may not fire for very short
	// windows or when context timestamps are disabled).
	if len(segments) == 0 {
		result := s.task.Result()
		if result != nil {
			// Collect only the segments that were added during this Transcribe call.
			// task.appendResult always appends; we want the ones with Start >= windowStartOffset.
			for _, seg := range result.Segments {
				if time.Duration(seg.Start) >= windowStartOffset {
					segments = append(segments, seg)
				}
			}
		}
	}

	// Deduplicate: only emit segments whose end time is strictly after the
	// last segment we already returned to the caller.
	var newSegments []*schema.Segment
	for _, seg := range segments {
		if time.Duration(seg.End) > s.lastSegmentEnd {
			newSegments = append(newSegments, seg)
		}
	}
	if len(newSegments) > 0 {
		s.lastSegmentEnd = time.Duration(newSegments[len(newSegments)-1].End)
	}

	// Trim the audio buffer. Keep the last keepSamples as overlap.
	s.trimBuffer()

	// Reset the new-sample counter.
	s.newSamples = 0

	return newSegments, nil
}

// trimBuffer retains at most keepSamples at the end of audioBuffer and
// advances windowOffset to reflect the dropped samples.
func (s *StreamSession) trimBuffer() {
	if len(s.audioBuffer) <= s.keepSamples {
		// Nothing to trim.
		return
	}
	dropped := len(s.audioBuffer) - s.keepSamples
	s.windowOffset += samplesToDuration(dropped)
	s.audioBuffer = s.audioBuffer[dropped:]
}

//////////////////////////////////////////////////////////////////////////////
// HELPERS

// msToSamples converts a duration in milliseconds to a sample count at 16 kHz.
func msToSamples(ms int) int {
	return ms * whisper.SampleRate / 1000
}

// samplesToDuration converts a sample count at 16 kHz to a time.Duration.
func samplesToDuration(n int) time.Duration {
	return time.Duration(n) * time.Second / time.Duration(whisper.SampleRate)
}

// rmsEnergy computes the root-mean-square energy of the provided samples.
// Returns 0 for an empty slice.
func rmsEnergy(samples []float32) float32 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += float64(s) * float64(s)
	}
	return float32(math.Sqrt(sum / float64(len(samples))))
}
