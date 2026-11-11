package httpclient_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Packages
	httpclient "github.com/mutablelogic/go-whisper/pkg/httpclient"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func TestTranslate_Language(t *testing.T) {
	// A server which records the form of a request to translate
	var path string
	var form map[string][]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		require.NoError(t, r.ParseMultipartForm(1<<20))
		form = r.MultipartForm.Value
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(schema.Transcription{Text: "Good morning"}))
	}))
	t.Cleanup(server.Close)
	client, err := httpclient.New(server.URL)
	require.NoError(t, err)

	// The language the audio is in is sent, to translate from
	result, err := client.Translate(t.Context(), "ggml-small", strings.NewReader("audio"), httpclient.WithLanguage("de"))
	require.NoError(t, err)
	assert.Equal(t, "Good morning", result.Text)
	assert.Equal(t, "/translate", path)
	assert.Equal(t, []string{"ggml-small"}, form["model"])
	assert.Equal(t, []string{"de"}, form["language"])

	// And without it, it isn't, so it's detected
	_, err = client.Translate(t.Context(), "ggml-small", strings.NewReader("audio"))
	require.NoError(t, err)
	assert.NotContains(t, form, "language")
}
