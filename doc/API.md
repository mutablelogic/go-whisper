# Whisper server API

Based on OpenAPI docs

## Ping

```html
GET /v1/ping
```

Returns a OK status to indicate the API is up and running.

## Models

### List Models

```html
GET /v1/models
```

Returns a list of available models. Example response:

```json
{
  "object": "list",
  "models": [
    {
      "id": "ggml-large-v3",
      "object": "model",
      "path": "ggml-large-v3.bin",
      "created": 1722090121
    },
    {
      "id": "ggml-medium-q5_0",
      "object": "model",
      "path": "ggml-medium-q5_0.bin",
      "created": 1722081999
    }
  ]
}
```

### Download Model

```html
POST /v1/models
POST /v1/models?stream={bool}
```

The request should be a application/json, multipart/form-data or application/x-www-form-urlencoded request with the following fields:

```json
{
  "path": "ggml-large-v3.bin"
}
```

Downloads a model from remote huggingface repository. If the optional `stream` argument is true,
the progress is streamed back to the client as a series of [text/event-stream](https://html.spec.whatwg.org/multipage/server-sent-events.html) events.

If the model is already downloaded, a 200 OK status is returned. If the model was downloaded, a 201 Created status is returned.

### Delete Model

```html
DELETE /v1/models/{model-id}
```

Deletes a model by it's ID. If the model is deleted, a 200 OK status is returned.

## Transcription and translation with file upload

### Transcription

This endpoint's purpose is to transcribe media files into text, in the language of the media file.

```html
POST /v1/audio/transcriptions
POST /v1/audio/transcriptions?stream={bool}
```

The request should be a multipart/form-data request with the following fields:

```json
{
  "model": "<model-id>",
  "file": "<binary data>",
  "language": "<language-code>",
  "response_format": "<response-format>",
}
```

Transcribes audio into the input language.

`file` (required) The audio file object (not file name) to transcribe. This can be audio or video, and the format is auto-detected. The "best" audio stream is selected from the file, and the audio is converted to 16 kHz mono PCM format during transcription.

`model-id` (required) ID of the model to use. This should have previously been downloaded.

`language` (optional) The language of the input audio in ISO-639-1 format. If not set, then the language is auto-detected.

`response_format` (optional, defaults to `json`). The format of the transcript output, in one of these options: json, text, srt, verbose_json, or vtt.

If the optional `stream` argument is true, the segments of the transcription are returned as a series of [text/event-stream](https://html.spec.whatwg.org/multipage/server-sent-events.html) events. Otherwise, the full transcription is returned in the response body.

### Translation

This is the same as transcription (above) except that the `language` parameter is not optional, and should be the language to translate the audio into.

```html
POST /v1/audio/translations
POST /v1/audio/translations?stream={bool}
```
