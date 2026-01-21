package schema

import (
	"encoding/json"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ContentTypeSRT = "text/subrip"
	ContentTypeVTT = "text/vtt"
)

var (
	ContentTypeSRTVariants = []string{
		ContentTypeSRT, "application/x-subrip", "text/srt",
	}
	ContentTypeVTTVariants = []string{
		ContentTypeVTT, "application/vtt",
	}
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func stringify[T any](v T) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
