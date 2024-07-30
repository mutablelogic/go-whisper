package whisper

import "encoding/json"

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libwhisper
#include <whisper.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	ContextParams C.struct_whisper_context_params
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func DefaultContextParams() ContextParams {
	return (ContextParams)(C.whisper_context_default_params())
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx ContextParams) MarshalJSON() ([]byte, error) {
	type j struct {
		UseGpu          bool `json:"use_gpu"`
		GpuDevice       int  `json:"gpu_device"`
		FlashAttn       bool `json:"flash_attn"`
		TokenTimestamps bool `json:"dtw_token_timestamps"`
	}
	return json.Marshal(j{
		UseGpu:          bool(ctx.use_gpu),
		GpuDevice:       int(ctx.gpu_device),
		FlashAttn:       bool(ctx.flash_attn),
		TokenTimestamps: bool(ctx.dtw_token_timestamps),
	})
}

func (ctx ContextParams) String() string {
	str, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(str)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (ctx *ContextParams) SetUseGpu(v bool) {
	ctx.use_gpu = (C.bool)(v)
}

func (ctx *ContextParams) SetGpuDevice(v int) {
	ctx.gpu_device = (C.int)(v)
}

func (ctx *ContextParams) SetFlashAttn(v bool) {
	ctx.flash_attn = (C.bool)(v)
}

func (ctx *ContextParams) SetTokenTimestamps(v bool) {
	ctx.dtw_token_timestamps = (C.bool)(v)
}
