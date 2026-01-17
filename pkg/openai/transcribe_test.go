package openai_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/multipart"
	"github.com/mutablelogic/go-server/pkg/types"
	openai "github.com/mutablelogic/go-whisper/pkg/openai"
	"github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_Transcribe_001(t *testing.T) {
	assert := assert.New(t)
	client := NewClient(t)
	assert.NotNil(client)

	// Open sample file
	f, err := os.Open(filepath.Join("../../samples/jfk.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	// Perform transcription
	resp, err := client.Transcribe(context.Background(), openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			File:   multipart.File{Body: f},
			Format: types.StringPtr(openai.FormatJson),
		},
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

	f, err := os.Open(filepath.Join("../../samples/de-podcast.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	// Perform transcription
	resp, err := client.Transcribe(context.Background(), openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			File:   multipart.File{Body: f},
			Format: types.StringPtr(openai.FormatText),
		},
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call transcribe endpoint")
	}

	t.Log(resp)

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

	// Perform transcription
	resp, err := client.Transcribe(context.Background(), openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			File:   multipart.File{Body: f},
			Format: types.StringPtr(openai.FormatSrt),
		},
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call transcribe endpoint")
	}

	t.Log(resp)

}

func Test_Transcribe_004(t *testing.T) {
	assert := assert.New(t)
	client := NewClient(t)
	assert.NotNil(client)

	f, err := os.Open(filepath.Join("../../samples/de-podcast.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	// Perform transcription
	resp, err := client.Transcribe(context.Background(), openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			File:   multipart.File{Body: f},
			Format: types.StringPtr(openai.FormatVtt),
		},
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call transcribe endpoint")
	}

	t.Log(resp)

}

///////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func NewClient(t *testing.T) *openai.Client {
	apikey := os.ExpandEnv("${OPENAI_API_KEY}")
	if apikey == "" {
		t.Skip("skipping test, OPENAI_API_KEY environment variable not set")
	}
	client, err := openai.New(apikey, client.OptTrace(os.Stderr, true))
	if err != nil {
		t.Fatalf("failed to create OpenAI client: %v", err)
	}
	return client
}
