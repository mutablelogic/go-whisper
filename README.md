# go-whisper

Speech-to-Text in golang. This is an early development version.

  * `cmd` are the start of the command-line tools
  * `third_party` is a submodule for the whisper.cpp source

The bindings for whisper are here:

  * https://github.com/ggerganov/whisper.cpp/tree/master/bindings/go

The API documentation for those bindings is here:

  * https://pkg.go.dev/github.com/ggerganov/whisper.cpp/bindings/go
  * https://pkg.go.dev/github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper

## Building

The following `Makefile` targets can be used:

  * `make submodule` - fetches the `third_party` submodules
  * `make whisper` - builds the `whisper` library
  * `make models` downloads the models to the `models` directory
  * `make cmd` builds the command-line tools, and places them in `build` directory

## Download Model

```bash
./build/go-model-download -out models ggml-small.en.bin
```

## Streaming Translation

There is one example of streaming translation in the `cmd` directory. It presently creates a lot
of problems (segmentation fauls, assertation errors), but it is a start. There's probbably some thread
safety issues in the `whisper` library?

```bash
make all
./build/stream -model models/ggml-base.en.bin -device 0 -window 5s
whisper_model_load: loading model from 'models/ggml-base.en.bin'
[speak now, press CTRL+C to stop]

[2.048s->4.048s] 4 score and
[4.096s->6.096s] 7 years ago our father
[6.144s->8.144s] was brought forth on this
[8.192s->10.192s] continent and new nation
[10.24s->12.24s] conceived in liberty
[12.288s->14.288s] and dedicated
[14.336s->16.336s] to the proposition the new man
[16.384s->18.384s] are created equal
^C
[stop speaking]
```
