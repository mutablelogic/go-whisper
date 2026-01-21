package httphandler

import (
	"net/http"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
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

	// Create text stream if requested
	var stream *httpresponse.TextStream
	var mimetype string
	if accept := r.Header.Get(types.ContentAcceptHeader); accept != "" {
		var err error
		mimetype, err = types.ParseContentType(accept)
		if err != nil {
			return httpresponse.Error(w, httpresponse.ErrBadRequest.With("invalid Accept header"), err.Error())
		}
	}
	if mimetype == types.ContentTypeTextStream {
		stream = httpresponse.NewTextStream(w)
		if stream == nil {
			return httpresponse.Error(w, httpresponse.ErrInternalError.With("cannot create text stream"))
		}
		defer stream.Close()
	}

	// Create segment writer if streaming
	var segmentWriter schema.SegmentWriter
	if stream != nil {
		segmentWriter = &streamSegmentWriter{stream: stream}
	}

	// Perform translation
	result, err := manager.Translate(r.Context(), segmentWriter, req.Audio.Body, &req.TranslateRequest)
	if err != nil {
		if stream != nil {
			stream.Write(schema.TranscribeStreamErrorType, err.Error())
			return nil
		}
		return httpresponse.Error(w, httperr(err))
	}

	// Return response based on Accept header
	if stream != nil {
		stream.Write(schema.TranscribeStreamDoneType, result.Summary())
		return nil
	}

	return writeTranscriptionResponse(w, r, mimetype, result)
}
