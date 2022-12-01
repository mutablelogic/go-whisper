package whisper_test

import (
	"testing"

	whisper "github.com/djthorpe/go-whisper/sys/whisper"
)

func Test_Whisper_000(t *testing.T) {
	ctx := whisper.Whisper_init("/usr/local/share/whisper/model")
	defer ctx.Whisper_free()
}
