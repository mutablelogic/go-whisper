# Running gowhisper with Docker

This guide covers running gowhisper in Docker with different GPU acceleration options.

## Docker Images

| Image | Description |
|-------|-------------|
| `ghcr.io/mutablelogic/go-whisper` | Vulkan support (multi-arch: amd64, arm64) |
| `ghcr.io/mutablelogic/go-whisper-cuda` | CUDA support for NVIDIA GPUs (multi-arch: amd64, arm64) |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GOWHISPER_DIR` | Directory for storing models | `/data` |
| `GOWHISPER_ADDR` | HTTP listen address | `0.0.0.0:8081` |
| `OPENAI_API_KEY` | OpenAI API key (optional) | |
| `ELEVENLABS_API_KEY` | ElevenLabs API key (optional) | |

## Running on CPU Only

For CPU-only operation, no special device flags are needed:

```bash
docker run --rm -p 8081:8081 \
  -v /path/to/models:/data \
  ghcr.io/mutablelogic/go-whisper run
```

## Running with Vulkan GPU Acceleration

### Prerequisites (Host)

```bash
# Install Vulkan tools
sudo apt install vulkan-tools

# Verify GPU is detected (should show something other than "llvmpipe")
vulkaninfo --summary
```

### AMD / Intel

```bash
docker run --rm -p 8081:8081 \
  --device=/dev/dri:/dev/dri \
  -v /usr/share/vulkan/icd.d:/usr/share/vulkan/icd.d:ro \
  -v /path/to/models:/data \
  ghcr.io/mutablelogic/go-whisper run --debug
```

> **Note:** Raspberry Pi does not support Vulkan GPU acceleration for whisper due to hardware limitations. Use CPU-only mode instead.

## Running with CUDA GPU Acceleration

### NVIDIA (Desktop, Server, and Jetson)

```bash
# Install NVIDIA container toolkit
sudo apt install nvidia-container-toolkit
sudo systemctl restart docker

docker run --rm -p 8081:8081 --runtime nvidia --gpus all \
  -v /path/to/models:/data \
  ghcr.io/mutablelogic/go-whisper-cuda run --debug
```

## Troubleshooting

### Verify Vulkan GPU Access

Run `vulkaninfo` inside the container to check if the GPU is accessible:

```bash
# AMD/Intel
docker run --rm -it \
  --device=/dev/dri:/dev/dri \
  -v /usr/share/vulkan/icd.d:/usr/share/vulkan/icd.d:ro \
  --entrypoint vulkaninfo \
  ghcr.io/mutablelogic/go-whisper --summary
```

If you see `deviceName = llvmpipe`, the GPU is not being passed through correctly.

### Check NVIDIA Driver

```bash
# On host
nvidia-smi

# In container (CUDA images only)
docker run --rm --runtime nvidia --gpus all --entrypoint nvidia-smi ghcr.io/mutablelogic/go-whisper-cuda
```

## Production Deployment

For production deployments, an example [Hashicorp Nomad](https://www.nomadproject.io/) job file is provided at [etc/gowhisper.nomad.hcl](../etc/gowhisper.nomad.hcl).

To deploy with Nomad:

```bash
# Create a variables file (gowhisper.vars.hcl)
dc            = ["dc1"]
data          = "/path/to/models"
docker_image  = "ghcr.io/mutablelogic/go-whisper-cuda"
docker_runtime = "nvidia"
devices       = ["/dev/nvidia0", "/dev/nvidiactl", "/dev/nvidia-uvm"]

# Run the job
nomad job run -var-file=gowhisper.vars.hcl etc/gowhisper.nomad.hcl
```

See the job file for all available configuration variables.
