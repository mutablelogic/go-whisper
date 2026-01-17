# go-whisper

[![Go Reference](https://pkg.go.dev/badge/github.com/mutablelogic/go-whisper.svg)](https://pkg.go.dev/github.com/mutablelogic/go-whisper)
[![License](https://img.shields.io/badge/license-Apache-blue.svg)](LICENSE)

Speech-to-Text in golang using [whisper.cpp](https://github.com/ggerganov/whisper.cpp).

## Features

- **Transcription & Translation**: Easily transcribe audio files and translate them to English
- **Providers**: Use models from OpenAI, ElevenLabs, and GGML
- **Command Line Interface**: Simple CLI for transcription and managing models
- **HTTP API Server**: OpenAPI-compatible server with streaming support
- **Model Management**: Download, list, and delete models
- **GPU Acceleration**: Support for CUDA, Vulkan, and Metal (macOS) acceleration
- **Docker Support**: Pre-built images for amd64 and arm64 architectures

For more information on features, see the [Features](doc/features.md) document.

## Project Structure

- `cmd` contains the command-line tool, which can also be run as an OpenAPI-compatible HTTP server
- `pkg` contains the `whisper` service and client
- `sys` contains the `whisper` bindings to the `whisper.cpp` library
- `third_party` is a submodule for the whisper.cpp source, and ffmpeg bindings

The following sections describe how to use whisper on the command-line, run it as a service,
download a model, run the server, and build the project.

## Using Docker

You can run whisper as a CLI command or in a Docker container. There are Docker images for arm64 and amd64 (Intel),
but these are currently not optimized for GPU, and are not recommended.

Support for CUDAin the docker container is still under development. When completed, you'll need to install the
[NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) first.

A Docker volume called "whisper" can be used for storing the Whisper language
models. You can see which models are available to download from the [HuggingFace whisper.cpp repository](https://huggingface.co/ggerganov/whisper.cpp).

The following command will run the server on port 8080 for an NVIDIA GPU:

```bash
docker volume create whisper
docker run \
  --name whisper-server --rm \
  --runtime nvidia --gpus all \ # When using a NVIDIA GPU
  -v whisper:/data -p 8080:80 \
  ghcr.io/mutablelogic/go-whisper:latest
```

The API is then available at `http://localhost:8080/api/v1` and it generally conforms to the [OpenAI API](https://platform.openai.com/docs/api-reference/audio) spec.

## API Examples

The API is available through the server and conforms generally to the OpenAI API spec. Here are some common usage examples:

### Download a model

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"path": "ggml-medium-q5_0.bin"}' \
  localhost:8080/api/v1/models?stream=true
```

### List available models

```bash
curl -X GET localhost:8080/api/v1/models
```

### Delete a model

```bash
curl -X DELETE localhost:8080/api/v1/models/ggml-medium-q5_0
```

### Transcribe an audio file

```bash
curl -F model=ggml-medium-q5_0 \
  -F file=@samples/jfk.wav \
  localhost:8080/api/v1/audio/transcriptions?stream=true
```

### Translate an audio file to English

```bash
curl -F model=ggml-medium-q5_0 \
  -F file=@samples/de-podcast.wav \
  -F language=en \
  localhost:8080/api/v1/audio/translations?stream=true
```

For more detailed API documentation, see the [API Reference](doc/API.md).

## Building

### Docker Images

If you are building a Docker image, you just need make and Docker installed:

- `GGML_CUDA=1 DOCKER_REGISTRY=docker.io/user make docker` - builds a Docker container with the server binary for CUDA, tagged to a specific registry
- `GGML_VULKAN=1 make docker` - builds a Docker container with the server binary for Vulkan
- `OS=linux DOCKER_REGISTRY=docker.io/user make docker` - builds a Docker container for Linux, with the server binary without CUDA, tagged to a specific registry

### From Source

It's recommended (especially for MacOS) to build the `whisper` binary without Docker, to utilize GPU acceleration.
You can use the `Makefile` in the root directory and have the following dependencies met:

- Recent version of Go (ie, 1.22+)
- C++ compiler and cmake
- For CUDA, you'll need the CUDA toolkit installed including the `nvcc` compiler
- For Vulkan, you'll need the Vulkan SDK installed
  - For the Rasperry Pi, install the following additional packages first: `sudo apt install libvulkan-dev libvulkan1 mesa-vulkan-drivers glslc`
- For Metal, you'll need Xcode installed on macOS
- For audio and video codec support (ie, x264, AAC, etc) when extracting the audio, you'll need to install appropriate codecs before building (see below).

The following `Makefile` targets can be used:

- `make` - creates the server binary, and places it in the `build` directory. Should link to Metal on macOS
- `GGML_CUDA=1 make whisper` - creates the server binary linked to CUDA, and places it in the `build` directory. Should work for amd64 and arm64 (Jetson) platforms
- `GGML_VULKAN=1 make whisper` - creates the server binary linked to Vulkan, and places it in the `build` directory. 

See all the other targets and variations in the `Makefile` for more information.

## Command Line Usage

The `whisper` command-line tool can be built with `make whisper` and provides various functionalities, both for running `whipser` directly
and for calling a transcriptions and translations service remotely.

```bash
# List available models
whisper models

# Download a model
whisper download ggml-medium-q5_0.bin

# Delete a model
whisper delete ggml-medium-q5_0

# Transcribe an audio file
whisper transcribe ggml-medium-q5_0 samples/jfk.wav

# Translate an audio file to English
whisper translate ggml-medium-q5_0 samples/de-podcast.wav

# Run the whisper server
whisper server --listen localhost:8080
```

You can also access transcription and translation functionalities from OpenAI-compatible and ElevenLabs-compatible services:

- Set `OPENAI_API_KEY` environment variable to your OpenAI API key to use the OpenAI-compatible endpoints.
- Set `ELEVENLABS_API_KEY` environment variable to your ElevenLabs API key
- Set `WHISPER_URL` environment variable to  the URL of the whisper server to use the OpenAI-compatible endpoints.

```bash
# List available remote models (including OpenAI and ElevenLabs models)
whisper models --remote

# Download a model (gowhisper service)
whisper download ggml-medium-q5_0.bin --remote

# Transcribe an audio file for subtitles (ElevenLabs)
whisper transcribe scribe_v1 samples/jfk.wav --format srt --diarize --remote

# Translate an audio file to English (OpenAI)
whisper translate whisper-1 samples/de-podcast.wav  --remote
```

## Contributing & License

This project is currently in development and subject to change. Please file feature requests and bugs 
in the [GitHub issues](https://github.com/mutablelogic/go-whisper/issues).
The license is Apache 2 so feel free to redistribute. Redistributions in either source
code or binary form must reproduce the copyright notice, and please link back to this
repository for more information:

> **go-whisper**\
> [https://github.com/mutablelogic/go-whisper/](https://github.com/mutablelogic/go-whisper/)\
> Copyright (c) David Thorpe, All rights reserved.
>
> **whisper.cpp**\
> [https://github.com/ggerganov/whisper.cpp](https://github.com/ggerganov/whisper.cpp)\
> Copyright (c) The ggml authors
>
> **ffmpeg**\
> [https://ffmpeg.org/](https://ffmpeg.org/)\
> Copyright (c) the FFmpeg developers

This software links to static libraries of [whisper.cpp](https://github.com/ggerganov/whisper.cpp) licensed under
the [MIT License](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html). This software links to static libraries of ffmpeg licensed under the
[LGPL 2.1 License](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html). 
