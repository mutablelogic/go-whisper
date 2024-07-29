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

// HTTP client for downloading models, includes the root URL of the models
type client struct {
	http.Client

	root *url.URL
}

// The client interface is used to download models
type Client interface {
	// Get a file from the server, writing the response to the writer
	// and returning the number of bytes copied
	Get(ctx context.Context, w io.Writer, path string) (int64, error)
}

// If the writer contains a Header method, it can be used to set the
// content type and length of the response, to measure progress
type Writer interface {
	io.Writer

	// Returns the header of the response. If the return value is
	// not nil, then the Get method will end before the response
	// data is written
	Header(http.Header) error
}

// Body reader which can be used to read the response body
// and return an error if the context is cancelled early
type reader struct {
	io.Reader
	ctx context.Context
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new client with the specified root URL for downloading models
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

// Get a model from the server, writing the response to the writer
func (c *client) Get(ctx context.Context, w io.Writer, path string) (int64, error) {
	// Construct a URL
	url := resolveUrl(c.root, path)
	if url == nil {
		return 0, fmt.Errorf("invalid path: %s", path)
	}

	// Make a request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return 0, err
	}

	// Perform the request
	response, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	// Unexpected status code
	if response.StatusCode != http.StatusOK {
		return 0, &HTTPError{response.StatusCode, response.Status}
	}

	// Set response header
	if writer, ok := w.(Writer); ok {
		if err := writer.Header(response.Header); err != nil {
			return 0, err
		}
	}

	// Write the response, cancelling if the context is cancelled or deadline
	// is exceeded. Return number of bytes copied
	return io.Copy(w, &reader{response.Body, ctx})
}

///////////////////////////////////////////////////////////////////////////////
// READER INTERFACE

func (r *reader) Read(p []byte) (n int, err error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.Reader.Read(p)
	}
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
