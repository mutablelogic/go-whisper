# go-whisper

Speech-to-Text in golang. This is an early development version.

* `cmd` contains an OpenAI-API compatible server
* `pkg` contains the `whisper` service and http gateway
* `sys` contains the `whisper` bindings to the `whisper.cpp` library
* `third_party` is a submodule for the whisper.cpp source

## Running

There are docker images for arm64 and amd64 (Intel). The arm64 image is built for
Jetson GPU support specifically, but it will also run on Raspberry Pi's.

In order to utilize a NVIDIA GPU, you'll need to have the 
[NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) first.

Have a volume available called "models" which can be used for storing the Whisper language
models. You can see which models are available [here](https://huggingface.co/ggerganov/whisper.cpp).

The following command will run the server on port 8080:

```bash
docker run \
  --name whisper-server --rm \
  --runtime nvidia --gpus all \ # When using a NVIDIA GPU
  -v whisper:/models -p 8080:8080 -e WHISPER_DATA=/models \
  ghcr.io/mutablelogic/go-whisper:latest
```

If you include a `-debug` flag at the end, you'll get more verbose output. The API is then
available at `http://localhost:8080/v1` and it generally conforms to the
[OpenAI API](https://platform.openai.com/docs/api-reference/audio).

In order to download a model, you can use the following command (for example):

```bash
curl -X POST -H "Content-Type: application/json" -d '{"Path" : "ggml-tiny.en-q8_0.bin" }' localhost:8080/v1/models  
```

To list the models available, you can use the following command:

```bash
curl -X GET localhost:8080/v1/models
```

To delete a model, you can use the following command:

```bash
curl -X DELETE localhost:8080/v1/models/ggml-tiny.en-q8_0.bin
```

And to transcribe an audio file, you can use the following command:

```bash
curl -F "model=ggml-tiny.en-q8_0.bin" -F "file=@samples/jfk.wav" -F "language=en" localhost:8080/v1/audio/transcriptions
```

There's more information on the API [here](doc/API.md).

## Building

If you want to build the server yourself for your specific combination of hardware,
you can use the `Makefile` in the root directory. You'll need go 1.22, `make` and 
a C++ compiler to build this project. The following `Makefile` targets can be used:

* `make server` - creates the server binary, and places it in the `build` directory
* `DOCKER_REGISTRY=docker.io/user make docker` - builds a docker container with the server binary

See all the other targets in the `Makefile` for more information.

## Status

Still in development. It only accepts mono WAV files at 16K sample rate, for example. It also
occasionally crashes, and the API is not fully implemented.


## Contributing & Distribution

__This module is currently in development and subject to change.__

Please do file feature requests and bugs [here](https://github.com/mutablelogic/go-whisper/issues).
The license is Apache 2 so feel free to redistribute. Redistributions in either source
code or binary form must reproduce the copyright notice, and please link back to this
repository for more information:

> __go-media__\
> [https://github.com/mutablelogic/go-whisper/](https://github.com/mutablelogic/go-whisper/)\
> Copyright (c) 2023-2024 David Thorpe, All rights reserved.
>
> __whisper.cpp__\
> [https://github.com/ggerganov/whisper.cpp](https://github.com/ggerganov/whisper.cpp)\
> Copyright (c) 2023-2024 The ggml authors

This software links to static libraries of [whisper.cpp](https://github.com/ggerganov/whisper.cpp) licensed under
the [MIT License](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html).
