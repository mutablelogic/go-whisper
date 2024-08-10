package store

import (
	"io"
	"net/http"
	"strconv"
)

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

type writer struct {
	io.Writer

	// Current and total bytes
	curBytes, totalBytes uint64

	// Callback function
	fn func(curBytes, totalBytes uint64)
}

// Collect number of bytes written
func (w *writer) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if err == nil && w.fn != nil {
		w.curBytes += uint64(n)
		w.fn(w.curBytes, w.totalBytes)
	}
	return n, nil
}

// Collect total number of bytes
func (w *writer) Header(h http.Header) error {
	if contentLength := h.Get("Content-Length"); contentLength != "" {
		if v, err := strconv.ParseUint(contentLength, 10, 64); err != nil {
			return err
		} else {
			w.totalBytes = v
		}
	}
	return nil
}
