package main

import (
	"reflect"
	"time"
	"unsafe"

	// Packages
	"github.com/djthorpe/go-whisper/pkg/whisper"
)

type Buffer struct {
	bytes []byte
}

func NewBuffer(size time.Duration) *Buffer {
	b := new(Buffer)
	b.bytes = make([]byte, int(size.Seconds()*whisper.SAMPLE_RATE*whisper.SAMPLE_SIZE))
	return b
}

func (b *Buffer) Bytes() []byte {
	return b.bytes
}

func (b *Buffer) Samples() []float32 {
	var samples []float32
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&samples)))
	sliceHeader.Cap = int(len(b.bytes) / whisper.SAMPLE_SIZE)
	sliceHeader.Len = int(len(b.bytes) / whisper.SAMPLE_SIZE)
	sliceHeader.Data = uintptr(unsafe.Pointer(&b.bytes[0]))
	return samples
}
