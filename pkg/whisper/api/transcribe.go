package api

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httprequest"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-whisper/pkg/whisper"
	"github.com/mutablelogic/go-whisper/pkg/whisper/schema"
	"github.com/mutablelogic/go-whisper/pkg/whisper/segmenter"
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
	Temperature *float32              `json:"temperature"`
	SegmentSize *time.Duration        `json:"segment_size"`
	ResponseFmt *string               `json:"response_format"`
}

const (
	minSegmentSize     = 5 * time.Second
	maxSegmentSize     = 10 * time.Minute
	defaultSegmentSize = 5 * time.Minute
)

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

	// Open file
	f, err := req.File.Open()
	if err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer f.Close()

	// Create a segmenter - read segments based on requested segment size
	segmenter, err := segmenter.New(f, req.SegmentDur(), whisper.SampleRate)
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get context for the model, perform transcription
	var result *schema.Transcription
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

		// TODO: Set temperature, etc

		// Read samples and transcribe them
		if err := segmenter.Decode(ctx, func(ts time.Duration, buf []float32) error {
			// Perform the transcription, return any errors
			return task.Transcribe(ctx, ts, buf, req.OutputSegments(), func(segment *schema.Segment) {
				fmt.Println("TODO: ", segment)
			})
		}); err != nil {
			return err
		}

		// End of transcription, get result
		result = task.Result()

		// Return success
		return nil
	}); err != nil {
		httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set task, duration
	result.Task = "transcribe"
	if translate {
		result.Task = "translate"
	}
	result.Duration = segmenter.Duration()

	// Return transcription
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

func (r reqTranscribe) OutputSegments() bool {
	// We want to output segments if the response format is  "srt", "verbose_json", "vtt"
	switch r.ResponseFormat() {
	case "srt", "verbose_json", "vtt":
		return true
	default:
		return false
	}
}

func (r reqTranscribe) SegmentDur() time.Duration {
	if r.SegmentSize == nil {
		return defaultSegmentSize
	}
	if *r.SegmentSize < minSegmentSize {
		return minSegmentSize
	}
	if *r.SegmentSize > maxSegmentSize {
		return maxSegmentSize
	}
	return *r.SegmentSize
}
