package httpclient

import (
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// FormatType represents the response format for transcription/translation requests
type FormatType string

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	FormatJSON FormatType = types.ContentTypeJSON
	FormatText FormatType = types.ContentTypeTextPlain
	FormatVTT  FormatType = schema.ContentTypeVTT
	FormatSRT  FormatType = schema.ContentTypeSRT
)
