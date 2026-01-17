package gowhisper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/mutablelogic/go-whisper/pkg/client/openai"
	"github.com/mutablelogic/go-whisper/pkg/schema"
)

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Translate performs a transcription request and returns the result in english
func (c *Client) Translate(ctx context.Context, req TranslationRequest) (*TranscriptionResponse, error) {
	var response TranscriptionResponse

	// Check file, set path if not provided
	if req.File.Body == nil {
		return nil, fmt.Errorf("file is required")
	} else if req.File.Path == "" {
		if f, ok := req.File.Body.(*os.File); ok {
			req.File.Path = filepath.Base(f.Name())
		}
	}

	// Set request options
	opts := []client.RequestOpt{
		client.OptPath(openai.TranslatePath),
	}
	if types.PtrBool(req.Stream) {
		opts = append(opts, client.OptTextStreamCallback(func(e client.TextStreamEvent) error {
			// Ignore non-data events
			if e.Data == "" {
				return nil
			}
			// Parse the event
			var evt schema.Event
			if err := e.Json(&evt); err != nil {
				return fmt.Errorf("failed to parse event: %w", err)
			} else if c.streamfn != nil {
				c.streamfn(evt)
			}

			// Return success
			return nil
		}))
	}

	// Create multipart request, and execute it
	if payload, err := client.NewMultipartRequest(req, client.ContentTypeAny); err != nil {
		return nil, err
	} else if err := c.Do(payload, &response, opts...); err != nil {
		return nil, err
	}

	// Return success
	return &response, nil
}
