package openai

import (
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
func New(apikey string, opts ...client.ClientOpt) (*Client, error) {
	opts = append([]client.ClientOpt{
		client.OptEndpoint(Endpoint),
		client.OptReqToken(client.Token{
			Scheme: "Bearer",
			Value:  apikey,
		}),
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
