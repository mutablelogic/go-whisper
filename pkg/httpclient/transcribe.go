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
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// transcriptionResponse handles both JSON and text responses
type transcriptionResponse struct {
	schema.Transcription
}

// Unmarshal implements client.Unmarshaler to handle different content types
func (r *transcriptionResponse) Unmarshal(header http.Header, reader io.Reader) error {
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

// Transcribe sends audio data for transcription. The audio parameter should be
// an io.Reader containing the audio data. The model parameter specifies which
// model to use. Use options to specify language and other parameters.
//
// Example:
//
//	file, _ := os.Open("audio.mp3")
//	result, err := client.Transcribe(ctx, "tiny", file,
//	    httpclient.WithLanguage("en"))
func (c *Client) Transcribe(ctx context.Context, model string, audio io.Reader, opts ...Opt) (*schema.Transcription, error) {
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

	// Set model and audio
	opt.Model = model
	opt.Audio.Body = audio

	// Infer filename from *os.File if not already set
	if opt.Audio.Path == "" {
		if f, ok := audio.(*os.File); ok && f != nil {
			opt.Audio.Path = filepath.Base(f.Name())
		}
	}
	if opt.Audio.Path == "" {
		opt.Audio.Path = "audio.mp3"
	}

	// Determine accept format
	accept := client.ContentTypeJson
	if opt.format != "" {
		accept = string(opt.format)
	}

	// Create multipart payload
	payload, err := client.NewMultipartRequest(&opt.TranscribeMultipartRequest, accept)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart request: %w", err)
	}

	// Perform request using custom unmarshaler
	var response transcriptionResponse
	if err := c.DoWithContext(ctx, payload, &response, client.OptPath("transcribe")); err != nil {
		return nil, err
	}

	// Return the response
	return &response.Transcription, nil
}
