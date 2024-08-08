//go:build cuda

package whisper

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: cuda-12.6 cublas-12.6 cudart-12.6
#cgo LDFLAGS: -L/usr/local/cuda/lib64/stubs -lcuda
*/
import "C"
