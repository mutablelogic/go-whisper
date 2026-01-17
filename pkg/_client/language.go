package client

import (

	// Packages
	"github.com/mutablelogic/go-whisper/pkg/client/elevenlabs"
	"github.com/mutablelogic/go-whisper/pkg/client/openai"
)

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// LanguageCode returns the two-letter (OpenAI) and three-letter (ElevenLabs)
// codes for a given language, or an empty string if the language
// is not recognized.
func LanguageCode(language string) (string, string) {
	language_openai, code_openai := openai.LanguageCode(language)
	language_elevenlabs, code_elevenlabs := elevenlabs.LanguageCode(language)
	if code_elevenlabs == "" && language_openai != "" {
		_, code_elevenlabs = elevenlabs.LanguageCode(language_openai)
	}
	if code_openai == "" && language_elevenlabs != "" {
		_, code_openai = openai.LanguageCode(language_elevenlabs)
	}
	return code_openai, code_elevenlabs
}
