package transcription

import (
	"encoding/json"
	"time"

	"github.com/mutablelogic/go-whisper/sys/whisper"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Transcription struct {
	Task     string        `json:"task,omitempty"`
	Language string        `json:"language,omitempty" writer:",width:8"`
	Duration time.Duration `json:"duration,omitempty" writer:",width:8,right"`
	Text     string        `json:"text" writer:",width:60,wrap"`
	Segments []Segment     `json:"segments,omitempty" writer:",width:40,wrap"`
}

type Segment struct {
	Id          int32         `json:"id"`
	Start       time.Duration `json:"start"`
	End         time.Duration `json:"end"`
	Text        string        `json:"text"`
	SpeakerTurn bool          `json:"speaker_turn,omitempty"`
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new transcription from a context
func New() *Transcription {
	return new(Transcription)
}

func NewSegment(ts time.Duration, seg *whisper.Segment) *Segment {
	// Dumb copy function
	return &Segment{
		Id:          seg.Id,
		Text:        seg.Text,
		Start:       seg.T0 + ts,
		End:         seg.T1 + ts,
		SpeakerTurn: seg.SpeakerTurn,
	}
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

func (s *Segment) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Append a transcription to the current transcription, with the offset timestamp
func (t *Transcription) Append(ctx *whisper.Context, ts time.Duration) {
	// Append the segment text
	for i := 0; i < ctx.NumSegments(); i++ {
		seg := ctx.Segment(i)
		if t.Text == "" {
			t.Text = seg.Text
		} else {
			t.Text += " " + seg.Text
		}
	}
}
