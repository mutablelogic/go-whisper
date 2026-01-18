
# Features

* *Transcription* is the process of converting spoken language into written text, in any language supported by the model.
* *Translation* is the process of converting spoken language into written text in English, regardless of the original language.
* *Diarization* is the process of identifying and separating different speakers in an audio recording.
* *Realtime* processing allows for transcription or translation of audio streams to be returned as it is being processed, rather than waiting for the entire audio file to be processed before returning results.

| Model(s) | Transcription | Translation to English | Diarization | Realtime |
|----------|---------------|-------------|-------------|-----------|
| GGML Whisper `*-en.bin` | ✅ |  |  | ✅ |
| GGML Whisper `*.bin` | ✅ | ✅ |  | ✅ |
| GGML Whisper `ggml-small.en-tdrz.bin`[^1] | ✅ |  |  ✅ | ✅ |
| OpenAI `whisper-1` [^2] | ✅ | ✅ |  | |
| OpenAI `gpt-4o-*-transcribe` [^4],[^5] | ✅ | |  | ✅ |
| ElevenLabs `scribe_v1`,`scribe_v2` [^3] | ✅ |  |  ✅ | |

[^1]: <https://huggingface.co/akashmjn/tinydiarize-whisper.cpp>
[^2]: <https://platform.openai.com/docs/models/whisper-1>
[^3]: <https://elevenlabs.io/docs/models#scribe-v1>
[^4]: <https://platform.openai.com/docs/models/gpt-4o-transcribe>
[^5]: <https://platform.openai.com/docs/models/gpt-4o-mini-transcribe>
