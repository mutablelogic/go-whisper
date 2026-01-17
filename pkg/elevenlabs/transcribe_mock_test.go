package elevenlabs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Packages
	goclient "github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/multipart"
)

func Test_GetTranscript_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/speech-to-text/transcripts/abc123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("xi-api-key"); got == "" {
			t.Fatalf("missing xi-api-key header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"language_code":"en","language_probability":0.9,"text":"hello"}`))
	}))
	t.Cleanup(server.Close)

	client, err := New("dummy", goclient.OptEndpoint(server.URL))
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	resp, err := client.GetTranscript(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("GetTranscript error: %v", err)
	}
	if resp.Text != "hello" {
		t.Fatalf("unexpected text: %q", resp.Text)
	}
}

func Test_DeleteTranscript_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/speech-to-text/transcripts/abc123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	client, err := New("dummy", goclient.OptEndpoint(server.URL))
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	if err := client.DeleteTranscript(context.Background(), "abc123"); err != nil {
		t.Fatalf("DeleteTranscript error: %v", err)
	}
}

func Test_Transcribe_RequestFields_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/speech-to-text" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if err := r.ParseMultipartForm(4 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		checkField := func(name, want string) {
			got := r.FormValue(name)
			if got != want {
				t.Fatalf("%s: got %q want %q", name, got, want)
			}
		}
		checkField("model_id", "scribe_v2")
		checkField("language_code", "en")
		checkField("tag_audio_events", "false")
		checkField("num_speakers", "3")
		checkField("timestamps_granularity", "word")
		checkField("diarize", "true")
		checkField("diarization_threshold", "0.22")
		checkField("file_format", "other")
		checkField("cloud_storage_url", "https://example.com/file.wav")
		checkField("enable_logging", "false")
		checkField("webhook", "true")
		checkField("webhook_id", "hook-1")
		checkField("temperature", "0.5")
		checkField("seed", "42")
		checkField("use_multi_channel", "true")
		checkField("entity_detection", "all")
		if got := r.FormValue("keyterms"); got != "foo" {
			t.Fatalf("keyterms: got %q", got)
		}
		if got := r.FormValue("additional_formats"); got == "" {
			t.Fatalf("additional_formats missing")
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("file: %v", err)
		}
		_ = file.Close()
		if !strings.HasPrefix(header.Filename, "sample") {
			t.Fatalf("file name unexpected: %s", header.Filename)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"language_code":"en","language_probability":0.9,"text":"ok"}`))
	}))
	t.Cleanup(server.Close)

	client, err := New("dummy", goclient.OptEndpoint(server.URL))
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	lang := "en"
	model := "scribe_v2"
	tags := false
	diarize := true
	dthr := 0.22
	fileFmt := "other"
	cloud := "https://example.com/file.wav"
	enableLog := false
	webhook := true
	whID := "hook-1"
	temp := 0.5
	seed := int64(42)
	multi := true
	ent := "all"
	keyterms := []string{"foo"}
	addFmt := []map[string]any{{"format": "srt"}}

	body := strings.NewReader("hello")

	_, err = client.Transcribe(context.Background(), TranscribeRequest{
		Model:          model,
		File:           multipart.File{Body: body, Path: "sample.wav"},
		Language:       &lang,
		TagAudioEvents: &tags,
		NumSpeakers:    uint64Ptr(3),
		Timestamps:     stringPtr("word"),
		Diarize:        &diarize,
		DiarizationThr: &dthr,
		FileFormat:     &fileFmt,
		CloudURL:       &cloud,
		EnableLogging:  &enableLog,
		Webhook:        &webhook,
		WebhookID:      &whID,
		WebhookMeta:    jsonRaw(`{"id":1}`),
		Temperature:    &temp,
		Seed:           &seed,
		UseMultiChan:   &multi,
		EntityDetect:   ent,
		Keyterms:       keyterms,
		AdditionalFmt:  addFmt,
	})
	if err != nil {
		t.Fatalf("Transcribe error: %v", err)
	}
}

func uint64Ptr(v uint64) *uint64       { return &v }
func stringPtr(v string) *string       { return &v }
func jsonRaw(s string) json.RawMessage { return json.RawMessage(s) }
