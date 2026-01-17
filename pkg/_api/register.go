package api

import (
	"net/http"
	"os"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/logger"
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/mutablelogic/go-whisper"
	"github.com/mutablelogic/go-whisper/pkg/client/openai"
)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterEndpoints(base string, whisper *whisper.Whisper, mux *http.ServeMux, debug bool) *http.ServeMux {
	// Create a new router
	if mux == nil {
		mux = http.NewServeMux()
	}

	// Create a logger
	logger := logger.New(os.Stderr, logger.Term, debug)

	// Not Found: GET /
	//   returns a not found response
	mux.HandleFunc("/", logger.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Error(w, httpresponse.ErrNotFound)
	}))

	// Health: GET /v1/health
	//   returns an empty OK response
	mux.HandleFunc(types.JoinPath(base, "health"), logger.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodGet:
			httpresponse.Empty(w, http.StatusOK)
		default:
			httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}))

	// List Models: GET /v1/models
	//   returns available models
	// Download Model: POST /v1/models?stream={bool}
	//   downloads a model from the server
	//   if stream is true then progress is streamed back to the client
	mux.HandleFunc(types.JoinPath(base, "models"), logger.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodGet:
			ListModels(r.Context(), w, whisper)
		case http.MethodPost:
			DownloadModel(r.Context(), w, r, whisper)
		default:
			httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}))

	// Get: GET /v1/models/{id}
	//   returns an existing model
	// Delete: DELETE /v1/models/{id}
	//   deletes an existing model
	mux.HandleFunc(types.JoinPath(base, "models/{id}"), logger.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		id := r.PathValue("id")
		switch r.Method {
		case http.MethodGet:
			GetModelById(r.Context(), w, whisper, id)
		case http.MethodDelete:
			DeleteModelById(r.Context(), w, whisper, id)
		default:
			httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}))

	// Translate: POST /v1/audio/translations
	//   Translates audio into english
	mux.HandleFunc(types.JoinPath(base, openai.TranslatePath), logger.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodPost:
			TranslateFile(r.Context(), whisper, w, r)
		default:
			httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}))

	// Transcribe: POST /v1/audio/transcriptions
	//   Transcribes audio into the input language - language parameter should be set to the source
	//   language of the audio
	mux.HandleFunc(types.JoinPath(base, openai.TranscribePath), logger.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodPost:
			TranscribeFile(r.Context(), whisper, w, r)
		default:
			httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	}))

	// Transcribe: POST /v1/audio/transcriptions/{model-id}
	//   Transcribes streamed media into the input language
	/*
		mux.HandleFunc(joinPath(base, "audio/transcriptions/{model}"), func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()

			model := r.PathValue("model")
			switch r.Method {
			case http.MethodPost:
				TranscribeStream(r.Context(), whisper, w, r, model)
			default:
				httpresponse.Error(w, http.StatusMethodNotAllowed)
			}
		})*/

	// Return mux
	return mux
}
