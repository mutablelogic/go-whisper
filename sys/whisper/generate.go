package whisper

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libwhisper
#cgo linux pkg-config: libwhisper-linux
#cgo darwin pkg-config: libwhisper-darwin
*/
import "C"

// Generate the whisper pkg-config files
// Setting the prefix to the base of the repository
//go:generate go run ../pkg-config --version "0.0.0" --prefix "../.." --cflags "-I$DOLLAR{prefix}/third_party/whisper.cpp/include -I$DOLLAR{prefix}/third_party/whisper.cpp/ggml/include" libwhisper.pc
//go:generate go run ../pkg-config --version "0.0.0" --prefix "../.." --cflags "-fopenmp" --libs "-L$DOLLAR{prefix}/third_party/whisper.cpp -lwhisper -lggml -lgomp -lm -lstdc++" libwhisper-linux.pc
//go:generate go run ../pkg-config --version "0.0.0" --prefix "../.." --libs "-L$DOLLAR{prefix}/third_party/whisper.cpp -lwhisper -lggml -lm -lstdc++ -framework Accelerate -framework Metal -framework Foundation -framework CoreGraphics" libwhisper-darwin.pc
