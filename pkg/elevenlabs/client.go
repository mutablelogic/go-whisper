// https://elevenlabs.io/docs/overview
package elevenlabs

import (
	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new client, with the elevenslabs token
func New(apikey string, opts ...client.ClientOpt) (*Client, error) {
	opts = append([]client.ClientOpt{
		client.OptEndpoint(Endpoint),
		client.OptHeader("xi-api-key", apikey),
	}, opts...)
	if client, err := client.New(opts...); err != nil {
		return nil, err
	} else {
		return &Client{Client: client}, nil
	}
}
