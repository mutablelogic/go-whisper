package whisper_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-whisper/sys/whisper"
)

func Test_contextparams_00(t *testing.T) {
	params := whisper.DefaultContextParams()
	t.Log(params)
}
