package httpclient

///////////////////////////////////////////////////////////////////////////////
// TYPES

// FormatType represents the response format for transcription/translation requests
type FormatType string

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// FormatJSON requests JSON response (default)
	FormatJSON FormatType = "application/json"

	// FormatText requests plain text response
	FormatText FormatType = "text/plain"

	// FormatVTT requests WebVTT subtitle format
	FormatVTT FormatType = "text/vtt"

	// FormatSRT requests SubRip subtitle format
	FormatSRT FormatType = "application/x-subrip"
)
