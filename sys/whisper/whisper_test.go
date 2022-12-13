package whisper_test

import (
	"testing"

	// Packages
	whisper "github.com/djthorpe/go-whisper/sys/whisper"
	assert "github.com/stretchr/testify/assert"
)

const (
	MODEL = "../../models/ggml-tiny.bin"
)

func Test_Whisper_000(t *testing.T) {
	assert := assert.New(t)
	ctx := whisper.Whisper_init(MODEL)
	assert.NotNil(ctx)
	ctx.Whisper_free()
}

func Test_Whisper_001(t *testing.T) {
	assert := assert.New(t)
	ctx := whisper.Whisper_init(MODEL)
	assert.NotNil(ctx)
	defer ctx.Whisper_free()

	params := ctx.Whisper_full_default_params(whisper.WHISPER_SAMPLING_GREEDY)
	t.Log(params)
}
