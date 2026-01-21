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
	"github.com/mutablelogic/go-server/pkg/types"
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
func (c *Client) Translate(ctx context.Context, model string, r io.Reader, opts ...Opt) (*schema.Transcription, error) {
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

	// Build TranslateMultipartRequest from the TranscribeMultipartRequest
	req := schema.TranslateMultipartRequest{
		TranslateRequest: opt.TranslateRequest,
		Audio:            opt.Audio,
	}
	req.Model = model

	// Create multipart payload
	payload, err := client.NewMultipartRequest(&req, types.ContentTypeJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart request: %w", err)
	}

	// If streaming callback provided, handle streaming
	var reqOpts []client.RequestOpt = []client.RequestOpt{
		client.OptPath("translate"),
	}
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
	}

	// Perform the request
	if err := c.DoWithContext(ctx, payload, &response, reqOpts...); err != nil {
		return nil, err
	}

	// Return the accumulated result
	return &response, nil
}
