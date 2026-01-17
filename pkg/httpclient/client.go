package httpclient

import (
	// Packages
	client "github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Client is a whisper HTTP client that wraps the base HTTP client
// and provides typed methods for interacting with the whisper API.
type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new whisper HTTP client with the given base URL and options.
// The url parameter should point to the whisper API endpoint, e.g.
// "http://localhost:8080/api/whisper".
func New(url string, opts ...client.ClientOpt) (*Client, error) {
	c := new(Client)
	if client, err := client.New(append(opts, client.OptEndpoint(url))...); err != nil {
		return nil, err
	} else {
		c.Client = client
	}
	return c, nil
}
