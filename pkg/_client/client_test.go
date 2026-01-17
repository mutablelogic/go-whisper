package client_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	// Packages
	"github.com/mutablelogic/go-whisper/pkg/client"
	"github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_Transcribe_001(t *testing.T) {
	assert := assert.New(t)
	client, err := client.New()
	if !assert.NoError(err) {
		assert.FailNow("failed to create client")
	}
	assert.NotNil(client)

	models, err := client.ListModels(context.Background())
	if !assert.NoError(err) {
		assert.FailNow("failed to list models")
	}
	if !assert.NotEmpty(models) {
		t.Skip("no models available")
	}

	for _, model := range models {
		t.Run(model.Id, func(t *testing.T) {
			// Open sample file
			f, err := os.Open(filepath.Join("../../samples/jfk.wav"))
			if !assert.NoError(err) {
				assert.FailNow("failed to open sample file")
			}
			defer f.Close()

			_, err = client.Transcribe(context.Background(), model.Id, f)
			if !assert.NoError(err) {
				assert.FailNow("failed to transcribe audio")
			}
		})
	}
}
