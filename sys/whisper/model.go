package whisper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type client struct {
	http.Client

	root *url.URL
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new client with the specified root URL to the models
func NewClient(abspath string) *client {
	url, err := url.Parse(abspath)
	if err != nil {
		return nil
	}
	return &client{
		root: url,
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *client) Get(ctx context.Context, w io.Writer, path string) error {
	// Construct a URL
	url := resolveUrl(c.root, path)
	if url == nil {
		return fmt.Errorf("invalid path: %s", path)
	}

	// Make a request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}

	// Perform the request
	response, err := c.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Unexpected status code
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	// Write the response
	if _, err := io.Copy(w, response.Body); err != nil {
		return err
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func resolveUrl(base *url.URL, path string) *url.URL {
	// Check arguments
	if base == nil {
		return nil
	}
	if path == "" || path == "/" {
		return base
	}

	// Construct an absolute URL
	query := base.Query()
	rel := url.URL{Path: path}
	abs := base.ResolveReference(&rel)
	abs.RawQuery = query.Encode()

	// Return the absolute URL
	return abs
}
