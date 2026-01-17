package gowhisper

import (
	"github.com/mutablelogic/go-whisper/pkg/client/openai"
)

/////////////////////////////////////////////////////////////////////////////////
// TYPES

type TranslationRequest struct {
	openai.TranslationRequest
	Stream   *bool   `json:"stream,omitempty"`
	Diarize  *bool   `json:"diarize,omitempty"`
	Language *string `json:"language,omitempty"`
}

type TranscriptionRequest struct {
	openai.TranscriptionRequest
	Diarize *bool `json:"diarize,omitempty"`
}

type TranscriptionResponse struct {
	openai.TranscriptionResponse
}
