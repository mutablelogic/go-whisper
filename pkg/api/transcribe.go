package api

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/go-audio/wav"
	"github.com/mutablelogic/go-whisper/pkg/httprequest"
	"github.com/mutablelogic/go-whisper/pkg/httpresponse"
	"github.com/mutablelogic/go-whisper/pkg/whisper"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqTranscribe struct {
	File           *multipart.FileHeader `json:"file"`
	Model          string                `json:"model"`
	Language       *string               `json:"language"`
	Prompt         *string               `json:"prompt"`
	ResponseFormat *string               `json:"response_format"`
	Temperature    *float32              `json:"temperature"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func TranscribeFile(ctx context.Context, service *whisper.Whisper, w http.ResponseWriter, r *http.Request) {
	var req reqTranscribe
	if err := httprequest.ReadBody(&req, r); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the model
	model := service.GetModelById(req.Model)
	if model == nil {
		httpresponse.Error(w, http.StatusNotFound, "model not found")
		return
	}

	// Check audio format - allow WAV or binary
	if req.File.Header.Get("Content-Type") != "audio/wav" && req.File.Header.Get("Content-Type") != httprequest.ContentTypeBinary {
		httpresponse.Error(w, http.StatusBadRequest, "unsupported audio format:", req.File.Header.Get("Content-Type"))
		return
	}

	// Open file
	f, err := req.File.Open()
	if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer f.Close()

	// Read samples
	buf, err := wav.NewDecoder(f).FullPCMBuffer()
	if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get context for the model, perform transcription
	var result *whisper.Transcription
	if err := service.WithModelContext(model, func(ctx *whisper.Context) error {
		var err error

		// Set parameters for transcription
		if req.Language != nil {
			if err := ctx.SetLanguage(*req.Language); err != nil {
				return err
			}
		}
		if req.Prompt != nil {
			ctx.SetPrompt(*req.Prompt)
		}
		if req.Temperature != nil {
			ctx.SetTemperature(*req.Temperature)
		}

		// Perform the transcription, return any errors
		result, err = service.Transcribe(ctx, buf.AsFloat32Buffer().Data)
		return err
	}); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return OK
	httpresponse.JSON(w, result, http.StatusOK, 2)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r reqTranscribe) Validate() error {
	if r.Model == "" {
		return fmt.Errorf("model is required")
	}
	if r.File == nil {
		return fmt.Errorf("file is required")
	}
	return nil
}
