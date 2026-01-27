package httpclient

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	// Packages
	client "github.com/mutablelogic/go-client"
	types "github.com/mutablelogic/go-server/pkg/types"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Transcribe sends audio data for transcription. The audio parameter should be
// an io.Reader containing the audio data. The model parameter specifies which
// model to use. Use options to specify language and other parameters.
//
// Example:
//
//	file, _ := os.Open("audio.mp3")
//	result, err := client.Transcribe(ctx, "tiny", file,
//	    httpclient.WithLanguage("en"))
func (c *Client) Transcribe(ctx context.Context, model string, r io.Reader, opts ...Opt) (*schema.Transcription, error) {
	// Apply options to build the request
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Set model and audio
	opt.Model = model
	opt.Audio.Body = r
	if opt.Audio.Path == "" {
		// Infer filename from *os.File if not already set
		if f, ok := r.(*os.File); ok && f != nil {
			opt.Audio.Path = filepath.Base(f.Name())
		}
	}

	// Create multipart payload
	payload, err := client.NewStreamingMultipartRequest(&opt.TranscribeMultipartRequest, types.ContentTypeJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart request: %w", err)
	}

	// Append path to request options
	reqOpts := append(opt.reqOpts, client.OptPath("transcribe"))

	// If streaming callback provided, handle streaming
	var response schema.Transcription
	if opt.segmentCallback != nil {
		reqOpts = append(reqOpts, client.OptReqHeader("Accept", "text/event-stream"))
		reqOpts = append(reqOpts, client.OptTextStreamCallback(func(evt client.TextStreamEvent) error {
			var segment schema.Segment
			if evt.Event == schema.TranscribeStreamDeltaType {
				if err := evt.Json(&segment); err != nil {
					return err
				}
				return opt.segmentCallback(&segment)
			}
			return nil
		}))
	} else if opt.format != "" {
		// Set Accept header based on requested format
		reqOpts = append(reqOpts, client.OptReqHeader("Accept", string(opt.format)))
	}

	// Perform the request
	if err := c.DoWithContext(ctx, payload, &response, reqOpts...); err != nil {
		return nil, err
	}

	// Return the accumulated result
	return &response, nil
}
