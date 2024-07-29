package api

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"

	// Packages

	"github.com/go-audio/wav"
	"github.com/mutablelogic/go-server/pkg/httprequest"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-whisper/pkg/whisper"
	"github.com/mutablelogic/go-whisper/pkg/whisper/task"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqTranscribe struct {
	File        *multipart.FileHeader `json:"file"`
	Model       string                `json:"model"`
	Language    *string               `json:"language"`
	Prompt      *string               `json:"prompt"`
	ResponseFmt *string               `json:"response_format"`
	Temperature *float32              `json:"temperature"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func TranscribeFile(ctx context.Context, service *whisper.Whisper, w http.ResponseWriter, r *http.Request, translate bool) {
	var req reqTranscribe
	if err := httprequest.Body(&req, r); err != nil {
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
	if err := service.WithModel(model, func(task *task.Context) error {
		// Check model
		if translate && !task.CanTranslate() {
			return ErrBadParameter.With("model is not multilingual, cannot translate")
		}

		// Set parameters for transcription & translation, default to english
		task.SetTranslate(translate)
		if req.Language != nil {
			if err := task.SetLanguage(*req.Language); err != nil {
				return err
			}
		} else if translate {
			if err := task.SetLanguage("en"); err != nil {
				return err
			}
		}
		// TODO Set prompt and temperature
		/*
			if req.Prompt != nil {
				ctx.SetPrompt(*req.Prompt)
			}
			if req.Temperature != nil {
				ctx.SetTemperature(*req.Temperature)
			}
		*/
		// Perform the transcription, return any errors
		return task.Transcribe(ctx, buf.AsFloat32Buffer().Data)
	}); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Response - TODO srt, vtt, verbose_json
	switch req.ResponseFormat() {
	case "text":
		httpresponse.Text(w, result.Text, http.StatusOK)
	default:
		httpresponse.JSON(w, result, http.StatusOK, 2)
	}
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
	if r.ResponseFmt != nil {
		switch *r.ResponseFmt {
		case "json", "text", "srt", "verbose_json", "vtt":
			break
		default:
			return fmt.Errorf("response_format must be one of: json, text, srt, verbose_json, vtt")
		}
	}
	return nil
}
func (r reqTranscribe) ResponseFormat() string {
	if r.ResponseFmt == nil {
		return "json"
	}
	return *r.ResponseFmt
}
