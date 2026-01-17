package openai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/mutablelogic/go-whisper/pkg/schema"
)

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Transcribe performs a transcription request in the language of the speech
func (c *Client) Transcribe(ctx context.Context, req TranscriptionRequest) (*TranscriptionResponse, error) {
	var response TranscriptionResponse

	// Set default model
	if req.Model == "" {
		req.Model = Models[0]
	} else if !slices.Contains(Models, req.Model) {
		return nil, fmt.Errorf("invalid model %q, must be one of %v", req.Model, Models)
	}

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
		client.OptPath(TranscribePath),
	}
	if types.PtrBool(req.Stream) {
		opts = append(opts, client.OptTextStreamCallback(func(e client.TextStreamEvent) error {
			// We ignore the event if it is the stream done text
			if strings.TrimSpace(e.Data) == streamDoneText {
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
