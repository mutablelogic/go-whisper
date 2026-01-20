package schema

import "encoding/json"

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Event struct {
	Type  string          `json:"type"`
	Delta string          `json:"delta,omitempty"` // transcript.text.delta
	Text  string          `json:"text,omitempty"`  // transcript.text.done and transcript.text.language
	JSON  json.RawMessage `json:"json,omitempty"`  // transcript.text.delta and transcript.text.done when format = json or verbose_json
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
)

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e Event) String() string {
	return stringify(e)
}
