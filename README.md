# go-whisper

Speech-to-Text in golang. These is an early development version:

  * `sys/whisper` directory contains the basic bindings
  * `pkg/whisper` provides a more golang-like package
  * `cmd` are the start of the command-line tools
  * `third_party` is a submodule for the whisper.cpp source

## Building

The following `Makefile` targets can be used:

  * `make submodule` - fetches the `third_party` submodules
  * `make whisper` - builds the `whisper` library
  * `make models` downloads the models to the `models` directory
  * `make cmd` builds the command-line tools, and places them in `build` directory
  * `make test` runs the tests


## Status

This is a work in progress!


