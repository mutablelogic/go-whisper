package httphandler

import (
	"net/http"
	"strings"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
	pkg "github.com/mutablelogic/go-whisper/pkg"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RegisterModelHandlers registers HTTP handlers for Model operations
func RegisterModelHandlers(router *http.ServeMux, prefix string, manager *pkg.Manager, middleware HTTPMiddlewareFuncs) {
	// GET /model - list all models
	// POST /model - download a model
	router.HandleFunc(joinPath(prefix, "model"), middleware.Wrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = modelList(w, r, manager)
		case http.MethodPost:
			_ = modelDownload(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}))

	// GET /model/{id} - get a specific model
	// DELETE /model/{id} - delete a specific model
	router.HandleFunc(joinPath(prefix, "model/{id}"), middleware.Wrap(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = modelGet(w, r, manager)
		case http.MethodDelete:
			_ = modelDelete(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}))
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// modelList handles GET /model requests to list all available models
func modelList(w http.ResponseWriter, r *http.Request, manager *pkg.Manager) error {
	// Get all models
	models := manager.ListModels(r.Context())

	// Return the response
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), models)
}

// modelGet handles GET /model/{id} requests to retrieve a specific model
func modelGet(w http.ResponseWriter, r *http.Request, manager *pkg.Manager) error {
	// Extract model ID from path
	id := r.PathValue("id")
	if id == "" {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With("model id is required"))
	}

	// Get the model
	model, err := manager.GetModel(r.Context(), id)
	if err != nil {
		return httpresponse.Error(w, httperr(err))
	}

	// Return the response
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), model)
}

// modelDownload handles POST /model requests to download a specific model
func modelDownload(w http.ResponseWriter, r *http.Request, manager *pkg.Manager) error {
	// Parse request body
	var req schema.DownloadModelRequest
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With(err.Error()))
	}

	// Validate model ID is provided
	if req.Model == "" {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With("model is required"))
	}

	// Check if streaming is requested via Accept header
	acceptHeader := r.Header.Get("Accept")
	wantsStream := strings.Contains(acceptHeader, types.ContentTypeTextStream)

	// Create text stream if requested
	var stream *httpresponse.TextStream
	if wantsStream {
		stream = httpresponse.NewTextStream(w)
		if stream == nil {
			return httpresponse.Error(w, httpresponse.ErrInternalError.With("cannot create text stream"))
		}
		defer stream.Close()
	}

	// Progress callback
	var progressFn func(cur, total uint64)
	if stream != nil {
		progressFn = func(cur, total uint64) {
			progress := map[string]any{
				"current": cur,
				"total":   total,
			}
			if total > 0 {
				progress["percent"] = float64(cur) / float64(total) * 100
			}
			stream.Write("progress", progress)
		}
	}

	// Download the model
	model, err := manager.DownloadModel(r.Context(), req.Model, progressFn)
	if err != nil {
		if stream != nil {
			stream.Write("error", err.Error())
			return nil
		}
		return httpresponse.Error(w, httperr(err))
	}

	// Return done
	if stream != nil {
		stream.Write("done", model)
		return nil
	}

	// Return model
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), model)
}

// modelDelete handles DELETE /model/{id} requests to delete a specific model
func modelDelete(w http.ResponseWriter, r *http.Request, manager *pkg.Manager) error {
	// Extract model ID from path
	id := r.PathValue("id")
	if id == "" {
		return httpresponse.Error(w, httpresponse.ErrBadRequest.With("model id is required"))
	}

	// Delete the model
	if err := manager.DeleteModel(r.Context(), id); err != nil {
		return httpresponse.Error(w, httperr(err))
	}

	// Return success with no content
	return httpresponse.JSON(w, http.StatusNoContent, httprequest.Indent(r), nil)
}
