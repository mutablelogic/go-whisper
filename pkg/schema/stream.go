package schema

//////////////////////////////////////////////////////////////////////////////
// TYPES

// StreamConfig is the configuration message sent by the client as the first
// WebSocket text message when initiating a streaming transcription session.
type StreamConfig struct {
	Model        string   `json:"model"`
	Language     *string  `json:"language,omitempty"`
	Temperature  *float64 `json:"temperature,omitempty"`
	Prompt       *string  `json:"prompt,omitempty"`
	Translate    bool     `json:"translate,omitempty"`
	StepMs       int      `json:"step_ms,omitempty"`
	LengthMs     int      `json:"length_ms,omitempty"`
	KeepMs       int      `json:"keep_ms,omitempty"`
	VadThreshold float32  `json:"vad_threshold,omitempty"`
}

// StreamEvent is a JSON message sent back to the client during a streaming
// transcription session.
type StreamEvent struct {
	Type    string   `json:"type"`
	Segment *Segment `json:"segment,omitempty"`
	Error   string   `json:"error,omitempty"`
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	StreamEventReady   = "stream.ready"
	StreamEventSegment = "stream.segment"
	StreamEventError   = "stream.error"
	StreamEventDone    = "stream.done"
)

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c *StreamConfig) String() string {
	return stringify(*c)
}

func (e *StreamEvent) String() string {
	return stringify(*e)
}
