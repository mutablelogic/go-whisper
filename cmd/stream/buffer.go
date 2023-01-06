package main

import (
	"reflect"
	"time"
	"unsafe"
	// Packages
)

type Buffer struct {
	bytes []byte
}

func NewBuffer(size time.Duration) *Buffer {
	b := new(Buffer)
	b.bytes = make([]byte, int(size.Seconds()*16000*4)) // whisper.SAMPLE_RATE*whisper.SAMPLE_SIZE))
	return b
}

func (b *Buffer) Bytes() []byte {
	return b.bytes
}

func (b *Buffer) Samples() []float32 {
	var samples []float32
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&samples)))
	sliceHeader.Cap = int(len(b.bytes) / 4) //whisper.SAMPLE_SIZE)
	sliceHeader.Len = int(len(b.bytes) / 4) // whisper.SAMPLE_SIZE)
	sliceHeader.Data = uintptr(unsafe.Pointer(&b.bytes[0]))
	return samples
}
