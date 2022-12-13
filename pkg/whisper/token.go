package whisper

import (
	"fmt"

	// Packages
	whisper "github.com/djthorpe/go-whisper/sys/whisper"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Token struct {
	id   whisper.Whisper_token
	text string
	p    float32
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (t *Token) Text() string {
	return t.text
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *Token) String() string {
	str := "<whisper.token"
	str += fmt.Sprint(" id=", t.id)
	str += fmt.Sprintf(" text=%q", t.text)
	str += fmt.Sprintf(" p=%v", t.p)
	str += ">"
	return str
}
