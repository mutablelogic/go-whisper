package schema

import (
	"encoding/json"
	"time"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Timestamp time.Duration

type Transcription struct {
	TranscriptionSummary
	Text     string     `json:"text,omitempty" writer:",width:60,wrap"`
	Segments []*Segment `json:"segments,omitempty" writer:",width:40,wrap"`
}

// TranscriptionSummary is a lightweight version of Transcription for streaming
// done events. It excludes segments since they are sent individually during streaming.
// This prevents the SSE event from exceeding bufio.Scanner's buffer limit.
type TranscriptionSummary struct {
	Task     string    `json:"task,omitempty"`
	Language string    `json:"language,omitempty"`
	Duration Timestamp `json:"duration,omitempty"`
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *Transcription) String() string {
	return stringify(*t)
}

// Summary returns a TranscriptionSummary suitable for streaming done events
func (t *Transcription) Summary() *TranscriptionSummary {
	return &TranscriptionSummary{
		Task:     t.Task,
		Language: t.Language,
		Duration: t.Duration,
	}
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	// We convert durations into float64 seconds
	return json.Marshal(time.Duration(t).Seconds())
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var seconds float64
	if err := json.Unmarshal(data, &seconds); err != nil {
		return err
	}
	*t = Timestamp(time.Duration(seconds * float64(time.Second)))
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func SecToTimestamp(sec float64) Timestamp {
	// Convert seconds to Timestamp
	return Timestamp(time.Duration(sec * float64(time.Second)))
}
