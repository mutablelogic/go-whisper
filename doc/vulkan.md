# Docker GPU Acceleration with Vulkan

The go-whisper Docker image includes Vulkan support for hardware-accelerated video encoding/decoding. To use GPU acceleration from within the container, you must pass through GPU devices from the host.

## Prerequisites

Install GPU drivers and Vulkan tools on your host system:

```bash
# NVIDIA
apt install nvidia-container-toolkit vulkan-tools
systemctl restart docker

# AMD/Intel/Raspberry Pi
apt install mesa-vulkan-drivers vulkan-tools

# Verify - you should see a GPU other than "deviceName = llvmpipe"
vulkaninfo --summary

# Pull Vulkan-enabled Docker image
docker pull ghcr.io/mutablelogic/go-whisper
```

## Running with GPU Acceleration

```bash
# NVIDIA
docker run --rm --name gowhisper -p 8081:8081 --gpus all ghcr.io/mutablelogic/go-whisper run --debug

# NVIDIA Jetson/Tegra
docker run --rm --name gowhisper -p 8081:8081 --runtime nvidia \
  --device=/dev/nvhost-ctrl --device=/dev/nvhost-ctrl-gpu \
  --device=/dev/nvhost-prof-gpu --device=/dev/nvmap \
  --device=/dev/nvhost-gpu --device=/dev/nvhost-as-gpu \
  -v /usr/lib/aarch64-linux-gnu/tegra:/usr/lib/aarch64-linux-gnu/tegra:ro \
  -v /usr/share/vulkan/icd.d:/usr/share/vulkan/icd.d:ro \
  ghcr.io/mutablelogic/go-whisper run --debug

# AMD/Intel/Raspberry Pi
docker run --rm --name gowhisper -p 8081:8081 \
  --device=/dev/dri:/dev/dri \
  -v /usr/share/vulkan/icd.d:/usr/share/vulkan/icd.d:ro \
  ghcr.io/mutablelogic/go-whisper run --debug
```

## Troubleshooting

If you encounter issues with Vulkan device access, ensure that the necessary device files are correctly passed into the container and that your user has appropriate permissions. You can check Vulkan installation and device availability using `vulkaninfo` in the container. You should see a GPU other than "deviceName = llvmpipe" listed.

```bash
# NVIDIA
docker run --rm -it --gpus all \
  --entrypoint vulkaninfo \
  ghcr.io/mutablelogic/go-whisper --summary

# NVIDIA Jetson/Tegra
docker run --rm -it --runtime nvidia \
  --device=/dev/nvhost-ctrl-gpu \
  --device=/dev/nvhost-prof-gpu --device=/dev/nvmap \
  --device=/dev/nvhost-gpu --device=/dev/nvhost-as-gpu \
  -v /usr/lib/aarch64-linux-gnu/tegra:/usr/lib/aarch64-linux-gnu/tegra:ro \
  -v /usr/share/vulkan/icd.d:/usr/share/vulkan/icd.d:ro \
  --entrypoint vulkaninfo \
  ghcr.io/mutablelogic/go-whisper --summary

# AMD/Intel/Raspberry Pi
docker run --rm -it \
  --device=/dev/dri:/dev/dri \
  -v /usr/share/vulkan/icd.d:/usr/share/vulkan/icd.d:ro \
  --entrypoint vulkaninfo \
  ghcr.io/mutablelogic/go-whisper --summary
  ```
