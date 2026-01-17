package schema

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Segment struct {
	Id          int32     `json:"id"`
	Start       Timestamp `json:"start"`
	End         Timestamp `json:"end"`
	Text        string    `json:"text"`
	Tokens      []string  `json:"tokens,omitempty"`       // TODO
	Speaker     string    `json:"speaker,omitempty"`      // TODO
	SpeakerTurn bool      `json:"speaker_turn,omitempty"` // TODO
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Segment) String() string {
	return stringify(*s)
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (seg *Segment) WriteSRT(w io.Writer, offset time.Duration) {
	fmt.Fprintf(w, "%d\n%s --> %s\n", seg.Id, tsToSrt(time.Duration(seg.Start)+offset), tsToSrt(time.Duration(seg.End)+offset))
	if seg.Speaker != "" {
		fmt.Fprintf(w, "[%s] ", seg.Speaker)
	} else if seg.SpeakerTurn {
		fmt.Fprintf(w, "[SPEAKER] ")
	}
	fmt.Fprintf(w, "%s\n\n", strings.TrimSpace(seg.Text))
}

func (seg *Segment) WriteVTT(w io.Writer, offset time.Duration) {
	text := strings.TrimSpace(seg.Text)
	if text != "" {
		fmt.Fprintf(w, "%s --> %s\n", tsToVtt(time.Duration(seg.Start)+offset), tsToVtt(time.Duration(seg.End)+offset))
		var opener, closer string
		if seg.Speaker != "" {
			opener = "<v " + seg.Speaker + ">"
			closer = "</v>"
		} else if seg.SpeakerTurn {
			opener = "<v " + "speaker" + ">"
			closer = "</v>"
		}
		fmt.Fprintf(w, "%s%s%s\n\n", opener, text, closer)
	}
}

var (
	reToken = regexp.MustCompile(`^\s*\[.*\]$`)
)

func (seg *Segment) WriteText(w io.Writer) {
	if isToken := reToken.MatchString(seg.Text); isToken && seg.Id > 0 {
		fmt.Fprint(w, "\n\n"+strings.TrimSpace(seg.Text)+"\n")
		return
	}
	if seg.Speaker != "" {
		fmt.Fprintf(w, "\n\n[%s] ", seg.Speaker)
	} else if seg.SpeakerTurn {
		fmt.Fprint(w, "\n\n[SPEAKER] ")
	}
	if seg.Id > 0 {
		fmt.Fprint(w, seg.Text)
	} else {
		fmt.Fprint(w, strings.TrimSpace(seg.Text))
	}
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func tsToSrt(ts time.Duration) string {
	// Extract hours, minutes, seconds, and milliseconds from the duration
	hours := int(ts.Hours())
	minutes := int(ts.Minutes()) % 60
	seconds := int(ts.Seconds()) % 60
	milliseconds := int(ts.Milliseconds()) % 1000

	// Format the timestamp in the SRT format
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, milliseconds)
}

func tsToVtt(ts time.Duration) string {
	// Extract hours, minutes, seconds, and milliseconds from the duration
	hours := int(ts.Hours())
	minutes := int(ts.Minutes()) % 60
	seconds := int(ts.Seconds()) % 60
	milliseconds := int(ts.Milliseconds()) % 1000

	// Format the timestamp in the SRT format
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}
