package whisper

import "errors"

var (
	ErrTranscriptionFailed = errors.New("whisper_full failed")
	ErrBadParameter        = errors.New("invalid parameter")
)

type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}
