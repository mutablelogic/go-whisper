package whisper

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libwhisper
#cgo darwin pkg-config: libwhisper-darwin
#cgo linux pkg-config: libwhisper-linux
#cgo arm64 pkg-config: cuda-12.2 cublas-12.2 cudart-12.2
*/
import "C"

// Generate the whisper pkg-config files
// Setting the prefix to the base of the repository
//go:generate go run ../pkg-config --version "0.0.0" --prefix "../.." --cflags "-I$DOLLAR{prefix}/third_party/whisper.cpp/include -I$DOLLAR{prefix}/third_party/whisper.cpp/ggml/include" --libs "-L$DOLLAR{prefix}/third_party/whisper.cpp -lwhisper -lggml -lm -lstdc++" libwhisper.pc
//go:generate go run ../pkg-config --version "0.0.0" --libs "-framework Accelerate -framework Metal -framework Foundation -framework CoreGraphics" libwhisper-darwin.pc
//go:generate go run ../pkg-config --version "0.0.0" --libs "-lgomp" libwhisper-linux.pc
