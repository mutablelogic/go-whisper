package schema

import (
	"encoding/json"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Segment struct {
	Id          int32     `json:"id"`
	Start       Timestamp `json:"start"`
	End         Timestamp `json:"end"`
	Text        string    `json:"text"`
	SpeakerTurn bool      `json:"speaker_turn,omitempty"` // TODO
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
