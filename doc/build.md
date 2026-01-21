# Building

You can build `gowhisper` in one of three ways:

1. Using Docker to build a container with the server binary, with either Vulkan or CUDA support built-in
2. Building the server from source using the provided `Makefile`, with either Vulkan, CUDA, or Metal support. This is the best option for macOS users to get GPU acceleration.
3. Building a CLI-only version of `gowhisper` without any server component

The following sections outline each method.

## Docker Images

If you are building a Docker image, you just need `make` and `docker` installed. Some examples:

- `GGML_CUDA=1 DOCKER_FILE=etc/Dockerfile.cuda DOCKER_REGISTRY=docker.io/user make docker` - builds a Docker container with the server binary for CUDA, tagged to a specific registry
- `GGML_VULKAN=1 make docker` - builds a Docker container with the server binary for Vulkan
- `OS=linux DOCKER_REGISTRY=docker.io/user make docker` - builds a Docker container for Linux, with the server binary without CUDA, tagged to a specific registry

If you are logged into Docker Hub, you can then use `make docker-push` to push the built image to your registry (with the appropriate `DOCKER_REGISTRY` set).

## From Source

It's recommended (especially for MacOS) to build the `whisper` binary without Docker, to utilize GPU acceleration or native CPU support, since the docker images do not utilize
any optimizations for your specific CPU.

However, in order to build from source, there are additional dependencies that need to be met:

| Dependency | Mac/Homebrew | Debian/Ubuntu | Fedora |
|------------|--------------|---------------|--------|
| Build tools | `xcode-select --install && brew install cmake pkg-config nasm` | `sudo apt install build-essential git cmake pkg-config nasm curl` | `sudo dnf install make git cmake pkg-config nasm gcc-c++` |
| Go (1.24+) | `brew install go` | [Download from golang.org](https://golang.org/dl/) | `sudo dnf install golang` |
| CUDA toolkit (12.6+) | [NVIDIA CUDA Toolkit](https://developer.nvidia.com/cuda-downloads) | [NVIDIA CUDA Toolkit](https://developer.nvidia.com/cuda-downloads) | [NVIDIA CUDA Toolkit](https://developer.nvidia.com/cuda-downloads) |
| Vulkan SDK | `brew install vulkan-sdk` | `sudo apt install libvulkan-dev libplacebo-dev libshaderc-dev glslang-tools glslc` | `sudo dnf install vulkan-loader-devel libplacebo-devel libshaderc-devel glslang glslc` |
| Audio/Video codecs (optional) | `brew install lame opus libvorbis libvpx x264 x265 dav1d` | `sudo apt install libmp3lame-dev libopus-dev libvorbis-dev libvpx-dev libx264-dev libx265-dev libdav1d-dev libnuma-dev` | `sudo dnf install lame-devel opus-devel libvorbis-devel libvpx-devel x264-devel x265-devel dav1d-devel numactl-devel` |

The following `Makefile` targets can be used:

- `make` - Build the server binary with Metal support (macOS)
- `GGML_CUDA=1 make` - Build the server binary with CUDA support (amd64 and arm64/Jetson)
- `GGML_VULKAN=1 make` - Build the server binary with Vulkan support

The resulting `gowhisper` binary is placed in the `build` directory.

## Client-Only Build

If you only need the CLI client (without the server component or GPU acceleration), you can build a minimal version with just:

- GNU Make
- Go (1.24+)

Build command:

```bash
make gowhisper-client
```

The resulting `gowhisper` binary is placed in the `build` directory and can connect to a remote gowhisper server.
