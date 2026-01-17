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

	// Return response based on Accept header
	return writeTranscriptionResponse(w, r, result)
}

// writeTranscriptionResponse writes transcription result in the requested format
func writeTranscriptionResponse(w http.ResponseWriter, r *http.Request, result *schema.Transcription) error {
	acceptHeader := r.Header.Get("Accept")

	// Determine response format based on Accept header
	switch {
	case acceptHeader == "text/plain":
		// Return plain text with segments (includes speaker labels if available)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		for _, seg := range result.Segments {
			if seg != nil {
				seg.WriteText(w)
			}
		}
		w.Write([]byte("\n")) // Final newline
		return nil
	case acceptHeader == "application/x-subrip" || acceptHeader == "text/subrip" || acceptHeader == "text/srt":
		// Return SRT format
		w.Header().Set("Content-Type", "application/x-subrip")
		w.WriteHeader(http.StatusOK)
		for _, seg := range result.Segments {
			if seg != nil {
				seg.WriteSRT(w, 0)
			}
		}
		return nil
	case acceptHeader == "text/vtt" || acceptHeader == "application/vtt":
		// Return VTT format
		w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("WEBVTT\n\n"))
		for _, seg := range result.Segments {
			if seg != nil {
				seg.WriteVTT(w, 0)
			}
		}
		return nil
	default:
		// Default to JSON
		return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), result)
	}
}
