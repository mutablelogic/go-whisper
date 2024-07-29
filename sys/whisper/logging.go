package whisper

import (
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libwhisper
#cgo darwin pkg-config: libwhisper-darwin
#cgo linux pkg-config: libwhisper-linux

#include <whisper.h>
#include <stdlib.h>

extern void callLog(enum ggml_log_level level,char* text, void* user_data);

// Logging callback
static void whisper_log_cb(enum ggml_log_level level, const char* text, void* user_data) {
	callLog(level, (char*)text, user_data);
}

// Set or unset logging callback
static void whisper_log_set_ex(void* user_data) {
	if (user_data) {
		whisper_log_set(whisper_log_cb, user_data);
	} else {
		whisper_log_set(NULL, NULL);
	}
}
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	LogLevel C.enum_ggml_log_level
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	cbLog func(level LogLevel, text string)
)

const (
	LogLevelDebug LogLevel = C.GGML_LOG_LEVEL_DEBUG
	LogLevelInfo  LogLevel = C.GGML_LOG_LEVEL_INFO
	LogLevelWarn  LogLevel = C.GGML_LOG_LEVEL_WARN
	LogLevelError LogLevel = C.GGML_LOG_LEVEL_ERROR
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v LogLevel) String() string {
	switch v {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "LOG"
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set logging output
func Whisper_log_set(fn func(level LogLevel, text string)) {
	cbLog = fn
	if fn == nil {
		C.whisper_log_set_ex(nil)
	} else {
		C.whisper_log_set_ex((unsafe.Pointer)(uintptr(1)))
	}
}

// Call logging output
func Whisper_log(level LogLevel, text string, user_data unsafe.Pointer) {
	cStr := C.CString(text)
	defer C.free(unsafe.Pointer(cStr))
	C.whisper_log_cb((C.enum_ggml_log_level)(level), cStr, user_data)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

//export callLog
func callLog(level C.enum_ggml_log_level, text *C.char, user_data unsafe.Pointer) {
	if cbLog != nil {
		cbLog(LogLevel(level), C.GoString(text))
	}
}
