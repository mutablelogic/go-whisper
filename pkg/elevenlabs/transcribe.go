package elevenlabs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	// Packages
	"github.com/mutablelogic/go-client"
)

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) Transcribe(ctx context.Context, req TranscribeRequest) (*TranscribeResponse, error) {
	var response TranscribeResponse

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

	// Create multipart request, and execute it
	if payload, err := client.NewMultipartRequest(req, client.ContentTypeAny); err != nil {
		return nil, err
	} else if err := c.Do(payload, &response, client.OptPath(TranscribePath)); err != nil {
		return nil, err
	}

	// Return success
	return &response, nil
}

// GetTranscript retrieves a previously created transcript by ID.
func (c *Client) GetTranscript(ctx context.Context, transcriptionID string) (*TranscribeResponse, error) {
	if transcriptionID == "" {
		return nil, fmt.Errorf("transcription_id is required")
	}

	var response TranscribeResponse
	if err := c.DoWithContext(ctx, client.MethodGet, &response, client.OptPath(TranscriptPath, transcriptionID)); err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteTranscript deletes a previously created transcript by ID.
func (c *Client) DeleteTranscript(ctx context.Context, transcriptionID string) error {
	if transcriptionID == "" {
		return fmt.Errorf("transcription_id is required")
	}

	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath(TranscriptPath, transcriptionID))
}
