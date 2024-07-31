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
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	// We convert durations into float64 seconds
	return json.Marshal(time.Duration(t).Seconds())
}
