package schema

import (
	"encoding/json"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Segment struct {
	Id          int32     `json:"id" writer:",right,width:5"`
	Start       Timestamp `json:"start" writer:",right,width:5"`
	End         Timestamp `json:"end" writer:",right,width:5"`
	Text        string    `json:"text" writer:",wrap,width:70"`
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
