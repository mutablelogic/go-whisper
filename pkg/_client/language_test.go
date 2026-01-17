package client_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-whisper/pkg/client"
	"github.com/stretchr/testify/assert"
)

///////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_Language_001(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		in, out1, out2 string
	}{
		{"english", "en", "eng"},
		{"en", "en", "eng"},
		{"eng", "en", "eng"},
		{"german", "de", "deu"},
	}
	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			test1, test2 := client.LanguageCode(test.in)
			assert.Equal(test.out1, test1)
			assert.Equal(test.out2, test2)
		})
	}
}
