package gowhisper

import (
	"context"
	"errors"
	"net/url"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-whisper/pkg/schema"
)

func (c *Client) DeleteModel(ctx context.Context, model string) error {
	return c.DoWithContext(ctx, client.MethodDelete, nil, client.OptPath("models", model))
}

func (c *Client) DownloadModel(ctx context.Context, path string, fn func(cur, total uint64)) (*schema.Model, error) {
	var req struct {
		Path string `json:"path"`
	}
	type resp struct {
		schema.Model
		Status    string `json:"status"`
		Total     uint64 `json:"total,omitempty"`
		Completed uint64 `json:"completed,omitempty"`
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
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &r,
		client.OptPath("models"),
		client.OptQuery(query),
		client.OptNoTimeout(),
		client.OptTextStreamCallback(func(evt client.TextStreamEvent) error {
			switch evt.Event {
			case schema.DownloadStreamProgressType:
				var r resp
				if err := evt.Json(&r); err != nil {
					return err
				} else {
					fn(r.Completed, r.Total)
				}
			case schema.DownloadStreamErrorType:
				var errstr string
				if err := evt.Json(&errstr); err != nil {
					return err
				} else {
					return errors.New(errstr)
				}
			case schema.DownloadStreamDoneType:
				if err := evt.Json(&r); err != nil {
					return err
				}
			}
			return nil
		}),
	); err != nil {
		return nil, err
	}

	// Return success
	return &r.Model, nil
}
