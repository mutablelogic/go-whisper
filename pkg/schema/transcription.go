package schema

import (
	"encoding/json"
	"time"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Timestamp time.Duration

type Transcription struct {
	Task     string     `json:"task,omitempty"`
	Language string     `json:"language,omitempty" writer:",width:8"`
	Duration Timestamp  `json:"duration,omitempty" writer:",width:8,right"`
	Text     string     `json:"text,omitempty" writer:",width:60,wrap"`
	Segments []*Segment `json:"segments,omitempty" writer:",width:40,wrap"`
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *Transcription) String() string {
	return stringify(*t)
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
