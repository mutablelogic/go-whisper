# go-whisper

Speech-to-Text in golang. This is an early development version.

* `cmd` contains an OpenAI-API compatible server
* `pkg` contains the `whisper` service and http gateway
* `sys` contains the `whisper` bindings to the `whisper.cpp` library
* `third_party` is a submodule for the whisper.cpp source

## Running

In order to utilize a NVIDIA GPU, you'll need to have the CUDA toolkit installed.

## Building

You'll need go 1.22, make and a C++ compiler to build this project.
The following `Makefile` targets can be used:

* `make server` - creates the server binary, and places it in the `build` directory
* `make docker` - builds a docker container with the server binary

The targetted architectures are `amd64` and `arm64`. The docker build includes CUDA support
for the `amd64` architecture and Jetson CUDA support for the `arm64` architecture.
