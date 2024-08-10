# go-whisper

Speech-to-Text in golang. This is an early development version.

* `cmd` contains an OpenAI-API compatible service
* `pkg` contains the `whisper` service and client
* `sys` contains the `whisper` bindings to the `whisper.cpp` library
* `third_party` is a submodule for the whisper.cpp source

## Running

(Note: Docker images are not created yet - this is some forward planning!)

You can either run the whisper service as a CLI command or in a docker container.
There are docker images for arm64 and amd64 (Intel). The arm64 image is built for
Jetson GPU support specifically, but it will also run on Raspberry Pi's.

In order to utilize a NVIDIA GPU, you'll need to install the
[NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) first.

A docker volume should be created called "whisper" can be used for storing the Whisper language
models. You can see which models are available to download locally [here](https://huggingface.co/ggerganov/whisper.cpp).

The following command will run the server on port 8080 for an NVIDIA GPU:

```bash
docker run \
  --name whisper-server --rm \
  --runtime nvidia --gpus all \ # When using a NVIDIA GPU
  -v whisper:/models -p 8080:8080 -e WHISPER_DATA=/models \
  ghcr.io/mutablelogic/go-whisper
```

If you include a `-debug` flag at the end, you'll get more verbose output. The API is then
available at `http://localhost:8080/v1` and it generally conforms to the
[OpenAI API](https://platform.openai.com/docs/api-reference/audio) spec.

### Sample Usage

In order to download a model, you can use the following command (for example):

```bash
curl -X POST -H "Content-Type: application/json" -d '{"Path" : "ggml-medium-q5_0.bin" }' localhost:8080/v1/models\?stream=true
```

To list the models available, you can use the following command:

```bash
curl -X GET localhost:8080/v1/models
```

To delete a model, you can use the following command:

```bash
curl -X DELETE localhost:8080/v1/models/ggml-medium-q5_0
```

To transcribe a media file into it's original language, you can use the following command:

```bash
curl -F model=ggml-medium-q5_0 -F file=@samples/jfk.wav localhost:8080/v1/audio/transcriptions\?stream=true
```

To translate a media file into a different language, you can use the following command:

```bash
curl -F model=ggml-medium-q5_0 -F file=@samples/de-podcast.wav -F language=en localhost:8080/v1/audio/translations\?stream=true
```

There's more information on the API [here](doc/API.md).

## Building

If you are building a docker image, you just need Docker installed:

* `DOCKER_REGISTRY=docker.io/user make docker` - builds a docker container with the
  server binary, tagged to a specific registry

If you want to build the server yourself for your specific combination of hardware,
you can use the `Makefile` in the root directory and have the following dependencies
met:

* Go 1.22
* C++ compiler
* FFmpeg 6.1 libraries (see [here](doc/build.md) for more information)
* For CUDA, you'll need the CUDA toolkit including the `nvcc` compiler

The following `Makefile` targets can be used:

* `make server` - creates the server binary, and places it in the `build` directory. Should
  link to Metal on macOS
* `GGML_CUDA=1 make server` - creates the server binary linked to CUDA, and places it
  in the `build` directory. Should work for amd64 and arm64 (Jetson) platforms

See all the other targets in the `Makefile` for more information.

## Developing

The `cmd/examples` directory contains a simple example of how to use the `whisper` package
in your own code.

## Status

Still in development. See this [issue](https://github.com/mutablelogic/go-whisper/issues/1) for
remaining tasks to be completed.

## Contributing & Distribution

__This module is currently in development and subject to change.__

Please do file feature requests and bugs [here](https://github.com/mutablelogic/go-whisper/issues).
The license is Apache 2 so feel free to redistribute. Redistributions in either source
code or binary form must reproduce the copyright notice, and please link back to this
repository for more information:

> __go-whisper__\
> [https://github.com/mutablelogic/go-whisper/](https://github.com/mutablelogic/go-whisper/)\
> Copyright (c) 2023-2024 David Thorpe, All rights reserved.
>
> __whisper.cpp__\
> [https://github.com/ggerganov/whisper.cpp](https://github.com/ggerganov/whisper.cpp)\
> Copyright (c) 2023-2024 The ggml authors

This software links to static libraries of [whisper.cpp](https://github.com/ggerganov/whisper.cpp) licensed under
the [MIT License](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html).
