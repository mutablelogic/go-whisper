package schema

import (
	"encoding/json"
	"time"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Segment struct {
	Id          int32         `json:"id"`
	Start       time.Duration `json:"start"`
	End         time.Duration `json:"end"`
	Text        string        `json:"text"`
	SpeakerTurn bool          `json:"speaker_turn,omitempty"` // TODO
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Segment) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
