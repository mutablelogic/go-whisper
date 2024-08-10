package client

import "time"

// Request options
type opts struct {
	Language    string        `json:"language,omitempty"`
	SegmentSize time.Duration `json:"segment_size,omitempty"`
	ResponseFmt string        `json:"response_format,omitempty"`
}

type Opt func(*opts) error

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func OptLanguage(language string) Opt {
	return func(o *opts) error {
		o.Language = language
		return nil
	}
}

func OptSegmentSize(v time.Duration) Opt {
	return func(o *opts) error {
		o.SegmentSize = v
		return nil
	}
}

func OptResponseFormat(v string) Opt {
	return func(o *opts) error {
		o.ResponseFmt = v
		return nil
	}
}
