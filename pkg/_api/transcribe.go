package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	// Packages
	segmenterPkg "github.com/mutablelogic/go-media/pkg/segmenter"
	"github.com/mutablelogic/go-server/pkg/httprequest"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/mutablelogic/go-whisper"
	"github.com/mutablelogic/go-whisper/pkg/client/gowhisper"
	"github.com/mutablelogic/go-whisper/pkg/client/openai"
	"github.com/mutablelogic/go-whisper/pkg/schema"
	whisperPkg "github.com/mutablelogic/go-whisper/pkg/whisper"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func TranscribeFile(ctx context.Context, service *whisper.Whisper, w http.ResponseWriter, r *http.Request) error {
	var req gowhisper.TranscriptionRequest
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, httpresponse.ErrBadRequest, err.Error())
	}
	return transcribe_file(ctx, service, w, req.File.Body, req.Model, types.PtrString(req.Format), types.PtrString(req.Language), types.PtrString(req.Prompt), req.Temperature, false, types.PtrBool(req.Diarize), types.PtrBool(req.Stream))
}

func TranslateFile(ctx context.Context, service *whisper.Whisper, w http.ResponseWriter, r *http.Request) error {
	var req gowhisper.TranslationRequest
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, httpresponse.ErrBadRequest, err.Error())
	}
	return transcribe_file(ctx, service, w, req.File.Body, req.Model, types.PtrString(req.Format), types.PtrString(req.Language), types.PtrString(req.Prompt), req.Temperature, true, types.PtrBool(req.Diarize), types.PtrBool(req.Stream))
}

func transcribe_file(ctx context.Context, service *whisper.Whisper, w http.ResponseWriter, r io.Reader, model, format, language, prompt string, temperature *float64, translate, diarize, realtime bool) error {
	// Create a text stream
	var stream *httpresponse.TextStream
	if realtime {
		if stream = httpresponse.NewTextStream(w); stream == nil {
			return httpresponse.Error(w, httpresponse.ErrInternalError.With("Cannot create text stream"))
		}
		defer stream.Close()
	}

	// Get the model
	model_ := service.GetModelById(model)
	if model_ == nil {
		err := httpresponse.ErrNotFound.Withf("Model not found: %q", model)
		if stream != nil {
			stream.Write(schema.TranscribeStreamErrorType, schema.Event{
				Type: schema.TranscribeStreamErrorType,
				Text: err.Error(),
			})
			return nil
		} else {
			return httpresponse.Error(w, err)
		}
	}

	// Check the format
	if format = strings.TrimSpace(format); format == "" {
		format = openai.Formats[0] // Default to first format
	} else if !slices.Contains(openai.Formats, format) {
		err := httpresponse.ErrBadRequest.Withf("Unsupported format: %q", format)
		if stream != nil {
			stream.Write(schema.TranscribeStreamErrorType, schema.Event{
				Type: schema.TranscribeStreamErrorType,
				Text: err.Error(),
			})
			return nil
		} else {
			return httpresponse.Error(w, err)
		}
	}

	// Start a translation task
	var result *schema.Transcription
	if err := service.WithModel(model_, func(task *whisperPkg.Task) error {
		task.SetTranslate(translate)
		task.SetDiarize(diarize)

		// Set language
		if language != "" {
			if err := task.SetLanguage(language); err != nil {
				return err
			}
		}

		// Set temperature
		if temperature != nil {
			if err := task.SetTemperature(types.PtrFloat64(temperature)); err != nil {
				return err
			}
		}

		// Set prompt
		if prompt = strings.TrimSpace(prompt); prompt != "" {
			if err := task.SetPrompt(prompt); err != nil {
				return err
			}
		}

		// Set response
		result = task.Result()

		// Decode, resample and segment the audio file
		return segment(ctx, task, r, func(seg *schema.Segment) {
			if stream == nil {
				return
			}

			// If the language has changed, write a language event
			if language != task.Language() {
				language = task.Language()
				stream.Write(schema.TranscribeStreamLanguageType, schema.Event{
					Type: schema.TranscribeStreamLanguageType,
					Text: language,
				})
			}

			// Format the text into the correct format
			var text bytes.Buffer
			switch format {
			case openai.FormatText:
				text.WriteString(seg.Text)
			case openai.FormatSrt:
				seg.WriteSRT(&text, 0)
			case openai.FormatVtt:
				seg.WriteVTT(&text, 0)
			case openai.FormatJson, openai.FormatVerboseJson:
				json.NewEncoder(&text).Encode(seg)
			}

			// Write the segment to the stream
			stream.Write(schema.TranscribeStreamDeltaType, schema.Event{
				Type:  schema.TranscribeStreamDeltaType,
				Delta: text.String(),
			})
		})
	}); err != nil {
		err := httpresponse.ErrInternalError.With(err.Error())
		if stream != nil {
			stream.Write(schema.TranscribeStreamErrorType, schema.Event{
				Type: schema.TranscribeStreamErrorType,
				Text: err.Error(),
			})
			return nil
		} else {
			return httpresponse.Error(w, err)
		}
	}

	// Response to client
	if stream == nil {
		return response(w, format, result)
	} else {
		stream.Write(schema.TranscribeStreamDoneType, schema.Event{
			Type: schema.TranscribeStreamDoneType,
			Text: result.Text,
		})
		return nil
	}
}

func segment(ctx context.Context, task *whisperPkg.Task, r io.Reader, fn func(seg *schema.Segment)) error {
	// Create a segmenter
	segmenter, err := segmenterPkg.NewFromReader(r, whisper.SampleRate)
	if err != nil {
		return err
	}

	// Read segments and perform transcription or  translation
	if err := segmenter.DecodeFloat32(ctx, func(ts time.Duration, buf []float32) error {
		return task.Transcribe(ctx, ts, buf, fn)
	}); err != nil {
		return err
	}

	// Return sucess
	return nil
}

func response(w http.ResponseWriter, format string, response *schema.Transcription) error {
	switch strings.ToLower(format) {
	case openai.FormatJson, openai.FormatVerboseJson:
		return httpresponse.JSON(w, http.StatusOK, 2, response)
	case openai.FormatText, "":
		return httpresponse.Write(w, http.StatusOK, types.ContentTypeTextPlain, func(w io.Writer) (int, error) {
			return w.Write([]byte(response.Text))
		})
	case openai.FormatSrt:
		return httpresponse.Write(w, http.StatusOK, "application/x-subrip", func(w io.Writer) (int, error) {
			for _, seg := range response.Segments {
				whisperPkg.WriteSegmentSrt(w, seg)
			}
			return 0, nil
		})
	case openai.FormatVtt:
		return httpresponse.Write(w, http.StatusOK, "text/vtt", func(w io.Writer) (int, error) {
			if _, err := w.Write([]byte("WEBVTT\n\n")); err != nil {
				return 0, err
			}
			for _, seg := range response.Segments {
				whisperPkg.WriteSegmentVtt(w, seg)
			}
			return 0, nil
		})
	}

	// Error - invalid format
	return httpresponse.ErrBadRequest.Withf("Invalid response format: %q", format)
}
