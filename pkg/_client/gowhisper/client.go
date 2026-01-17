package gowhisper

import (
	"context"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
	streamfn func(schema.Event)
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new client, with the elevenslabs token
func New(endpoint string, opts ...client.ClientOpt) (*Client, error) {
	opts = append([]client.ClientOpt{
		client.OptEndpoint(endpoint),
	}, opts...)
	if client, err := client.New(opts...); err != nil {
		return nil, err
	} else {
		return &Client{Client: client}, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) SetStreamCallback(fn func(schema.Event)) {
	c.streamfn = fn
}

func (c *Client) Ping(ctx context.Context) error {
	return c.DoWithContext(ctx, client.MethodGet, nil, client.OptPath("health"))
}

///////////////////////////////////////////////////////////////////////////////
// MODELS

func (c *Client) ListModels(ctx context.Context) ([]schema.Model, error) {
	var response struct {
		Models []schema.Model `json:"models"`
	}
	if err := c.DoWithContext(ctx, client.MethodGet, &response, client.OptPath("models")); err != nil {
		return nil, err
	}

	// Return success
	return response.Models, nil
}
