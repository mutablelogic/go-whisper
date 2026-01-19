# Go-Whisper HTTP API

This document describes the HTTP API implemented by the `gowhisper` server, based on the handlers in [`pkg/httphandler`](https://pkg.go.dev/github.com/mutablelogic/go-whisper/pkg/httphandler) and the client library in [`pkg/httpclient`](https://pkg.go.dev/github.com/mutablelogic/go-whisper/pkg/httpclient).

## Base URL Structure

By default, all API endpoints are prefixed with `/api/whisper` which can be configured via the server's `--prefix` flag.

Example base URL: `http://localhost:8081/api/whisper`

## Models

### List Models

```http
GET /api/whisper/model
```

Returns a list of available models currently loaded or downloaded. This will
include models from commercial providers if configured.

**Response Format:**

```json
{
  "object": "list", 
  "models": [
    {
      "id": "ggml-large-v3",
      "object": "model",
      "path": "ggml-large-v3.bin",
      "created": 1722090121,
      "owned_by": "whisper"
    }
  ]
}
```

### Get Model

```http
GET /api/whisper/model/{id}
```

Retrieves detailed information about a specific model by ID.

**Parameters:**

- `id` (path): Model identifier

**Response:** Model object with details

### Download Model from Hugging Face

```http
POST /api/whisper/model
```

Downloads a new Whisper model from [Hugging Face](https://huggingface.co/ggerganov/whisper.cpp) and stores it locally.

**Request Body:**

```json
{
  "model": "ggml-medium-q5_0"
}
```

**Streaming Support:**
Add `Accept: text/event-stream` header for real-time download progress.

**Streaming Response Events:**

```http
event: progress
data: {"current": 10159256, "total": 539212467, "percent": 1.88}

event: done
data: {"id":"ggml-medium-q5_0","object":"model","path":"ggml-medium-q5_0.bin","created":1722411778}
```

**Standard Response:** Model object upon completion

### Delete Model

```http
DELETE /api/whisper/model/{id}
```

Removes a downloaded model from the system.

**Parameters:**

- `id` (path): Model identifier

**Response:** HTTP 204 No Content on success

## Audio Processing

### Transcription

```http
POST /api/whisper/transcribe
```

Transcribes audio files to text in the original language.

**Request Format:** `multipart/form-data`

**Form Fields:**

- `audio` (file, required): Audio file to transcribe
- `model` (string, required): Model ID to use
- `prompt` (string, optional): Text to guide the model's style
- `temperature` (number, optional): Sampling temperature (0-1)
- `language` (string, optional): Two-letter language code
- `stream` (boolean, optional): Enable streaming response
- `diarize` (boolean, optional): Identify and separate speakers
- `filename` (string, optional): Original filename with extension

**Response Formats:**

The response format is determined by the `Accept` header:

- `application/json` (default): Full transcription object with segments
- `text/plain`: Plain text output with speaker labels if available
- `application/x-subrip` or `text/srt`: SubRip subtitle format
- `text/vtt`: WebVTT subtitle format

**JSON Response Example:**

```json
{
  "object": "transcription",
  "language": "en",
  "segments": [
    {
      "start": 0.0,
      "end": 5.0,
      "text": "Hello, this is a transcription example.",
      "speaker": "SPEAKER_00"
    }
  ]
}
```

### Translation

```http
POST /api/whisper/translate
```

Translates audio files to English text.

**Request Format:** Same as transcription (`multipart/form-data`)

**Form Fields:**

- `audio` (file, required): Audio file to translate
- `model` (string, required): Model ID to use
- `prompt` (string, optional): Text to guide the model's style
- `temperature` (number, optional): Sampling temperature (0-1)
- `stream` (boolean, optional): Enable streaming response
- `diarize` (boolean, optional): Identify and separate speakers
- `filename` (string, optional): Original filename with extension

**Note:** Translation always outputs English text regardless of input language.

**Response Formats:** Same as transcription (JSON, plain text, SRT, VTT)

## Error Handling

The API uses standard HTTP status codes and returns JSON error responses:

```json
{
  "error": "Error message description"
}
```

**Common Status Codes:**

- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `204 No Content`: Successful deletion
- `400 Bad Request`: Invalid request parameters
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Server is loading/not ready

## Content Type Support

### Supported Audio Formats

The server supports various audio formats including:

- WAV, MP3, FLAC, M4A, OGG
- Any format supported by the underlying Whisper implementation

### Response Content Types

- `application/json`: Structured response data
- `text/plain`: Plain text transcriptions
- `application/x-subrip`: SubRip subtitle format (.srt)
- `text/vtt`: WebVTT format for web video
- `text/event-stream`: Server-sent events for streaming

## OpenTelemetry (OTEL) Integration

The Go-Whisper server includes comprehensive OpenTelemetry support for observability and monitoring.

### OTEL Configuration

Configure OTEL via command-line flags or environment variables:

```bash
# Command-line configuration
gowhisper run --otel.endpoint="http://jaeger:14268/api/traces" \
                 --otel.name="go-whisper" \
                 --otel.header="Authorization=Bearer token"

# Alternatively, with environment variables
export OTEL_EXPORTER_OTLP_ENDPOINT="http://jaeger:14268/api/traces"
export OTEL_SERVICE_NAME="go-whisper"
```

### Distributed Tracing

The server automatically creates traces for:

- **HTTP Requests**: All incoming API requests are traced
- **Model Operations**: Model loading, downloading, and deletion
- **Audio Processing**: Transcription and translation operations
- **External API Calls**: OpenAI and ElevenLabs integrations

### Trace Context Propagation

The server supports standard OTEL trace context propagation:

- Extracts trace context from incoming HTTP headers
- Propagates context to downstream services

### Metrics and Spans

**HTTP Request Spans:**

- Operation name: `http.request`
- Attributes include method, URL, status code, user agent

**Model Operation Spans:**

- `whisper.model.load`
- `whisper.model.download`
- `whisper.model.delete`

**Audio Processing Spans:**

- `whisper.transcribe`
- `whisper.translate`

**Custom Attributes:**

- Model ID and parameters
- Audio file metadata
- Processing duration
- Error details
