package httpresponse

import (
	"encoding/json"
	"net/http"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// ErrorResponse is a generic error response which is served as JSON using the
// ServeError method
type responseError struct {
	Code   int    `json:"code"`
	Reason string `json:"message,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ContentTypeKey   = "Content-Type"
	ContentLengthKey = "Content-Length"
	ContentTypeJSON  = "application/json"
	ContentTypeText  = "text/plain"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// JSON is a utility function to serve an arbitary object as JSON
func JSON(w http.ResponseWriter, v interface{}, code int, indent uint) error {
	if w == nil {
		return nil
	}
	w.Header().Set(ContentTypeKey, ContentTypeJSON)
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	if indent > 0 {
		enc.SetIndent("", strings.Repeat(" ", int(indent)))
	}
	return enc.Encode(v)
}

// Text is a utility function to serve plaintext
func Text(w http.ResponseWriter, v string, code int) {
	if w == nil {
		return
	}
	w.Header().Set(ContentTypeKey, ContentTypeText)
	w.WriteHeader(code)
	w.Write([]byte(v + "\n"))
}

// Empty is a utility function to serve an empty response
func Empty(w http.ResponseWriter, code int) {
	if w == nil {
		return
	}
	w.Header().Set(ContentLengthKey, "0")
	w.WriteHeader(code)
}

// Error is a utility function to serve a JSON error notice
func Error(w http.ResponseWriter, code int, reason ...string) error {
	if w == nil {
		return nil
	}
	err := responseError{code, strings.Join(reason, " ")}
	if len(reason) == 0 {
		err.Reason = http.StatusText(int(code))
	}
	return JSON(w, err, code, 0)
}

// Cors is a utility function to set the CORS headers
// on a pre-flight request
func Cors(w http.ResponseWriter, origin string) {
	if w != nil {
		w.Header().Add("Access-Control-Allow-Methods", "*")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Origin", origin)
		w.Header().Set(ContentLengthKey, "0")
		w.WriteHeader(http.StatusOK)
	}
}
