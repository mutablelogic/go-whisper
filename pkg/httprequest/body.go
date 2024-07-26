package httprequest

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"

	"golang.org/x/exp/slices"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ContentTypeJson           = "application/json"
	ContentTypeTextXml        = "text/xml"
	ContentTypeApplicationXml = "application/xml"
	ContentTypeText           = "text/"
	ContentTypeBinary         = "application/octet-stream"
	ContentTypeFormData       = "multipart/form-data"
	ContextTypeUrlEncoded     = "application/x-www-form-urlencoded"
)

const (
	maxMemory = 10 << 20 // 10 MB in-memory cache for multipart form
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ReadBody reads the body of an HTTP request and decodes it into a struct v.
// You can include the mimetypes that are acceptable, otherwise it will read
// the body based on the content type.
func ReadBody(v any, r *http.Request, accept ...string) error {
	// Parse the content type
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	// Check whether we'll accept it
	if len(accept) > 0 && !slices.Contains(accept, contentType) {
		return fmt.Errorf("unexpected content type %q", contentType)
	}

	// Read the body
	switch contentType {
	case ContentTypeJson:
		return readJson(v, r)
	case ContentTypeFormData:
		return readFormData(v, r)
	case ContextTypeUrlEncoded:
		return readForm(v, r)
	}
	return fmt.Errorf("unexpected content type %q", contentType)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func readJson(v any, r *http.Request) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func readForm(v any, r *http.Request) error {
	values := url.Values{}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Set the values
	for k, v := range r.MultipartForm.Value {
		values[k] = v
	}

	// Read values into the struct
	return ReadQuery(v, values)
}

func readFormData(v any, r *http.Request) error {
	values := url.Values{}

	// Parse the form
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		return err
	}

	// Set the values
	for k, v := range r.MultipartForm.Value {
		values[k] = v
	}

	// Read values into the struct
	if err := ReadQuery(v, values); err != nil {
		return err
	}

	// Read files into the struct
	if err := readFiles(v, r.MultipartForm.File); err != nil {
		return err
	}

	// Return success
	return nil
}
