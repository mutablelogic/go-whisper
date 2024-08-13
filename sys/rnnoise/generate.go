package rnnoise

// Generate the rnnoise pkg-config files
// Setting the prefix to the base of the repository
//go:generate go run ../pkg-config --version "0.0.0" --prefix "../.." --cflags "-I$DOLLAR{prefix}/third_party/rnnoise/include" --libs "-L$DOLLAR{prefix}/third_party/rnnoise/build -lrnnoise" rnnoise.pc
