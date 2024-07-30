# Notes on building

## Package Config

libwhisper.pc

```pkg-config
prefix=/Users/djt/Projects/go-whisper/

Name: libwhisper
Description: Whisper is a C/C++ library for speech transcription, translation and diarization.
Version: 0.0.0
Cflags: -I${prefix}/third_party/whisper.cpp/include -I${prefix}/third_party/whisper.cpp/ggml/include
Libs: -L${prefix}/third_party/whisper.cpp -lwhisper -lggml -lm -lstdc++
```

libwhisper-darwin.pc

```pkg-config
prefix=/Users/djt/Projects/go-whisper/

Name: libwhisper-darwin
Description: Whisper is a C/C++ library for speech transcription, translation and diarization.
Version: 0.0.0
Libs: -framework Accelerate -framework Metal -framework Foundation -framework CoreGraphics
```

I don't know what the windows one should be as I don't have a windows machine.

## FFmpeg

Required for decoding media files into audio which is suitable for audio detection and transcription.

### MacOS

On Macintosh with homebrew, for example:

```bash
brew install ffmpeg@6 chromaprint make
brew link ffmpeg@6
```

### Debian

If you're using Debian you may not be able to get the ffmpeg 6 unless you first of all add the debi-multimedia repository. You can do this by adding the following line to your /etc/apt/sources.list file:

Add the repository as privileged user:

```bash
echo "deb https://www.deb-multimedia.org $(lsb_release -sc) main" >> /etc/apt/sources.list
apt update -y -oAcquire::AllowInsecureRepositories=true
apt install -y --force-yes deb-multimedia-keyring
apt install -y libavcodec-dev libavdevice-dev libavfilter-dev libavutil-dev libswscale-dev libswresample-dev
```

### Ubuntu 22.04

Easier with Ubuntu! Installing FFmpeg 6.1 libraries:

```bash
add-apt-repository -y ppa:ubuntuhandbook1/ffmpeg6
apt-get update
apt-get install -y libavcodec-dev libavdevice-dev libavfilter-dev libavutil-dev libswscale-dev libswresample-dev
```
