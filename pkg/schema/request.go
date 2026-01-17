package schema

import (
	// Packages
	gomultipart "github.com/mutablelogic/go-client/pkg/multipart"
)

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
