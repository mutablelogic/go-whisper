package api

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"

	"github.com/mutablelogic/go-whisper/pkg/httprequest"
	"github.com/mutablelogic/go-whisper/pkg/httpresponse"
	"github.com/mutablelogic/go-whisper/pkg/whisper"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type respModels struct {
	Object string           `json:"object,omitempty"`
	Models []*whisper.Model `json:"models"`
}

type reqDownloadModel struct {
	Path string `json:"path"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func ListModels(ctx context.Context, w http.ResponseWriter, service *whisper.Whisper) {
	httpresponse.JSON(w, respModels{
		Object: "list",
		Models: service.ListModels(),
	}, http.StatusOK, 2)
}

func DownloadModel(ctx context.Context, w http.ResponseWriter, r *http.Request, service *whisper.Whisper) {
	var req reqDownloadModel
	if err := httprequest.ReadBody(&req, r); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Download the model
	model, err := service.DownloadModel(ctx, req.Name(), req.DestPath())
	if err != nil {
		httpresponse.Error(w, http.StatusBadGateway, err.Error())
		return
	}

	// Return the model information
	httpresponse.JSON(w, model, http.StatusCreated, 2)
}

func GetModelById(ctx context.Context, w http.ResponseWriter, service *whisper.Whisper, id string) {
	model := service.GetModelById(id)
	if model == nil {
		httpresponse.Error(w, http.StatusNotFound)
		return
	}
	httpresponse.JSON(w, model, http.StatusOK, 2)
}

func DeleteModelById(ctx context.Context, w http.ResponseWriter, service *whisper.Whisper, id string) {
	model := service.GetModelById(id)
	if model == nil {
		httpresponse.Error(w, http.StatusNotFound)
		return
	}
	if err := service.DeleteModelById(model.Id); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpresponse.Empty(w, http.StatusOK)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Validate the request
func (r reqDownloadModel) Validate() error {
	if r.Path == "" {
		return errors.New("missing path")
	}
	return nil
}

// Return the model name
func (r reqDownloadModel) Name() string {
	return filepath.Base(r.Path)
}

// Return the model path
func (r reqDownloadModel) DestPath() string {
	return filepath.Dir(r.Path)
}
