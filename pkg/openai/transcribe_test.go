package openai_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/multipart"
	"github.com/mutablelogic/go-server/pkg/types"
	openai "github.com/mutablelogic/go-whisper/pkg/openai"
	"github.com/mutablelogic/go-whisper/pkg/schema"
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

func Test_Transcribe_Diarization(t *testing.T) {
	assert := assert.New(t)
	// Create client with 2-minute timeout for diarization (it can take a while)
	client := NewClientWithTimeout(t, 2*time.Minute)
	assert.NotNil(client)

	// Open multi-language sample file (has multiple speakers)
	f, err := os.Open(filepath.Join("../../samples/multi-lang.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	// Create context with 2-minute timeout for diarization (it can take a while)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Perform transcription with diarization
	resp, err := client.Transcribe(ctx, openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			Model:  "gpt-4o-transcribe-diarize",
			File:   multipart.File{Body: f},
			Format: types.StringPtr(openai.FormatDiarizedJson),
		},
		ChunkingStrategy: &openai.ChunkingStrategy{
			Type: openai.ChunkingStrategyServerVAD,
		},
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call transcribe endpoint with diarization")
	}

	// Verify response
	assert.NotEmpty(resp.Text)
	t.Logf("Transcription: %s", resp.Text)

	// Check segments have speaker labels
	if len(resp.Segment) > 0 {
		for i, seg := range resp.Segment {
			t.Logf("Segment %d: Speaker=%q, Text=%q, Start=%v, End=%v",
				i, seg.Speaker, seg.Text, seg.Start, seg.End)
		}
	}
}

func Test_Transcribe_ChunkingStrategyAuto(t *testing.T) {
	assert := assert.New(t)
	client := NewClientWithTimeout(t, 2*time.Minute)
	assert.NotNil(client)

	// Open sample file
	f, err := os.Open(filepath.Join("../../samples/jfk.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Perform transcription with auto chunking strategy
	resp, err := client.Transcribe(ctx, openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			Model:  "gpt-4o-transcribe",
			File:   multipart.File{Body: f},
			Format: types.StringPtr(openai.FormatJson),
		},
		ChunkingStrategy: &openai.ChunkingStrategy{
			Type: openai.ChunkingStrategyAuto,
		},
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call transcribe endpoint with auto chunking")
	}

	assert.NotEmpty(resp.Text)
	t.Logf("Transcription: %s", resp.Text)
}

func Test_Transcribe_Streaming(t *testing.T) {
	assert := assert.New(t)
	client := NewClientWithTimeout(t, 2*time.Minute)
	assert.NotNil(client)

	f, err := os.Open(filepath.Join("../../samples/jfk.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Track streaming events
	var events []schema.Event
	client.SetStreamCallback(func(evt schema.Event) {
		t.Logf("Stream event: type=%s delta=%q text=%q", evt.Type, evt.Delta, evt.Text)
		events = append(events, evt)
	})

	// Perform streaming transcription
	_, err = client.Transcribe(ctx, openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			Model:  "gpt-4o-transcribe",
			File:   multipart.File{Body: f},
			Format: types.StringPtr(openai.FormatText),
		},
		Stream: types.BoolPtr(true),
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call streaming transcribe endpoint")
	}

	// Verify we received streaming events
	assert.NotEmpty(events, "should have received streaming events")
	t.Logf("Received %d streaming events", len(events))

	// Check for expected event types
	var hasDelta, hasDone bool
	var doneText string
	for _, evt := range events {
		if evt.Type == schema.TranscribeStreamDeltaType {
			hasDelta = true
		}
		if evt.Type == schema.TranscribeStreamDoneType {
			hasDone = true
			doneText = evt.Text
		}
	}
	assert.True(hasDelta, "should have received delta events")
	assert.True(hasDone, "should have received done event")
	assert.NotEmpty(doneText, "done event should contain complete text")
	t.Logf("Text from done event: %s", doneText)
}

func Test_Transcribe_Streaming_Json(t *testing.T) {
	assert := assert.New(t)
	client := NewClientWithTimeout(t, 2*time.Minute)
	assert.NotNil(client)

	f, err := os.Open(filepath.Join("../../samples/jfk.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Track streaming events
	var events []schema.Event
	client.SetStreamCallback(func(evt schema.Event) {
		t.Logf("Stream event: type=%s delta=%q text=%q", evt.Type, evt.Delta, evt.Text)
		events = append(events, evt)
	})

	// Perform streaming transcription with json format
	_, err = client.Transcribe(ctx, openai.TranscriptionRequest{
		TranslationRequest: openai.TranslationRequest{
			Model:  "gpt-4o-transcribe",
			File:   multipart.File{Body: f},
			Format: types.StringPtr(openai.FormatJson),
		},
		Stream: types.BoolPtr(true),
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call streaming transcribe endpoint")
	}

	// Verify we received streaming events
	assert.NotEmpty(events, "should have received streaming events")
	t.Logf("Received %d streaming events", len(events))

	// Check for expected event types
	var hasDelta, hasDone bool
	var doneText string
	for _, evt := range events {
		if evt.Type == schema.TranscribeStreamDeltaType {
			hasDelta = true
		}
		if evt.Type == schema.TranscribeStreamDoneType {
			hasDone = true
			doneText = evt.Text
		}
	}
	assert.True(hasDelta, "should have received delta events")
	assert.True(hasDone, "should have received done event")
	assert.NotEmpty(doneText, "done event should contain complete text")
	t.Logf("Text from done event: %s", doneText)

	// Note: When streaming, resp.Text may be empty since body was consumed by stream callback
	// The complete text is available in the transcript.text.done event
}

///////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func NewClient(t *testing.T) *openai.Client {
	return NewClientWithTimeout(t, 0)
}

func NewClientWithTimeout(t *testing.T, timeout time.Duration) *openai.Client {
	apikey := os.ExpandEnv("${OPENAI_API_KEY}")
	if apikey == "" {
		t.Skip("skipping test, OPENAI_API_KEY environment variable not set")
	}
	opts := []client.ClientOpt{client.OptTrace(os.Stderr, true)}
	if timeout > 0 {
		opts = append(opts, client.OptTimeout(timeout))
	}
	c, err := openai.New(apikey, opts...)
	if err != nil {
		t.Fatalf("failed to create OpenAI client: %v", err)
	}
	return c
}
