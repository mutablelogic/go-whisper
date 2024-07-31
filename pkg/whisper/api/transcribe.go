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

type queryTranscribe struct {
	Stream bool `json:"stream"`
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
	var query queryTranscribe
	if err := httprequest.Query(&query, r.URL.Query()); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, err.Error())
		return
	}
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

	// Create a text stream
	var stream *httpresponse.TextStream
	if query.Stream {
		if stream = httpresponse.NewTextStream(w); stream == nil {
			httpresponse.Error(w, http.StatusInternalServerError, "Cannot create text stream")
			return
		}
		defer stream.Close()
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

		// Create response
		result = task.Result()
		result.Task = "transcribe"
		if translate {
			result.Task = "translate"
		}
		result.Duration = schema.Timestamp(segmenter.Duration())
		result.Language = task.Language()

		// Output the header
		if stream != nil {
			stream.Write("task", result)
		}

		// Read samples and transcribe them
		if err := segmenter.Decode(ctx, func(ts time.Duration, buf []float32) error {
			// Perform the transcription, return any errors
			return task.Transcribe(ctx, ts, buf, req.OutputSegments() || stream != nil, func(segment *schema.Segment) {
				if stream != nil {
					stream.Write("segment", segment)
				}
			})
		}); err != nil {
			return err
		}

		// Set the language
		result.Language = task.Language()

		// Return success
		return nil
	}); err != nil {
		if stream != nil {
			stream.Write("error", err.Error())
		} else {
			httpresponse.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Return transcription if not streaming
	if stream == nil {
		httpresponse.JSON(w, result, http.StatusOK, 2)
	} else {
		stream.Write("ok")
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
