package task

import (
	"time"

	// Packages
	"github.com/mutablelogic/go-whisper/pkg/whisper/schema"
	"github.com/mutablelogic/go-whisper/sys/whisper"
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newSegment(ts time.Duration, seg *whisper.Segment) *schema.Segment {
	// Dumb copy function
	return &schema.Segment{
		Id:          seg.Id,
		Text:        seg.Text,
		Start:       seg.T0 + ts,
		End:         seg.T1 + ts,
		SpeakerTurn: seg.SpeakerTurn,
	}
}
