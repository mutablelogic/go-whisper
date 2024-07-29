package whisper

import "errors"

var (
	ErrTranscriptionFailed = errors.New("whisper_full failed")
)

type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}
