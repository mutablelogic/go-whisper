# go-whisper

<<<<<<< HEAD
Speech-to-Text in golang. This is an early development version.

  * `cmd` are the start of the command-line tools
  * `third_party` is a submodule for the whisper.cpp source

The bindings for whisper are here:

  * https://github.com/ggerganov/whisper.cpp/tree/master/bindings/go

The API documentation for those bindings is here:

  * https://pkg.go.dev/github.com/ggerganov/whisper.cpp/bindings/go
  * https://pkg.go.dev/github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper

=======
Speech-to-Text in golang. These is an early development version:

  * `sys/whisper` directory contains the basic bindings
  * `pkg/whisper` provides a more golang-like package
  * `cmd` are the start of the command-line tools
  * `third_party` is a submodule for the whisper.cpp source

>>>>>>> 03e06c3edefcf0c4c74cb4ba7e8263e97d81fcd8
## Building

The following `Makefile` targets can be used:

  * `make submodule` - fetches the `third_party` submodules
  * `make whisper` - builds the `whisper` library
  * `make models` downloads the models to the `models` directory
  * `make cmd` builds the command-line tools, and places them in `build` directory
<<<<<<< HEAD
=======
  * `make test` runs the tests

>>>>>>> 03e06c3edefcf0c4c74cb4ba7e8263e97d81fcd8

## Status

This is a work in progress!


