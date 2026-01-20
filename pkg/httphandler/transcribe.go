package httphandler

import (
	"net/http"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
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

	// Create text stream if requested
	var stream *httpresponse.TextStream
	mimetype, err := types.ParseContentType(r.Header.Get(types.ContentAcceptHeader))
	if err != nil {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With("invalid Accept header"), err.Error())
	} else if mimetype == types.ContentTypeTextStream {
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

	// Perform transcription
	result, err := manager.Transcribe(r.Context(), segmentWriter, req.Audio.Body, &req.TranscribeRequest)
	if err != nil {
		if stream != nil {
			stream.Write(schema.TranscribeStreamErrorType, err.Error())
			return nil
		}
		return httpresponse.Error(w, httperr(err))
	}

	// Return response based on Accept header
	if stream != nil {
		// Send summary without segments to avoid exceeding SSE buffer limits.
		// Segments were already streamed via TranscribeStreamDeltaType.
		stream.Write(schema.TranscribeStreamDoneType, result.Summary())
		return nil
	}

	return writeTranscriptionResponse(w, r, mimetype, result)
}

const (
	ContentTypeSRT  = "application/x-subrip"
	ContentTypeSRT1 = "text/srt"
	ContentTypeSRT2 = "text/subrip"
	ContentTypeVTT  = "text/vtt"
	ContentTypeVTT1 = "application/vtt"
)

// writeTranscriptionResponse writes transcription result in the requested format
func writeTranscriptionResponse(w http.ResponseWriter, r *http.Request, mimetype string, result *schema.Transcription) error {
	switch mimetype {
	case types.ContentTypeTextPlain:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		for _, seg := range result.Segments {
			if seg != nil {
				seg.WriteText(w)
			}
		}
		schema.WriteTextTrailer(w)
		return nil
	case ContentTypeSRT1, ContentTypeSRT2, ContentTypeSRT:
		w.Header().Set("Content-Type", ContentTypeSRT)
		w.WriteHeader(http.StatusOK)
		for _, seg := range result.Segments {
			seg.WriteSRT(w, 0)
		}
		return nil
	case ContentTypeVTT, ContentTypeVTT1:
		w.Header().Set("Content-Type", ContentTypeVTT)
		w.WriteHeader(http.StatusOK)
		for _, seg := range result.Segments {
			seg.WriteVTT(w, 0)
		}
		return nil
	default:
		// Default to JSON
		return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), result)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE TYPES

// streamSegmentWriter implements schema.SegmentWriter by emitting SSE events
type streamSegmentWriter struct {
	stream *httpresponse.TextStream
}

// Write emits a segment event to the text stream
func (w *streamSegmentWriter) Write(seg *schema.Segment) {
	w.stream.Write(schema.TranscribeStreamDeltaType, seg)
}
