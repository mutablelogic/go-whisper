package whisper

import "encoding/json"

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Transcription struct {
	Task     string                 `json:"task,omitempty"`
	Language string                 `json:"language,omitempty" writer:",width:8"`
	Duration float64                `json:"duration,omitempty" writer:",width:8,right"`
	Text     string                 `json:"text" writer:",width:60,wrap"`
	Segments []TranscriptionSegment `json:"segments,omitempty" writer:",width:40,wrap"`
}

type TranscriptionSegment struct {
	Id    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
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
