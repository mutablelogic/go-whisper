package openai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Transcribe performs a transcription request in the language of the speech.
// If streamfn is provided, streaming mode is enabled and events will be passed to the callback.
func (c *Client) Transcribe(ctx context.Context, req TranscriptionRequest, streamfn func(schema.Event)) (*TranscriptionResponse, error) {
	var response TranscriptionResponse

	// Set default model
	if req.Model == "" {
		req.Model = Models[0]
	} else if !slices.Contains(Models, req.Model) && !slices.Contains(DiarizeModels, req.Model) {
		return nil, fmt.Errorf("invalid model %q, must be one of %v or %v", req.Model, Models, DiarizeModels)
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

	// Build the actual request to send - wrap with stream field if streaming
	type streamingRequest struct {
		TranscriptionRequest
		Stream bool `json:"stream,omitempty"`
	}
	actualReq := streamingRequest{TranscriptionRequest: req}

	if streamfn != nil {
		actualReq.Stream = true
		opts = append(opts, client.OptTextStreamCallback(func(e client.TextStreamEvent) error {
			// We ignore the event if it is the stream done text
			if strings.TrimSpace(e.Data) == streamDoneText {
				return nil
			}

			// Parse the event
			var evt schema.Event
			if err := e.Json(&evt); err != nil {
				return fmt.Errorf("failed to parse event: %w", err)
			} else if streamfn != nil {
				streamfn(evt)
			}

			// Return success
			return nil
		}))
	}

	// Create multipart request, and execute it
	if payload, err := client.NewStreamingMultipartRequest(actualReq, client.ContentTypeAny); err != nil {
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &response, opts...); err != nil {
		return nil, err
	}

	// Return success
	return &response, nil
}
