package schema

import (
	"bytes"
	"encoding/json"
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
	Start       Timestamp `json:"start,omitempty"`
	End         Timestamp `json:"end,omitempty"`
	Text        string    `json:"text"`
	Tokens      []string  `json:"tokens,omitempty"`       // TODO
	Speaker     string    `json:"speaker,omitempty"`      // TODO
	SpeakerTurn bool      `json:"speaker_turn,omitempty"` // TODO
}

// SegmentWriter defines an interface for writing segments in realtime
type SegmentWriter interface {
	Write(seg *Segment)
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
	// Write header if first segment
	if seg.Id == 0 {
		// Note: VTT header must be followed by a blank line
		fmt.Fprint(w, "WEBVTT\n\n")
	}

	// Write text segment
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
	} else if seg.Id > 0 {
		// Add newline between segments when there's no speaker label
		fmt.Fprint(w, "\n")
	}
	fmt.Fprint(w, strings.TrimSpace(seg.Text))
}

func (seg *Segment) WriteJSON(w io.Writer) {
	// Write header if first segment, or connecting comma otherwise
	if seg.Id == 0 {
		fmt.Fprint(w, "[\n  ")
	} else {
		fmt.Fprint(w, ",")
	}
	if json, err := json.MarshalIndent(seg, "  ", "  "); err == nil {
		json = bytes.TrimSpace(json)
		fmt.Fprint(w, string(json))
	}
}

func WriteJSONTrailer(w io.Writer) {
	fmt.Fprint(w, "\n]\n")
}

func WriteTextTrailer(w io.Writer) {
	fmt.Fprint(w, "\n\n")
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
