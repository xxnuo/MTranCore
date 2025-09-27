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

The service uses standardized error codes across all protocols:

| Code | Constant                     | Description                                |
| ---- | ---------------------------- | ------------------------------------------ |
| 0    | `CodeSuccess`                | Operation completed successfully           |
| 1000 | `CodePoweronInvalidParams`   | Invalid parameters for poweron request     |
| 1001 | `CodePoweronPathNotExists`   | Model path does not exist                  |
| 1002 | `CodePoweronIncompleteFiles` | Model files are incomplete or missing      |
| 1003 | `CodePoweronInternalError`   | Internal error during poweron              |
| 1009 | `CodePoweronUnknownError`    | Unknown poweron error                      |
| 1100 | `CodePoweroffInvalidParams`  | Invalid parameters for poweroff request    |
| 1101 | `CodePoweroffWaitingTask`    | Server is shutting down, waiting for tasks |
| 1109 | `CodePoweroffInternalError`  | Internal error during poweroff             |
| 1200 | `CodeComputeInvalidParams`   | Invalid parameters for compute request     |
| 1201 | `CodeComputeFailure`         | Translation computation failed             |
| 1209 | `CodeComputeInternalError`   | Internal error during compute              |

## gRPC Protocol

### Connection

Connect to the gRPC server using the configured address (default: `0.0.0.0:8991`).

```
grpc://<host>:<port>
```

### Service: TranslatorService

The gRPC service is defined in `translator.proto`:

```proto
service TranslatorService {
  rpc Health(HealthRequest) returns (HealthResponse);
  rpc Poweron(PoweronRequest) returns (PoweronResponse);
  rpc Poweroff(PoweroffRequest) returns (PoweroffResponse);
  rpc Ready(ReadyRequest) returns (ReadyResponse);
  rpc Compute(ComputeRequest) returns (ComputeResponse);
  rpc ComputeStream(stream ComputeRequest) returns (stream ComputeResponse);
}
```

### Methods

#### Health

Check server health status.

**Request**: `HealthRequest` (empty)

**Response**: `HealthResponse`

```proto
message HealthResponse {
  int32 code = 1;      // Always 0 for healthy server
  string message = 2;  // "OK"
}
```

#### Poweron

Load the translation engine with model files.

**Request**: `PoweronRequest`

```proto
message PoweronRequest {
  string path = 1;  // Path to model directory (absolute or relative to work_dir)
}
```

**Response**: `PoweronResponse`

```proto
message PoweronResponse {
  int32 code = 1;      // 0 on success, error code otherwise
  string message = 2;  // Description of the result
}
```

**Error Codes**:

- `1000`: Path is required
- `1001`: Path does not exist
- `1002`: Model files are incomplete
- `1003`: Internal error loading engine

#### Poweroff

Shut down the server gracefully or forcefully.

**Request**: `PoweroffRequest`

```proto
message PoweroffRequest {
  int32 time = 1;   // Seconds to wait before shutdown (default: 0)
  bool force = 2;   // Force shutdown without waiting for active streams
}
```

**Response**: `PoweroffResponse`

```proto
message PoweroffResponse {
  int32 code = 1;      // 0 on success, error code otherwise
  string message = 2;  // Description of shutdown status
}
```

**Error Codes**:

- `1100`: Time must be non-negative
- `1101`: Waiting for active streams to complete (non-force shutdown)

**Behavior**:

- If `force=true`: Immediate shutdown
- If `force=false`: Waits up to 30 seconds for active streams, then forces shutdown

#### Ready

Check if the translation engine is loaded and ready.

**Request**: `ReadyRequest` (empty)

**Response**: `ReadyResponse`

```proto
message ReadyResponse {
  int32 code = 1;      // Always 0
  string message = 2;  // "OK"
  bool ready = 3;      // true if engine is loaded
}
```

#### Compute

Translate a single text.

**Request**: `ComputeRequest`

```proto
message ComputeRequest {
  string text = 1;  // Text to translate (required)
  bool html = 2;    // Whether to treat input as HTML (default: false)
}
```

**Response**: `ComputeResponse`

```proto
message ComputeResponse {
  int32 code = 1;             // 0 on success, error code otherwise
  string message = 2;         // Description of the result
  string translated_text = 3; // Translated text (on success)
}
```

**Error Codes**:

- `1200`: Text is required or engine not ready
- `1201`: Translation failed
- `1209`: Server is shutting down

#### ComputeStream

Translate multiple texts using bidirectional streaming.

**Request Stream**: `ComputeRequest`

```proto
message ComputeRequest {
  string text = 1;  // Text to translate
  bool html = 2;    // Whether to treat input as HTML
}
```

**Response Stream**: `ComputeResponse`

```proto
message ComputeResponse {
  int32 code = 1;             // 0 on success, error code otherwise
  string message = 2;         // Description of the result
  string translated_text = 3; // Translated text (on success)
}
```

**Behavior**:

- Client sends multiple `ComputeRequest` messages
- Server responds with corresponding `ComputeResponse` for each request
- Client closes stream by sending EOF
- Empty text returns empty translation with success code

## HTTP Protocol

### Connection

Connect to the HTTP server using the configured address (default: `http://0.0.0.0:8989`).

All endpoints use JSON for request and response bodies.

### Standard Response Format

Success responses:

```json
{
  "code": 0,
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

Check server health status.

**Response**: HTTP 200

```json
(empty body with status OK)
```

#### POST /poweron

Load the translation engine with model files.

**Request Body**:

```json
{
  "path": "path/to/model" // Absolute or relative to work_dir
}
```

**Success Response**: HTTP 200

```json
{
  "code": 0,
  "data": {
    "message": "Engine loaded successfully"
  }
}
```

**Error Responses**:

- HTTP 400 (Invalid parameters):
  ```json
  {
    "code": 1000,
    "message": "path is required"
  }
  ```
- HTTP 404 (Path not found):
  ```json
  {
    "code": 1001,
    "message": "path does not exist: /full/path"
  }
  ```
- HTTP 400 (Incomplete files):
  ```json
  {
    "code": 1002,
    "message": "Missing required model files"
  }
  ```
- HTTP 500 (Internal error):
  ```json
  {
    "code": 1003,
    "message": "Internal error message"
  }
  ```

#### POST /poweroff

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
  "code": 0,
  "data": {
    "message": "Server is shutting down"
  }
}
```

**Error Response**: HTTP 200

```json
{
  "code": 1101,
  "message": "Server is shutting down, waiting for requests to complete"
}
```

#### GET /ready

Check if the translation engine is loaded and ready.

**Success Response**: HTTP 200

```json
{
  "code": 0,
  "data": {
    "ready": true // or false
  }
}
```

#### POST /compute

Translate text.

**Request Body**:

```json
{
  "text": "Text to translate", // Required
  "html": false // Optional, default: false
}
```

**Success Response**: HTTP 200

```json
{
  "code": 0,
  "data": {
    "translated_text": "Translated text"
  }
}
```

**Error Responses**:

- HTTP 400 (Invalid parameters):
  ```json
  {
    "code": 1200,
    "message": "text is required"
  }
  ```
- HTTP 503 (Engine not ready):
  ```json
  {
    "code": 1200,
    "message": "Engine is not ready. Please call poweron first"
  }
  ```
- HTTP 503 (Server shutting down):
  ```json
  {
    "code": 1209,
    "message": "Server is shutting down"
  }
  ```
- HTTP 500 (Translation failed):
  ```json
  {
    "code": 1201,
    "message": "Translation failed: error details"
  }
  ```

## WebSocket Protocol

### Connection

Connect to the WebSocket server at the configured address (default: `ws://0.0.0.0:8990/ws`).

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
  "code": 0, // Error code
  "msg": "success", // Message or error description
  "data": {
    /* type-specific data */
  }
}
```

### Message Types

#### poweron

Load the translation engine with model files.

**Client Message**:

```json
{
  "type": "poweron",
  "data": {
    "path": "path/to/model"
  }
}
```

**Server Response**:

```json
{
  "type": "poweron",
  "code": 0,
  "msg": "success",
  "data": {
    "message": "Engine loaded successfully"
  }
}
```

**Error Codes**: Same as HTTP/gRPC poweron errors (1000-1009)

#### poweroff

Shut down the server.

**Client Message**:

```json
{
  "type": "poweroff",
  "data": {
    "time": 0, // Optional, default: 0
    "force": false // Optional, default: false
  }
}
```

**Server Response**:

```json
{
  "type": "poweroff",
  "code": 0,
  "msg": "success",
  "data": {
    "message": "Server is shutting down"
  }
}
```

**Note**: The WebSocket connection is closed after poweroff response.

#### ready

Check if the translation engine is loaded.

**Client Message**:

```json
{
  "type": "ready",
  "data": {}
}
```

**Server Response**:

```json
{
  "type": "ready",
  "code": 0,
  "msg": "success",
  "data": {
    "ready": true
  }
}
```

#### compute

Translate text.

**Client Message**:

```json
{
  "type": "compute",
  "data": {
    "text": "Text to translate",
    "html": false
  }
}
```

**Server Response**:

```json
{
  "type": "compute",
  "code": 0,
  "msg": "success",
  "data": {
    "translated_text": "Translated text"
  }
}
```

**Error Codes**: Same as HTTP/gRPC compute errors (1200-1209)

### Unknown Message Type

If an unknown message type is sent:

**Server Response**:

```json
{
  "type": "unknown_type",
  "code": 1009,
  "msg": "Unknown message type: unknown_type"
}
```

## Configuration

The worker service can be configured using environment variables:

### General Settings

| Variable    | Default | Description                                |
| ----------- | ------- | ------------------------------------------ |
| `LOG_LEVEL` | `info`  | Logging level (debug, info, warn, error)   |
| `WORK_DIR`  | `./`    | Working directory for relative model paths |

### HTTP Server

| Variable      | Default   | Description        |
| ------------- | --------- | ------------------ |
| `ENABLE_HTTP` | `true`    | Enable HTTP server |
| `HTTP_HOST`   | `0.0.0.0` | HTTP server host   |
| `HTTP_PORT`   | `8989`    | HTTP server port   |

### WebSocket Server

| Variable           | Default   | Description             |
| ------------------ | --------- | ----------------------- |
| `ENABLE_WEBSOCKET` | `true`    | Enable WebSocket server |
| `WEBSOCKET_HOST`   | `0.0.0.0` | WebSocket server host   |
| `WEBSOCKET_PORT`   | `8990`    | WebSocket server port   |

### gRPC Server

| Variable      | Default   | Description        |
| ------------- | --------- | ------------------ |
| `ENABLE_GRPC` | `true`    | Enable gRPC server |
| `GRPC_HOST`   | `0.0.0.0` | gRPC server host   |
| `GRPC_PORT`   | `8991`    | gRPC server port   |

### Example Configuration

Disable WebSocket and change HTTP port:

```bash
export ENABLE_WEBSOCKET=false
export HTTP_PORT=9000
./worker
```

Run only gRPC server:

```bash
export ENABLE_HTTP=false
export ENABLE_WEBSOCKET=false
export ENABLE_GRPC=true
./worker
```

## Common Workflows

### 1. Basic Translation Flow

1. **Check Health** (optional)

   - gRPC: Call `Health()`
   - HTTP: `GET /health`
   - WebSocket: Not typically needed

2. **Load Engine**

   - gRPC: Call `Poweron(path="/path/to/model")`
   - HTTP: `POST /poweron` with `{"path": "/path/to/model"}`
   - WebSocket: Send `{"type": "poweron", "data": {"path": "/path/to/model"}}`

3. **Check Ready** (optional)

   - gRPC: Call `Ready()`
   - HTTP: `GET /ready`
   - WebSocket: Send `{"type": "ready"}`

4. **Translate**
   - gRPC: Call `Compute(text="Hello", html=false)`
   - HTTP: `POST /compute` with `{"text": "Hello", "html": false}`
   - WebSocket: Send `{"type": "compute", "data": {"text": "Hello", "html": false}}`

### 2. Batch Translation (gRPC Stream)

1. Open `ComputeStream` connection
2. Send multiple `ComputeRequest` messages
3. Receive corresponding `ComputeResponse` for each request
4. Close stream when done

### 3. Graceful Shutdown

1. **Wait for completion**:

   - Call `Poweroff(time=5, force=false)` to wait 5 seconds + up to 30 seconds for active tasks

2. **Force shutdown**:
   - Call `Poweroff(time=0, force=true)` for immediate shutdown

## Notes

- All three protocols share the same translation engine instance
- Model files must include: `model.*.intgemm.alphas.bin`, `lex.*.s2t.bin`, `vocab.*.spm`
- Paths can be absolute or relative to the configured `WORK_DIR`
- The service handles concurrent requests through an internal queue for HTTP and WebSocket
- gRPC `ComputeStream` is optimized for batch processing
- Empty text in compute requests returns empty translation with success code
- All servers can be enabled/disabled independently via environment variables
