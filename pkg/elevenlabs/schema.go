package elevenlabs

import (
	"encoding/json"

	// Packages
	"github.com/mutablelogic/go-client/pkg/multipart"
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/mutablelogic/go-whisper/pkg/schema"
)

/////////////////////////////////////////////////////////////////////////////////
// TYPES

type TranscribeRequest struct {
	Model          string           `json:"model_id"` // scribe_v1, scribe_v1_experimental
	File           multipart.File   `json:"file"`
	Language       *string          `json:"language_code,omitempty"`
	TagAudioEvents *bool            `json:"tag_audio_events,omitempty"`
	NumSpeakers    *uint64          `json:"num_speakers,omitempty"`
	Timestamps     *string          `json:"timestamps_granularity,omitempty"` // none, word, character
	Diarize        *bool            `json:"diarize,omitempty"`
	DiarizationThr *float64         `json:"diarization_threshold,omitempty"`
	FileFormat     *string          `json:"file_format,omitempty"` // pcm_s16le_16, other
	CloudURL       *string          `json:"cloud_storage_url,omitempty"`
	EnableLogging  *bool            `json:"enable_logging,omitempty"`
	Webhook        *bool            `json:"webhook,omitempty"`
	WebhookID      *string          `json:"webhook_id,omitempty"`
	WebhookMeta    json.RawMessage  `json:"webhook_metadata,omitempty"`
	Temperature    *float64         `json:"temperature,omitempty"`
	Seed           *int64           `json:"seed,omitempty"`
	UseMultiChan   *bool            `json:"use_multi_channel,omitempty"`
	EntityDetect   any              `json:"entity_detection,omitempty"`
	Keyterms       []string         `json:"keyterms,omitempty"`
	AdditionalFmt  []map[string]any `json:"additional_formats,omitempty"`
}

type TranscribeResponse struct {
	Language    string           `json:"language_code"`
	Probability float64          `json:"language_probability"`
	Text        string           `json:"text"`
	Words       []TranscribeWord `json:"words,omitempty"` // Only present if timestamps_granularity is set to word or character
}

type TranscribeWord struct {
	Text    string           `json:"text"`
	Type    string           `json:"type"`              // word, spacing, audio_event
	Logprob float64          `json:"logprob,omitempty"` // -inf -> 0.0
	Start   float64          `json:"start"`
	End     float64          `json:"end"`
	Speaker *string          `json:"speaker_id,omitempty"`
	Chars   []TranscribeChar `json:"characters,omitempty"` // Only present if timestamps_granularity is set to character
}

type TranscribeChar struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

/////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	Endpoint       = "https://api.elevenlabs.io/v1"
	TranscribePath = "speech-to-text"             // Endpoint for transcription
	TranscriptPath = "speech-to-text/transcripts" // Base path for transcript get/delete
)

var (
	Models = []string{"scribe_v2", "scribe_v1", "scribe_v1_experimental"}
)

/////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s TranscribeResponse) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (s TranscribeWord) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return segments of a transcription response
func (r *TranscribeResponse) Segments() *schema.Transcription {
	t := &schema.Transcription{
		Task:     "transcribe",
		Language: r.Language,
		Text:     r.Text,
	}

	// Current segment
	for _, word := range r.Words {
		// Append to current segment or create a new one
		t.Segments = appendSegment(t.Segments, word)
	}

	// Return transcription
	return t
}

/////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func appendSegment(slice []*schema.Segment, word TranscribeWord) []*schema.Segment {
	// Determine the speaker for the segment
	var speaker string
	if word.Type == "audio_event" {
		speaker = "audio_event"
	} else if word.Speaker != nil {
		speaker = types.PtrString(word.Speaker)
	}

	if len(slice) == 0 {
		// Create a new segment if none exist
		seg := &schema.Segment{
			Id:      1,
			Start:   schema.SecToTimestamp(word.Start),
			End:     schema.SecToTimestamp(word.End),
			Text:    word.Text,
			Speaker: speaker,
		}
		return append(slice, seg)
	}

	// Get the current segment
	seg := slice[len(slice)-1]

	// Determine if we need a new segment
	var new bool
	if word.Type == "audio_event" || seg.Speaker == "audio_event" {
		// If the word is an audio event...
		new = true
	} else if word.Speaker != nil && seg.Speaker != types.PtrString(word.Speaker) {
		// If the speaker has changed...
		new = true
	}

	if new {
		slice = append(slice, &schema.Segment{
			Id:      seg.Id + 1,
			Start:   schema.SecToTimestamp(word.Start),
			End:     schema.SecToTimestamp(word.End),
			Text:    word.Text,
			Speaker: speaker,
		})
	} else {
		seg.End = schema.SecToTimestamp(word.End)
		seg.Text += word.Text
	}

	// Return segment
	return slice
}
