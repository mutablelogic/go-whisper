package schema

import "encoding/json"

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Event struct {
	Type    string          `json:"type"`
	Delta   string          `json:"delta,omitempty"`   // transcript.text.delta
	Text    string          `json:"text,omitempty"`    // transcript.text.done and transcript.text.language
	JSON    json.RawMessage `json:"json,omitempty"`    // transcript.text.delta and transcript.text.done when format = json or verbose_json
	Id      string          `json:"id,omitempty"`      // Segment ID (for transcript.text.segment)
	Start   Timestamp       `json:"start,omitempty"`   // Segment start time
	End     Timestamp       `json:"end,omitempty"`     // Segment end time
	Speaker string          `json:"speaker,omitempty"` // Speaker label (for diarized segments)
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DownloadStreamProgressType = "download.progress"
	DownloadStreamErrorType    = "download.error"
	DownloadStreamDoneType     = "download.done"
)

const (
	TranscribeStreamDeltaType    = "transcript.text.delta"
	TranscribeStreamDoneType     = "transcript.text.done"
	TranscribeStreamErrorType    = "transcript.text.error"
	TranscribeStreamLanguageType = "transcript.text.language"
	TranscribeStreamSegmentType  = "transcript.text.segment" // Diarized segment event
)

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e Event) String() string {
	return stringify(e)
}
