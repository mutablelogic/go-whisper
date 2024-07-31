package api

import (
	"net/http"
	"path/filepath"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-whisper/pkg/whisper"
)

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterEndpoints(base string, mux *http.ServeMux, whisper *whisper.Whisper) {
	// Health: GET /v1/health
	//   returns an empty OK response
	mux.HandleFunc(joinPath(base, "health"), wrapLogging(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodGet:
			httpresponse.Empty(w, http.StatusOK)
		default:
			httpresponse.Error(w, http.StatusMethodNotAllowed)
		}
	}))

	// List Models: GET /v1/models
	//   returns available models
	// Download Model: POST /v1/models?stream={bool}
	//   downloads a model from the server
	//   if stream is true then progress is streamed back to the client
	mux.HandleFunc(joinPath(base, "models"), wrapLogging(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodGet:
			ListModels(r.Context(), w, whisper)
		case http.MethodPost:
			DownloadModel(r.Context(), w, r, whisper)
		default:
			httpresponse.Error(w, http.StatusMethodNotAllowed)
		}
	}))

	// Get: GET /v1/models/{id}
	//   returns an existing model
	// Delete: DELETE /v1/models/{id}
	//   deletes an existing model
	mux.HandleFunc(joinPath(base, "models/{id}"), wrapLogging(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		id := r.PathValue("id")
		switch r.Method {
		case http.MethodGet:
			GetModelById(r.Context(), w, whisper, id)
		case http.MethodDelete:
			DeleteModelById(r.Context(), w, whisper, id)
		default:
			httpresponse.Error(w, http.StatusMethodNotAllowed)
		}
	}))

	// Translate: POST /v1/audio/translations
	//   Translates audio into english or another language  - language parameter should be set to the
	//   destination language of the audio. Will default to english if not set.
	mux.HandleFunc(joinPath(base, "audio/translations"), wrapLogging(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodPost:
			TranscribeFile(r.Context(), whisper, w, r, true)
		default:
			httpresponse.Error(w, http.StatusMethodNotAllowed)
		}
	}))

	// Transcribe: POST /v1/audio/transcriptions
	//   Transcribes audio into the input language - language parameter should be set to the source
	//   language of the audio
	mux.HandleFunc(joinPath(base, "audio/transcriptions"), wrapLogging(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodPost:
			TranscribeFile(r.Context(), whisper, w, r, false)
		default:
			httpresponse.Error(w, http.StatusMethodNotAllowed)
		}
	}))

	// Transcribe: POST /v1/audio/transcriptions/{model-id}
	//   Transcribes streamed media into the input language
	mux.HandleFunc(joinPath(base, "audio/transcriptions/{model}"), wrapLogging(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		model := r.PathValue("model")
		switch r.Method {
		case http.MethodPost:
			TranscribeStream(r.Context(), whisper, w, r, model)
		default:
			httpresponse.Error(w, http.StatusMethodNotAllowed)
		}
	}))
}

/////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func joinPath(base, rel string) string {
	return filepath.Join(base, rel)
}
