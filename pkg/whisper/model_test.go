package whisper_test

import (
	"testing"

	"github.com/mutablelogic/go-whisper/pkg/whisper"
	"github.com/stretchr/testify/assert"
)

func Test_model_001(t *testing.T) {
	assert := assert.New(t)
	model, err := whisper.New("path")
	if !assert.NoError(err) {
		t.SkipNow()
	}
	t.Log(model)
}
