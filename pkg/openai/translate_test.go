package openai_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	// Packages
	"github.com/mutablelogic/go-client/pkg/multipart"
	"github.com/mutablelogic/go-server/pkg/types"
	openai "github.com/mutablelogic/go-whisper/pkg/openai"
	"github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_Translate_001(t *testing.T) {
	assert := assert.New(t)
	client := NewClient(t)
	assert.NotNil(client)

	f, err := os.Open(filepath.Join("../../samples/de-podcast.wav"))
	if !assert.NoError(err) {
		assert.FailNow("failed to open sample file")
	}
	defer f.Close()

	// Perform translation into english
	resp, err := client.Translate(context.Background(), openai.TranslationRequest{
		File:   multipart.File{Body: f},
		Format: types.StringPtr(openai.FormatJson),
	})
	if !assert.NoError(err) {
		assert.FailNow("failed to call translate endpoint")
	}

	t.Log(resp)

}
