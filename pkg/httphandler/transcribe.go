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

// RegisterTranscribeHandlers registers HTTP handlers for Transcribe operations
func RegisterTranscribeHandlers(router *http.ServeMux, prefix string, manager *pkg.Manager, middleware HTTPMiddlewareFuncs) {
	router.HandleFunc(joinPath(prefix, "transcribe"), middleware.Wrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_ = transcribeCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}))
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// transcribeCreate handles POST /transcribe requests to transcribe audio
func transcribeCreate(w http.ResponseWriter, r *http.Request, manager *pkg.Manager) error {
	// Read request parameters including audio file using httprequest.Read
	req := new(schema.TranscribeMultipartRequest)
	if err := httprequest.Read(r, req); err != nil {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With("failed to read request"), err.Error())
	}

	// Validate audio file is present
	if req.Audio.Body == nil {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With("missing or invalid audio field"))
	}

	// Perform transcription
	result, err := manager.Transcribe(r.Context(), req.Audio.Body, &req.TranscribeRequest)
	if err != nil {
		return httpresponse.Error(w, httperr(err))
	}

	// Return the response
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), result)
}
