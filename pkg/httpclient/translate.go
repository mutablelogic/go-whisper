package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Packages
	client "github.com/mutablelogic/go-client"
	gomultipart "github.com/mutablelogic/go-client/pkg/multipart"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// translationResponse handles both JSON and text responses
type translationResponse struct {
	schema.Transcription
}

// Unmarshal implements client.Unmarshaler to handle different content types
func (r *translationResponse) Unmarshal(header http.Header, reader io.Reader) error {
	contentType := header.Get("Content-Type")

	// Check if it's JSON
	if strings.Contains(contentType, "application/json") {
		// Unmarshal as JSON
		return json.NewDecoder(reader).Decode(&r.Transcription)
	}

	// Otherwise treat as plain text
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	r.Text = string(data)
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Translate sends audio data for translation to English. The audio parameter
// should be an io.Reader containing the audio data. The model parameter specifies
// which model to use. Use options to specify other parameters.
//
// Example:
//
//	file, _ := os.Open("audio.mp3")
//	result, err := client.Translate(ctx, "tiny", file,
//	    httpclient.WithPrompt("technical content"))
func (c *Client) Translate(ctx context.Context, model string, audio io.Reader, opts ...Opt) (*schema.Transcription, error) {
	if model == "" {
		return nil, fmt.Errorf("model cannot be empty")
	}
	if audio == nil {
		return nil, fmt.Errorf("audio reader cannot be nil")
	}

	// Apply options to build the request
	opt, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}

	// Determine filename
	filename := opt.Audio.Path
	if filename == "" {
		if f, ok := audio.(*os.File); ok && f != nil {
			filename = filepath.Base(f.Name())
		}
	}
	if filename == "" {
		filename = "audio.mp3"
	}

	// Build TranslateMultipartRequest from the TranscribeMultipartRequest
	req := &schema.TranslateMultipartRequest{
		TranslateRequest: opt.TranslateRequest,
		Audio: gomultipart.File{
			Path: filename,
			Body: audio,
		},
	}
	req.Model = model

	// Determine accept format
	accept := client.ContentTypeJson
	if opt.format != "" {
		accept = string(opt.format)
	}

	// Create multipart payload
	payload, err := client.NewMultipartRequest(req, accept)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart request: %w", err)
	}

	// Perform request using custom unmarshaler
	var response translationResponse
	if err := c.DoWithContext(ctx, payload, &response, client.OptPath("translate")); err != nil {
		return nil, err
	}

	// Return the response
	return &response.Transcription, nil
}
