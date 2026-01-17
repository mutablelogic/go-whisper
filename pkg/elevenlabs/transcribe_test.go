package elevenlabs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	// Packages
	multipart "github.com/mutablelogic/go-client/pkg/multipart"
	types "github.com/mutablelogic/go-server/pkg/types"
	elevenlabs "github.com/mutablelogic/go-whisper/pkg/elevenlabs"
	assert "github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_Transcribe_001(t *testing.T) {
	assert := assert.New(t)
	client := NewClient(t)
	assert.NotNil(client)

	f, err := os.Open(filepath.Join("../../samples/jfk.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	resp, err := client.Transcribe(context.Background(), elevenlabs.TranscribeRequest{
		File: multipart.File{Body: f},
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call transcribe endpoint")
	}

	t.Log(resp)

}

func Test_Transcribe_002(t *testing.T) {
	assert := assert.New(t)
	client := NewClient(t)
	assert.NotNil(client)

	f, err := os.Open(filepath.Join("../../samples/en-office.mp3"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	resp, err := client.Transcribe(context.Background(), elevenlabs.TranscribeRequest{
		File:    multipart.File{Body: f},
		Diarize: types.BoolPtr(true),
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call transcribe endpoint")
	}

	t.Log(resp.Segments())

}

func Test_Transcribe_003(t *testing.T) {
	assert := assert.New(t)
	client := NewClient(t)
	assert.NotNil(client)

	f, err := os.Open(filepath.Join("../../samples/de-podcast.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	resp, err := client.Transcribe(context.Background(), elevenlabs.TranscribeRequest{
		File:    multipart.File{Body: f},
		Diarize: types.BoolPtr(false),
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call transcribe endpoint")
	}

	t.Log(resp.Segments())

}

func Test_Transcript_EmptyID(t *testing.T) {
	client := NewClient(t)
	if client == nil {
		t.Skip("client not created")
	}

	if _, err := client.GetTranscript(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty transcription id")
	}
	if err := client.DeleteTranscript(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty transcription id")
	}
}

///////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func NewClient(t *testing.T) *elevenlabs.Client {
	apikey := os.ExpandEnv("${ELEVENLABS_API_KEY}")
	if apikey == "" {
		t.Skip("skipping test, ELEVENLABS_API_KEY environment variable not set")
	}
	client, err := elevenlabs.New(apikey)
	if err != nil {
		t.Fatalf("failed to create ElevenLabs client: %v", err)
	}
	return client
}
