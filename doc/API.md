# Draft API

(From OpenAPI docs)

## Create transcription - upload

```html
POST /v1/audio/transcriptions
```

Transcribes audio into the input language.

**body - Required**
The audio file object (not file name) to transcribe, in one of these formats: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, or webm.

**model - string - Required**
ID of the model to use.

**language - string - Optional**
The language of the input audio. Supplying the input language in ISO-639-1 format will improve accuracy and latency.

**prompt - string - Optional**
An optional text to guide the model's style or continue a previous audio segment. The prompt should match the audio language.

**response_format - string - Optional**
Defaults to json
The format of the transcript output, in one of these options: json, text, srt, verbose_json, or vtt.

**temperature - number - Optional**
Defaults to 0
The sampling temperature, between 0 and 1. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic. If set to 0, the model will use log probability to automatically increase the temperature until certain thresholds are hit.

**timestamp_granularities[] - array - Optional**
Defaults to segment
The timestamp granularities to populate for this transcription. response_format must be set verbose_json to use timestamp granularities. Either or both of these options are supported: word, or segment. Note: There is no additional latency for segment timestamps, but generating word timestamps incurs additional latency.

**Returns**
The transcription object or a verbose transcription object.

### Example request

```bash
curl https://localhost/v1/audio/transcriptions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: multipart/form-data" \
  -F file="@/path/to/file/audio.mp3" \
  -F model="whisper-1"
```

```json
{
  "text": "Imagine the wildest idea that you've ever had, and you're curious about how it might scale to something that's a 100, a 1,000 times bigger. This is a place where you can get to do that."
}
```
