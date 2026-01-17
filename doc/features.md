
# Features

| Model(s) | Transcription | Translation to English | Diarization | Streaming |
|----------|---------------|-------------|-------------|-----------|
| GGML Whisper `*-en.bin` | ✅ |  |  | ✅ |
| GGML Whisper `*.bin` | ✅ | ✅ |  | ✅ |
| GGML Whisper `ggml-small.en-tdrz.bin`[^1] | ✅ |  |  ✅ | ✅ |
| OpenAI `whisper-1` [^2] | ✅ | ✅ |  | |
| OpenAI `gpt-4o-*-transcribe` [^4],[^5] | ✅ | |  | ✅ |
| ElevenLabs `scribe_v1` [^3] | ✅ |  |  ✅ | |

[^1]: <https://huggingface.co/akashmjn/tinydiarize-whisper.cpp>
[^2]: <https://platform.openai.com/docs/models/whisper-1>
[^3]: <https://elevenlabs.io/docs/models#scribe-v1>
[^4]: <https://platform.openai.com/docs/models/gpt-4o-transcribe>
[^5]: <https://platform.openai.com/docs/models/gpt-4o-mini-transcribe>
