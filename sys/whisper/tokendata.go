package whisper

import (
	"encoding/json"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo LDFLAGS: -lwhisper -lggml -lm -lstdc++
#cgo darwin LDFLAGS: -framework Accelerate -framework Metal -framework Foundation -framework CoreGraphics
#include <whisper.h>
#include <stdlib.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *TokenData) MarshalJSON() ([]byte, error) {
	type j struct {
		Id    Token   `json:"id"`
		Tid   Token   `json:"tid"`
		P     float32 `json:"p"`
		Plog  float32 `json:"plog"`
		Pt    float32 `json:"pt"`
		Ptsum float32 `json:"ptsum"`
		T0    int64   `json:"t0,omitempty"`
		T1    int64   `json:"t1,omitempty"`
		Vlen  float32 `json:"vlen"`
	}
	return json.Marshal(j{
		Id:    Token(t.id),
		Tid:   Token(t.tid),
		P:     float32(t.p),
		Plog:  float32(t.plog),
		Pt:    float32(t.pt),
		Ptsum: float32(t.ptsum),
		T0:    int64(t.t0),
		T1:    int64(t.t1),
		Vlen:  float32(t.vlen),
	})
}

func (t TokenData) String() string {
	data, err := json.MarshalIndent(&t, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
