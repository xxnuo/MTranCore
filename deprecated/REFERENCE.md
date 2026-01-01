# Worker Protocol Reference

This document provides a comprehensive reference for all communication protocols supported by the MTranCore Worker service.

## Table of Contents

- [Overview](#overview)
- [Error Codes](#error-codes)
- [gRPC Protocol](#grpc-protocol)
- [HTTP Protocol](#http-protocol)
- [WebSocket Protocol](#websocket-protocol)
- [Configuration](#configuration)

## Overview

The MTranCore Worker service provides translation capabilities through three different protocols:

- **gRPC**: High-performance RPC protocol with streaming support
- **HTTP**: RESTful API using JSON
- **WebSocket**: Real-time bidirectional communication

All protocols share the same business logic and error codes, providing consistent behavior across different communication methods.

## Error Codes

The service uses simplified HTTP-standard error codes:

| Code | Constant | Description |
| ---- | -------- | ----------- |
| 200  | `CodeSuccess` | Operation completed successfully |
| 400  | `CodeInvalidParams` | Invalid request parameters |
| 500  | `CodeTransFailure` | Translation failed |
| 502  | `CodeInternalError` | Internal server error |
| 503  | `CodeNotReady` | Translation engine not ready |

## gRPC Protocol

### Connection

Connect to the gRPC server using the configured address (default: `0.0.0.0:8988`).

```
grpc://<host>:<port>
```

### Service: TranslatorService

The gRPC service is defined in `worker.proto`:

```proto
service TranslatorService {
  rpc Health(HealthRequest) returns (HealthResponse);
  rpc Trans(TransRequest) returns (TransResponse);
  rpc TransStream(stream TransRequest) returns (stream TransResponse);
  rpc Exit(ExitRequest) returns (ExitResponse);
}
```

### Methods

#### Health

Check server health and engine ready status.

**Request**: `HealthRequest` (empty)

**Response**: `HealthResponse`

```proto
message HealthResponse {
  int32 code = 1;      // 200 for healthy server
  string message = 2;  // "OK"
  bool ready = 3;      // Engine ready status
}
```

#### Trans

Translate a single text.

**Request**: `TransRequest`

```proto
message TransRequest {
  string text = 1;  // Text to translate (required)
  bool html = 2;    // Whether to treat input as HTML (default: false)
}
```

**Response**: `TransResponse`

```proto
message TransResponse {
  int32 code = 1;             // 200 on success, error code otherwise
  string message = 2;         // Description of the result
  string translated_text = 3; // Translated text (on success)
}
```

**Error Codes**:

- `400`: Text is required
- `500`: Translation failed
- `503`: Engine not ready or server shutting down

#### TransStream

Translate multiple texts using bidirectional streaming.

**Request Stream**: `TransRequest`

```proto
message TransRequest {
  string text = 1;  // Text to translate
  bool html = 2;    // Whether to treat input as HTML
}
```

**Response Stream**: `TransResponse`

```proto
message TransResponse {
  int32 code = 1;             // 200 on success, error code otherwise
  string message = 2;         // Description of the result
  string translated_text = 3; // Translated text (on success)
}
```

**Behavior**:

- Client sends multiple `TransRequest` messages
- Server responds with corresponding `TransResponse` for each request
- Client closes stream by sending EOF
- Empty text returns empty translation with success code

#### Exit

Shut down the server gracefully or forcefully.

**Request**: `ExitRequest`

```proto
message ExitRequest {
  int32 time = 1;   // Seconds to wait before shutdown (default: 0)
  bool force = 2;   // Force shutdown without waiting for active streams
}
```

**Response**: `ExitResponse`

```proto
message ExitResponse {
  int32 code = 1;      // 200 on success, error code otherwise
  string message = 2;  // Description of shutdown status
}
```

**Error Codes**:

- `400`: Time must be non-negative

**Behavior**:

- If `force=true`: Immediate shutdown
- If `force=false`: Waits up to 30 seconds for active streams, then forces shutdown

## HTTP Protocol

### Connection

Connect to the HTTP server using the configured address (default: `http://0.0.0.0:8988`).

All endpoints use JSON for request and response bodies.

### Standard Response Format

Success responses:

```json
{
  "code": 200,
  "data": {
    /* endpoint-specific data */
  }
}
```

Error responses:

```json
{
  "code": <error_code>,
  "message": "Error description"
}
```

### Endpoints

#### GET /health

Check server health and engine ready status.

**Response**: HTTP 200

```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "ready": true
  }
}
```

#### POST /trans

Translate text.

**Request Body**:

```json
{
  "text": "Text to translate", // Required
  "html": false // Optional, default: false
}
```

**Success Response**: HTTP 200

Content-Type: `text/plain; charset=utf-8`

Returns the translated text directly as plain text (not JSON).

**Error Responses**:

- HTTP 400 (Invalid parameters):
  ```json
  {
    "code": 400,
    "message": "text is required"
  }
  ```
- HTTP 503 (Engine not ready):
  ```json
  {
    "code": 503,
    "message": "Translation engine not ready"
  }
  ```
- HTTP 503 (Server shutting down):
  ```json
  {
    "code": 503,
    "message": "server is shutting down"
  }
  ```
- HTTP 500 (Translation failed):
  ```json
  {
    "code": 500,
    "message": "Translation failed: error details"
  }
  ```

#### POST /exit

Shut down the server gracefully or forcefully.

**Request Body** (optional):

```json
{
  "time": 0, // Seconds to wait (default: 0)
  "force": false // Force shutdown (default: false)
}
```

**Success Response**: HTTP 200

```json
{
  "code": 200,
  "data": {
    "message": "Shutdown initiated"
  }
}
```

**Error Response**: HTTP 400

```json
{
  "code": 400,
  "message": "time must be non-negative"
}
```

## WebSocket Protocol

### Connection

Connect to the WebSocket server at the configured address (default: `ws://0.0.0.0:8988/ws`).

### Message Format

All messages are JSON objects with the following structure:

**Client Message**:

```json
{
  "type": "message_type",
  "data": {
    /* type-specific data */
  }
}
```

**Server Response**:

```json
{
  "type": "message_type", // Same as request type
  "code": 200, // Error code
  "msg": "success", // Message or error description
  "data": {
    /* type-specific data */
  }
}
```

### Message Types

#### health

Check server health and engine ready status.

**Client Message**:

```json
{
  "type": "health",
  "data": {}
}
```

**Server Response**:

```json
{
  "type": "health",
  "code": 200,
  "msg": "OK",
  "data": {
    "ready": true
  }
}
```

#### trans

Translate text.

**Client Message**:

```json
{
  "type": "trans",
  "data": {
    "text": "Text to translate",
    "html": false
  }
}
```

**Server Response**:

```json
{
  "type": "trans",
  "code": 200,
  "msg": "OK",
  "data": {
    "translated_text": "Translated text"
  }
}
```

**Error Codes**: 400, 500, 503

#### exit

Shut down the server.

**Client Message**:

```json
{
  "type": "exit",
  "data": {
    "time": 0, // Optional, default: 0
    "force": false // Optional, default: false
  }
}
```

**Server Response**:

```json
{
  "type": "exit",
  "code": 200,
  "msg": "Shutdown initiated"
}
```

**Note**: The WebSocket connection is closed after exit response.

### Unknown Message Type

If an unknown message type is sent:

**Server Response**:

```json
{
  "type": "unknown_type",
  "code": 400,
  "msg": "Unknown message type: unknown_type"
}
```

## Configuration

The worker service can be configured using environment variables:

### General Settings

| Variable      | Default   | Description                                          |
| ------------- | --------- | ---------------------------------------------------- |
| `LOG_LEVEL`   | `info`    | Logging level (debug, info, warn, error)             |
| `MODEL_DIR`   | `./`      | Model directory for auto-loading on startup          |
| `MODEL_PATH`  | (empty)   | Model file path for auto-loading                     |
| `LEXICAL_SHORTLIST_PATH` | (empty) | Lexical shortlist file path for auto-loading |
| `SERVER_HOST` | `0.0.0.0` | Server host address (shared by all enabled services) |
| `SERVER_PORT` | `8988`    | Server port (shared by all enabled services)         |
| `MAX_LENGTH_BREAK` | `128` | Maximum sentence length before breaking          |

### Service Control

All services (HTTP, WebSocket, gRPC) run on the same port using connection multiplexing.

| Variable           | Default | Description          |
| ------------------ | ------- | -------------------- |
| `ENABLE_HTTP`      | `true`  | Enable HTTP REST API |
| `ENABLE_WEBSOCKET` | `true`  | Enable WebSocket API |
| `ENABLE_GRPC`      | `true`  | Enable gRPC API      |

### gRPC Unix Domain Socket (Advanced)

For better local performance, gRPC can also listen on a Unix domain socket:

| Variable           | Default | Description                                         |
| ------------------ | ------- | --------------------------------------------------- |
| `GRPC_UNIX_SOCKET` | (empty) | Path to Unix socket file (e.g., `/tmp/mtrancore.sock`) |

When enabled, gRPC will listen on both TCP and Unix socket simultaneously. Unix domain sockets provide 20-40% better performance for local IPC by avoiding TCP/IP stack overhead.

### Example Configuration

Engine auto-loading with model directory:

```bash
export MODEL_DIR=./models/enzh
./worker
```

Engine auto-loading with individual file paths:

```bash
export MODEL_PATH=./models/enzh/model.enzh.intgemm.alphas.bin
export LEXICAL_SHORTLIST_PATH=./models/enzh/lex.50.50.enzh.s2t.bin
./worker
```

Disable WebSocket and change server port:

```bash
export ENABLE_WEBSOCKET=false
export SERVER_PORT=9000
./worker
```

Run only gRPC server:

```bash
export ENABLE_HTTP=false
export ENABLE_WEBSOCKET=false
export ENABLE_GRPC=true
./worker
```

Enable gRPC with Unix socket for better local performance:

```bash
export GRPC_UNIX_SOCKET=/tmp/mtrancore.sock
./worker
```

This will start both TCP (on port 8988) and Unix socket listeners for gRPC.

## Common Workflows

### 1. Basic Translation Flow

**Note**: Engine is auto-loaded on startup via `-model-dir` or individual file path parameters. No need to call poweron.

1. **Check Health** (includes ready status)

   - gRPC: Call `Health()`
   - HTTP: `GET /health`
   - WebSocket: Send `{"type": "health"}`

2. **Translate**
   - gRPC: Call `Trans(text="Hello", html=false)`
   - HTTP: `POST /trans` with `{"text": "Hello", "html": false}`
   - WebSocket: Send `{"type": "trans", "data": {"text": "Hello", "html": false}}`

### 2. Batch Translation (gRPC Stream)

1. Open `TransStream` connection
2. Send multiple `TransRequest` messages
3. Receive corresponding `TransResponse` for each request
4. Close stream when done

### 3. Graceful Shutdown

1. **Wait for completion**:

   - Call `Exit(time=5, force=false)` to wait 5 seconds + up to 30 seconds for active tasks

2. **Force shutdown**:
   - Call `Exit(time=0, force=true)` for immediate shutdown

## Notes

- Engine is auto-loaded on startup via configuration parameters
- All three protocols share the same translation engine instance
- Model files must include: `model.*.intgemm.alphas.bin`, `lex.*.s2t.bin`, `vocab.*.spm`
- Paths can be absolute or relative to the configured `MODEL_DIR`
- The service handles concurrent requests through an internal queue for HTTP and WebSocket
- gRPC `TransStream` is optimized for batch processing
- Empty text in trans requests returns empty translation with success code
- All servers can be enabled/disabled independently via environment variables
- Fatal WASM errors are detected automatically and trigger process exit (exit code 1)
