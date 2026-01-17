package httphandler

import (
	"errors"
	"net/http"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
	pkg "github.com/mutablelogic/go-whisper/pkg"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type HTTPMiddlewareFuncs []func(http.HandlerFunc) http.HandlerFunc

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RegisterHandlers registers all whisper HTTP handlers on the provided
// router with the given path prefix. The manager must be non-nil.
func RegisterHandlers(router *http.ServeMux, prefix string, manager *pkg.Manager, middleware HTTPMiddlewareFuncs) {
	RegisterModelHandlers(router, prefix, manager, middleware)
	RegisterTranscribeHandlers(router, prefix, manager, middleware)
	RegisterTranslateHandlers(router, prefix, manager, middleware)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (w HTTPMiddlewareFuncs) Wrap(handler http.HandlerFunc) http.HandlerFunc {
	if len(w) == 0 {
		return handler
	}
	for i := len(w) - 1; i >= 0; i-- {
		handler = w[i](handler)
	}
	return handler
}

func joinPath(prefix, path string) string {
	return types.JoinPath(prefix, path)
}

// httperr converts pkg errors to appropriate HTTP errors.
// Returns the original error if it's already an httpresponse.Err,
// otherwise maps pkg errors to their HTTP equivalents.
func httperr(err error) error {
	if err == nil {
		return nil
	}

	// If already an HTTP error, return as-is
	var httpErr httpresponse.Err
	if errors.As(err, &httpErr) {
		return err
	}

	// Map pkg errors to HTTP errors
	switch {
	case errors.Is(err, ErrNotFound):
		return httpresponse.ErrNotFound.With(err.Error())
	case errors.Is(err, ErrBadParameter):
		return httpresponse.ErrBadRequest.With(err.Error())
	default:
		return httpresponse.ErrInternalError.With(err.Error())
	}
}
