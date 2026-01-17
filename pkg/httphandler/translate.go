package httphandler

import (
	"net/http"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pkg "github.com/mutablelogic/go-whisper/pkg"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RegisterTranslateHandlers registers HTTP handlers for Translate operations
func RegisterTranslateHandlers(router *http.ServeMux, prefix string, manager *pkg.Manager, middleware HTTPMiddlewareFuncs) {
	router.HandleFunc(joinPath(prefix, "translate"), middleware.Wrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_ = translateCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}))
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// translateCreate handles POST /translate requests to translate audio to English
func translateCreate(w http.ResponseWriter, r *http.Request, manager *pkg.Manager) error {
	// Read request parameters including audio file using httprequest.Read
	req := new(schema.TranslateMultipartRequest)
	if err := httprequest.Read(r, req); err != nil {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With("failed to read request"), err.Error())
	}

	// Validate audio file is present
	if req.Audio.Body == nil {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With("missing or invalid audio field"))
	}

	// Perform translation
	result, err := manager.Translate(r.Context(), req.Audio.Body, &req.TranslateRequest)
	if err != nil {
		return httpresponse.Error(w, httperr(err))
	}

	// Return the response
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), result)
}
