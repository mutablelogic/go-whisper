package whisper

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/whisper.cpp
#cgo LDFLAGS: -L${SRCDIR}/../../build -lwhisper -lm -lstdc++
#include <whisper.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	whisper_context C.struct_whisper_context
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Allocates all memory needed for the model and loads the model from the given file.
// Returns NULL on failure.
func Whisper_init(path string) *whisper_context {
	if ctx := C.whisper_init(C.CString(path)); ctx != nil {
		return (*whisper_context)(ctx)
	} else {
		return nil
	}
}

// Frees all memory allocated by the model.
func (ctx *whisper_context) Whisper_free() {
	C.whisper_free((*C.struct_whisper_context)(ctx))
}

/*
   WHISPER_API struct whisper_context * whisper_init(const char * path_model);

   WHISPER_API void whisper_free(struct whisper_context * ctx);
*/
