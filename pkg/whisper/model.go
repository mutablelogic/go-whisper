package whisper

import (
	"os"

	// Packages
	whisper "github.com/ggerganov/whisper.cpp/bindings/go"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type model struct {
	path string
	ctx  *whisper.Context
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(path string) (*model, error) {
	model := new(model)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	} else if ctx := whisper.Whisper_init(path); ctx == nil {
		return nil, ErrUnableToLoadModel
	} else {
		model.ctx = ctx
		model.path = path
	}

	// Return success
	return model, nil
}

func (model *model) Close() error {
	if model.ctx != nil {
		model.ctx.Whisper_free()
	}

	// Release resources
	model.ctx = nil

	// Return success
	return nil
}
