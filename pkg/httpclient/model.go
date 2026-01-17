package httpclient

import (
	"context"
	"fmt"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListModels returns a list of all available models from the whisper API.
func (c *Client) ListModels(ctx context.Context) ([]*schema.Model, error) {
	req := client.NewRequest()

	// Perform request
	var response []*schema.Model
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("model")); err != nil {
		return nil, err
	}

	// Return the response
	return response, nil
}

// GetModel retrieves a specific model by its ID.
func (c *Client) GetModel(ctx context.Context, id string) (*schema.Model, error) {
	if id == "" {
		return nil, fmt.Errorf("model id cannot be empty")
	}

	req := client.NewRequest()

	// Perform request
	var response schema.Model
	if err := c.DoWithContext(ctx, req, &response, client.OptPath("model", id)); err != nil {
		return nil, err
	}

	// Return the response
	return &response, nil
}

// DownloadModel downloads a model with the given ID. The optional progress
// callback receives current and total bytes as the download progresses.
func (c *Client) DownloadModel(ctx context.Context, id string, progressFn func(cur, total uint64)) (*schema.Model, error) {
	if id == "" {
		return nil, fmt.Errorf("model id cannot be empty")
	}

	// Create request body
	reqBody := schema.DownloadModelRequest{
		Model: id,
	}

	req, err := client.NewJSONRequest(reqBody)
	if err != nil {
		return nil, err
	}

	// Set up options for the request
	opts := []client.RequestOpt{client.OptPath("model")}

	// If progress callback is provided, request streaming and set up text stream handler
	var response schema.Model
	if progressFn != nil {
		opts = append(opts, client.OptReqHeader("Accept", "text/event-stream"))
		opts = append(opts, client.OptTextStreamCallback(func(evt client.TextStreamEvent) error {
			switch evt.Event {
			case "progress":
				// Parse progress data
				var progress struct {
					Current uint64  `json:"current"`
					Total   uint64  `json:"total"`
					Percent float64 `json:"percent,omitempty"`
				}
				if err := evt.Json(&progress); err == nil {
					progressFn(progress.Current, progress.Total)
				}
			case "done":
				// Parse final model data - populate response
				if err := evt.Json(&response); err != nil {
					return fmt.Errorf("failed to parse model data: %w", err)
				}
			case "error":
				return fmt.Errorf("download error: %s", evt.Data)
			}
			return nil
		}))
	}

	// Perform request (streaming or regular based on whether progressFn was provided)
	if err := c.DoWithContext(ctx, req, &response, opts...); err != nil {
		return nil, err
	}

	return &response, nil
}

// DeleteModel deletes the model identified by id.
func (c *Client) DeleteModel(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("model id cannot be empty")
	}

	req := client.NewRequestEx(http.MethodDelete, "")

	// Perform request - expect 204 No Content
	if err := c.DoWithContext(ctx, req, nil, client.OptPath("model", id)); err != nil {
		return err
	}

	return nil
}
