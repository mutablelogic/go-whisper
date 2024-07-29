package client

import (
	"context"
	"io"
	"net/url"
	"os"
	"path/filepath"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/multipart"
	"github.com/mutablelogic/go-server/pkg/httprequest"
	"github.com/mutablelogic/go-whisper/pkg/whisper"
	"github.com/mutablelogic/go-whisper/pkg/whisper/model"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new client, with the endpoint of the whisper service
// ie, http://localhost:8080/v1
func New(endpoint string, opts ...client.ClientOpt) (*Client, error) {
	if client, err := client.New(append(opts, client.OptEndpoint(endpoint))...); err != nil {
		return nil, err
	} else {
		return &Client{Client: client}, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// MODELS

func (c *Client) ListModels(ctx context.Context) ([]model.Model, error) {
	var models struct {
		Models []model.Model `json:"models"`
	}
	if err := c.DoWithContext(ctx, client.MethodGet, &models, client.OptPath("models")); err != nil {
		return nil, err
	}
	// Return success
	return models.Models, nil
}

func (c *Client) DeleteModel(ctx context.Context, model string) error {
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("models", model))
}

func (c *Client) DownloadModel(ctx context.Context, path string, fn func(status string, cur, total int64)) (model.Model, error) {
	var req struct {
		Path string `json:"path"`
	}
	type resp struct {
		model.Model
		Status    string `json:"status"`
		Total     int64  `json:"total,omitempty"`
		Completed int64  `json:"completed,omitempty"`
	}

	// stream=true for progress reports
	query := url.Values{}
	if fn != nil {
		query.Set("stream", "true")
	}

	// Download the model
	req.Path = path

	var r resp
	if payload, err := client.NewJSONRequest(req); err != nil {
		return model.Model{}, err
	} else if err := c.DoWithContext(ctx, payload, &r,
		client.OptPath("models"),
		client.OptQuery(query),
		client.OptNoTimeout(),
		client.OptJsonStreamCallback(func(v any) error {
			if v, ok := v.(*resp); ok && fn != nil {
				fn(v.Status, v.Completed, v.Total)
			}
			return nil
		}),
	); err != nil {
		return model.Model{}, err
	}

	// Return success
	return r.Model, nil
}

func (c *Client) Transcribe(ctx context.Context, model string, r io.Reader, opt ...Opt) (*whisper.Transcription, error) {
	var request struct {
		File  multipart.File `json:"file"`
		Model string         `json:"model"`
		opts
	}
	var response whisper.Transcription

	// Get the name from the io.Reader
	name := ""
	if f, ok := r.(*os.File); ok {
		name = filepath.Base(f.Name())
	}

	// Create the request
	request.Model = model
	request.File = multipart.File{
		Path: name,
		Body: r,
	}
	for _, o := range opt {
		if err := o(&request.opts); err != nil {
			return nil, err
		}
	}

	// Request->Response
	if payload, err := client.NewMultipartRequest(request, httprequest.ContentTypeFormData); err != nil {
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &response, client.OptPath("audio/transcriptions"), client.OptNoTimeout()); err != nil {
		return nil, err
	}

	// Return success
	return &response, nil
}

func (c *Client) Translate(ctx context.Context, model string, r io.Reader, opt ...Opt) (*whisper.Transcription, error) {
	var request struct {
		File  multipart.File `json:"file"`
		Model string         `json:"model"`
		opts
	}
	var response whisper.Transcription

	// Get the name from the io.Reader
	name := ""
	if f, ok := r.(*os.File); ok {
		name = filepath.Base(f.Name())
	}

	// Create the request
	request.Model = model
	request.File = multipart.File{
		Path: name,
		Body: r,
	}
	for _, o := range opt {
		if err := o(&request.opts); err != nil {
			return nil, err
		}
	}

	// Request->Response
	if payload, err := client.NewMultipartRequest(request, httprequest.ContentTypeFormData); err != nil {
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &response, client.OptPath("audio/translations"), client.OptNoTimeout()); err != nil {
		return nil, err
	}

	// Return success
	return &response, nil
}
