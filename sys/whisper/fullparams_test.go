package whisper_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-whisper/sys/whisper"
)

func Test_fullparams_00(t *testing.T) {
	var params = whisper.DefaultFullParams(whisper.SAMPLING_GREEDY)
	t.Log(params)
}
