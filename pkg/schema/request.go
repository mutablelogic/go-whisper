package schema

//////////////////////////////////////////////////////////////////////////////
// TYPES

// TranslateRequest represents a request to translate audio to English
type TranslateRequest struct {
	Model       string   `json:"model" help:"Model ID to use for translation"`
	Prompt      *string  `json:"prompt,omitempty" help:"Text to guide the model's style"`
	Temperature *float64 `json:"temperature,omitempty" help:"Sampling temperature (0-1)"`
	Stream      *bool    `json:"stream,omitempty" help:"Enable streaming response"`
	Diarize     *bool    `json:"diarize,omitempty" help:"Identify and separate speakers"`
}

// TranscribeRequest represents a request to transcribe audio
type TranscribeRequest struct {
	TranslateRequest
	Language *string `json:"language,omitempty" help:"Language of the audio (two-letter code)"`
}
