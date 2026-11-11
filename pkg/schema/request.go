package schema

import (
	// Packages
	gomultipart "github.com/mutablelogic/go-client/pkg/multipart"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// TranslateRequest represents a request to translate audio to English, from
// the language it's in, when it's known, or else the language is detected
type TranslateRequest struct {
	Model       string   `json:"model" help:"Model ID to use for translation"`
	Prompt      *string  `json:"prompt,omitempty" help:"Text to guide the model's style"`
	Temperature *float64 `json:"temperature,omitempty" help:"Sampling temperature (0-1)"`
	Filename    *string  `json:"filename,omitempty" help:"Optional filename with extension for the audio file"`
	Language    *string  `json:"language,omitempty" help:"Language the audio is in (two-letter code), which is detected when it isn't given. A translation is from it, into English."`
}

// TranscribeRequest represents a request to transcribe audio
type TranscribeRequest struct {
	TranslateRequest
	Diarize *bool `json:"diarize,omitempty" help:"Identify and separate speakers"`
}

// DownloadModelRequest represents a request to download a model
type DownloadModelRequest struct {
	Model string `json:"model" help:"Model ID to download"`
}

// TranslateMultipartRequest wraps TranslateRequest with an audio file for multipart form handling
type TranslateMultipartRequest struct {
	TranslateRequest
	Audio gomultipart.File `json:"audio"`
}

// TranscribeMultipartRequest wraps TranscribeRequest with an audio file for multipart form handling
type TranscribeMultipartRequest struct {
	TranscribeRequest
	Audio gomultipart.File `json:"audio"`
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r *TranslateRequest) String() string {
	return stringify(*r)
}

func (r *TranscribeRequest) String() string {
	return stringify(*r)
}

func (r *DownloadModelRequest) String() string {
	return stringify(*r)
}
