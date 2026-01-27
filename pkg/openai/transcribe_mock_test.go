package openai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Packages
	goclient "github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/multipart"
)

func Test_Transcribe_TimestampGranularities_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/audio/transcriptions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(2 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		vals := r.MultipartForm.Value["timestamp_granularities"]
		if len(vals) != 2 || vals[0] != "word" || vals[1] != "segment" {
			t.Fatalf("timestamp_granularities: %+v", vals)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"ok"}`))
	}))
	t.Cleanup(server.Close)

	client, err := New("dummy", goclient.OptEndpoint(server.URL+"/"))
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	body := strings.NewReader("audio")
	_, err = client.Transcribe(context.Background(), TranscriptionRequest{
		TranslationRequest: TranslationRequest{
			File: multipart.File{Body: body, Path: "sample.wav"},
		},
		Timestamps: []string{"word", "segment"},
	}, nil)
	if err != nil {
		t.Fatalf("Transcribe error: %v", err)
	}
}
