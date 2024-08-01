package task

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	// Packages
	"github.com/mutablelogic/go-whisper/pkg/whisper/schema"
	"github.com/mutablelogic/go-whisper/sys/whisper"
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newSegment(ts time.Duration, offset int32, seg *whisper.Segment) *schema.Segment {
	// Dumb copy function
	return &schema.Segment{
		Id:          offset + seg.Id,
		Text:        seg.Text,
		Start:       schema.Timestamp(seg.T0 + ts),
		End:         schema.Timestamp(seg.T1 + ts),
		SpeakerTurn: seg.SpeakerTurn,
	}
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WriteSegmentSrt(w io.Writer, seg *schema.Segment) {
	fmt.Fprintf(w, "%d\n%s --> %s\n", seg.Id, tsToSrt(time.Duration(seg.Start)), tsToSrt(time.Duration(seg.End)))
	if seg.SpeakerTurn {
		fmt.Fprintf(w, "[SPEAKER] ")
	}
	fmt.Fprintf(w, "%s\n\n", strings.TrimSpace(seg.Text))
}

func WriteSegmentVtt(w io.Writer, seg *schema.Segment) {
	fmt.Fprintf(w, "%s --> %s\n", tsToVtt(time.Duration(seg.Start)), tsToVtt(time.Duration(seg.End)))
	if seg.SpeakerTurn {
		fmt.Fprintf(w, "<v Speaker>")
	}
	fmt.Fprintf(w, "%s\n\n", strings.TrimSpace(seg.Text))
}

var (
	reToken = regexp.MustCompile(`^\s*\[.*\]$`)
)

func WriteSegmentText(w io.Writer, seg *schema.Segment) {
	if isToken := reToken.MatchString(seg.Text); isToken && seg.Id > 0 {
		fmt.Fprint(w, "\n\n"+strings.TrimSpace(seg.Text)+"\n")
		return
	}
	if seg.SpeakerTurn {
		fmt.Fprint(w, "\n\n[SPEAKER]")
	}
	if seg.Id > 0 || seg.SpeakerTurn {
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
