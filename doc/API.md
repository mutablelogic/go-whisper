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
Example streaming response:

```text
event: ping

event: progress
data: {"status":"downloading ggml-medium-q5_0.bin","total":539212467,"completed":10159256}

event: progress
data: {"status":"downloading ggml-medium-q5_0.bin","total":539212467,"completed":21895036}

event: progress
data: {"status":"downloading ggml-medium-q5_0.bin","total":539212467,"completed":33540592}

event: ok
data: {"id":"ggml-medium-q5_0","object":"model","path":"ggml-medium-q5_0.bin","created":1722411778}
```

### Delete Model

```html
DELETE /v1/models/{model-id}
```

Deletes a model by it's ID. If the model is deleted, a 200 OK status is returned.

## Transcription and translation with file upload

### Transcription

Transcribes audio into the input language. This endpoint's purpose is to transcribe media files into text, in the language of the media file, based on the model's supported languages.

```html
POST /v1/audio/transcriptions
```

The request should be a multipart/form-data request with the [following fields](../pkg/client/gowhisper/schema.go):

```json
{
  "model": "<model-id>",
  "file": "<binary data>",
  "prompt": "<optional-prompt>",
  "response_format": "<optional-response-format>",
  "temperature": "<optional-temperature>",
  "stream": "<optional-stream-boolean>",
  "language": "<optional-language>"
}
```

The response depends on the `response_format` and `stream` parameters:

* `response_format` can be one of `json`, `text`, `srt`, `verbose_json`, or `vtt`.
* `stream`

TODO

### Translation

This is the same as transcription (above) except that the `language` parameter is always set to 'en', to translate the audio into English.
The request should be a multipart/form-data request with the [following fields](../pkg/client/gowhisper/schema.go):

```html
POST /v1/audio/translations
```

TODO
